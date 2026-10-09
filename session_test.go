package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// tokenFromResponseCookie extrait le jeton du cookie de session émis par
// GMAuthRefresh dans la réponse.
func tokenFromResponseCookie(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	res := w.Result()
	defer res.Body.Close()
	for _, c := range res.Cookies() {
		if c.Name == "session_token" && c.Value != "" {
			return c.Value
		}
	}
	t.Fatalf("aucun cookie session_token émis")
	return ""
}

// sessionExpiry lit l'échéance du jeton dans GMSessions sous le verrou.
func sessionExpiry(t *testing.T, token string) time.Time {
	t.Helper()
	GMSessionsMu.RLock()
	s, exists := GMSessions[token]
	GMSessionsMu.RUnlock()
	if !exists {
		t.Fatalf("le jeton %q n'est plus dans GMSessions", token)
	}
	return s.expiry
}

// TestSessionRefreshKeepsSameToken vérifie le rafraîchissement glissant sur
// place : le jeton n'est pas remplacé, le cookie de la réponse porte la même
// valeur et l'expiration est repoussée même s'il restait plus de 60 secondes.
func TestSessionRefreshKeepsSameToken(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-sliding"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(290 * time.Second), // presque 5 min restantes
	}

	r := httptest.NewRequest("GET", "/siremis/view/main", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	if rc := GMAuthRefresh(w, r); rc != 0 {
		t.Fatalf("le rafraîchissement doit réussir, code obtenu %d", rc)
	}

	if got := tokenFromResponseCookie(t, w); got != token {
		t.Fatalf("le jeton ne doit pas tourner, obtenu %q au lieu de %q", got, token)
	}

	// la requête n'est pas modifiée : son cookie porte toujours le jeton
	if c, err := r.Cookie("session_token"); err != nil || c.Value != token {
		t.Fatalf("le cookie de la requête doit rester intact, obtenu %v, %v", c, err)
	}

	GMSessionsMu.RLock()
	s, exists := GMSessions[token]
	count := len(GMSessions)
	GMSessionsMu.RUnlock()

	if !exists {
		t.Fatalf("le jeton doit rester présent dans GMSessions")
	}
	if count != 1 {
		t.Fatalf("une seule session doit exister, obtenu %d", count)
	}
	if s.username != "admin" {
		t.Fatalf("l'utilisateur doit être conservé, obtenu %q", s.username)
	}
	if s.expiry.Before(time.Now().Add(250 * time.Second)) {
		t.Fatalf("l'expiration doit être repoussée même si > 60 s restantes, obtenu %v", s.expiry)
	}
}

// TestSessionRefreshSlidesExpiry vérifie que l'échéance est bien repoussée à
// chaque rafraîchissement, y compris quand il reste peu de temps : la durée
// restante revient vers les 300 secondes à chaque appel.
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
		t.Fatalf("l'échéance doit être repoussée, avant %v après %v", before, after)
	}
	if after.Before(time.Now().Add(250 * time.Second)) {
		t.Fatalf("l'échéance doit revenir vers 300 s, obtenu %v", after)
	}
}

// TestSessionRefreshExpiredFails vérifie qu'une session expirée n'est pas
// ressuscitée.
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
		t.Fatalf("une session expirée doit être refusée, code obtenu %d", rc)
	}
	GMSessionsMu.RLock()
	_, exists := GMSessions[token]
	GMSessionsMu.RUnlock()
	if exists {
		t.Fatalf("la session expirée doit être supprimée")
	}
}

// TestSessionRefreshThenAuthActive vérifie qu'après GMAuthRefresh, la
// vérification de session du même traitement (GMSessionAuthActive) reconnaît
// toujours la session. Avec l'expiration glissante sur place, le jeton du
// cookie de la requête reste valide : c'est le cas des pages menu, charts et
// jsonrpc, dont le handler vérifie la session après le rafraîchissement fait
// au niveau du routeur.
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
		t.Fatalf("la session doit rester active après le rafraîchissement sur place")
	}
}

// TestSessionRefreshConcurrentReads reproduit la course de la rotation : deux
// requêtes chevauchantes partant du même cookie. Avec le rafraîchissement sur
// place, la requête « perdante » reste valide puisque le jeton n'est jamais
// supprimé.
func TestSessionRefreshConcurrentReads(t *testing.T) {
	GMSessions = map[string]GMSession{}
	GMSessionsMu = sync.RWMutex{}

	token := "test-token-concurrent"
	GMSessions[token] = GMSession{
		username: "admin",
		expiry:   time.Now().Add(290 * time.Second),
	}

	// les deux requêtes portent le même jeton, comme un double-clic
	r1 := httptest.NewRequest("GET", "/siremis/view/main", nil)
	r1.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	r2 := httptest.NewRequest("GET", "/siremis/menu/main", nil)
	r2.AddCookie(&http.Cookie{Name: "session_token", Value: token})

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	if rc := GMAuthRefresh(w1, r1); rc != 0 {
		t.Fatalf("première requête refusée, code %d", rc)
	}
	if rc := GMAuthRefresh(w2, r2); rc != 0 {
		t.Fatalf("seconde requête refusée, code %d", rc)
	}
	if !GMSessionAuthActive(w1, r1) {
		t.Fatalf("la première requête doit rester active après le second rafraîchissement")
	}
	if !GMSessionAuthActive(w2, r2) {
		t.Fatalf("la seconde requête doit rester active après le second rafraîchissement")
	}
	if got := tokenFromResponseCookie(t, w2); got != token {
		t.Fatalf("le jeton ne doit pas tourner, obtenu %q au lieu de %q", got, token)
	}
}
