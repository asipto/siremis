package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// tokenFromResponseCookie extracts the session token from the cookie that
// GMAuthRefresh set on the response.
func tokenFromResponseCookie(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	res := w.Result()
	defer res.Body.Close()
	for _, c := range res.Cookies() {
		if c.Name == "session_token" && c.Value != "" {
			return c.Value
		}
	}
	t.Fatalf("no session_token cookie was set")
	return ""
}

// sessionExpiry reads a token's expiry from GMSessions under the lock.
func sessionExpiry(t *testing.T, token string) time.Time {
	t.Helper()
	GMSessionsMu.RLock()
	s, exists := GMSessions[token]
	GMSessionsMu.RUnlock()
	if !exists {
		t.Fatalf("token %q is no longer in GMSessions", token)
	}
	return s.expiry
}

// TestSessionRefreshKeepsSameToken checks the in-place sliding refresh: the
// token is not replaced, the response cookie carries the same value, and the
// expiry is pushed forward even when more than 60 seconds remained.
func TestSessionRefreshKeepsSameToken(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-sliding"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(290 * time.Second), // almost 5 min left
	}

	r := httptest.NewRequest("GET", "/siremis/view/main", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	if rc := GMAuthRefresh(w, r); rc != 0 {
		t.Fatalf("refresh must succeed, got code %d", rc)
	}

	if got := tokenFromResponseCookie(t, w); got != token {
		t.Fatalf("token must not rotate, got %q instead of %q", got, token)
	}

	// the request is left untouched: its cookie still carries the token
	if c, err := r.Cookie("session_token"); err != nil || c.Value != token {
		t.Fatalf("the request cookie must stay intact, got %v, %v", c, err)
	}

	GMSessionsMu.RLock()
	s, exists := GMSessions[token]
	count := len(GMSessions)
	GMSessionsMu.RUnlock()

	if !exists {
		t.Fatalf("the token must still be present in GMSessions")
	}
	if count != 1 {
		t.Fatalf("exactly one session must exist, got %d", count)
	}
	if s.username != "admin" {
		t.Fatalf("the username must be preserved, got %q", s.username)
	}
	if s.expiry.Before(time.Now().Add(250 * time.Second)) {
		t.Fatalf("expiry must be pushed forward even with > 60 s left, got %v", s.expiry)
	}
}

// TestSessionRefreshSlidesExpiry checks that the expiry is pushed forward on
// every refresh, including when little time remains: the remaining lifetime
// comes back towards 300 seconds on each call.
func TestSessionRefreshSlidesExpiry(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-slide"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(10 * time.Second),
	}

	r := httptest.NewRequest("GET", "/siremis/view/list", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	before := sessionExpiry(t, token)
	GMAuthRefresh(w, r)
	after := sessionExpiry(t, token)

	if !after.After(before) {
		t.Fatalf("expiry must be pushed forward, before %v after %v", before, after)
	}
	if after.Before(time.Now().Add(250 * time.Second)) {
		t.Fatalf("expiry must come back towards 300 s, got %v", after)
	}
}

// TestSessionRefreshExpiredFails checks that an expired session is not
// resurrected.
func TestSessionRefreshExpiredFails(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-expired"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(-1 * time.Minute),
	}

	r := httptest.NewRequest("GET", "/siremis/view/main", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	if rc := GMAuthRefresh(w, r); rc >= 0 {
		t.Fatalf("an expired session must be refused, got code %d", rc)
	}
	GMSessionsMu.RLock()
	_, exists := GMSessions[token]
	GMSessionsMu.RUnlock()
	if exists {
		t.Fatalf("the expired session must be deleted")
	}
}

// TestSessionRefreshThenAuthActive checks that after GMAuthRefresh, the
// session check of the same request (GMSessionAuthActive) still recognises the
// session. With the in-place sliding expiry, the request cookie's token stays
// valid: this is the case of the menu, charts and jsonrpc pages, whose handler
// checks the session after the router-level refresh.
func TestSessionRefreshThenAuthActive(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-active"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(290 * time.Second),
	}

	r := httptest.NewRequest("GET", "/siremis/menu/main", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	GMAuthRefresh(w, r)

	if !GMSessionAuthActive(w, r) {
		t.Fatalf("the session must stay active after the in-place refresh")
	}
}

// TestSessionRefreshConcurrentReads reproduces the rotation race: two
// overlapping requests carrying the same cookie. With the in-place refresh,
// the "losing" request stays valid because the token is never deleted.
func TestSessionRefreshConcurrentReads(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-concurrent"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(290 * time.Second),
	}

	// both requests carry the same token, like a double click
	r1 := httptest.NewRequest("GET", "/siremis/view/main", nil)
	r1.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	r2 := httptest.NewRequest("GET", "/siremis/menu/main", nil)
	r2.AddCookie(&http.Cookie{Name: "session_token", Value: token})

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	if rc := GMAuthRefresh(w1, r1); rc != 0 {
		t.Fatalf("first request refused, code %d", rc)
	}
	if rc := GMAuthRefresh(w2, r2); rc != 0 {
		t.Fatalf("second request refused, code %d", rc)
	}
	if !GMSessionAuthActive(w1, r1) {
		t.Fatalf("the first request must stay active after the second refresh")
	}
	if !GMSessionAuthActive(w2, r2) {
		t.Fatalf("the second request must stay active after the second refresh")
	}
	if got := tokenFromResponseCookie(t, w2); got != token {
		t.Fatalf("token must not rotate, got %q instead of %q", got, token)
	}
}
