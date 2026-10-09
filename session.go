package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GMSession struct {
	username string
	expiry   time.Time
}

func GMSessionCookie(r *http.Request, token string, expiresAt time.Time) *http.Cookie {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if r != nil && r.TLS != nil {
		cookie.Secure = true
	}
	return cookie
}

func (s GMSession) IsExpired() bool {
	return s.expiry.Before(time.Now())
}

var GMSessions = map[string]GMSession{}
var GMSessionsMu sync.RWMutex

func GMCheckPasswords(cfgPassword string, valPassword string) bool {
	if strings.HasPrefix(cfgPassword, "sha256:") {
		hash := sha256.Sum256([]byte(valPassword))
		return (strings.TrimPrefix(cfgPassword, "sha256:") == hex.EncodeToString(hash[:]))
	}
	if strings.HasPrefix(cfgPassword, "sha1:") {
		hash := sha1.Sum([]byte(valPassword))
		return (strings.TrimPrefix(cfgPassword, "sha1:") == hex.EncodeToString(hash[:]))
	}
	if strings.HasPrefix(cfgPassword, "text:") {
		return (valPassword == strings.TrimPrefix(cfgPassword, "text:"))
	}
	if strings.HasPrefix(cfgPassword, "md5:") {
		hash := md5.Sum([]byte(valPassword))
		return (strings.TrimPrefix(cfgPassword, "md5:") == hex.EncodeToString(hash[:]))
	}

	return (valPassword == cfgPassword)
}

func GMSessionAuthCheck(w http.ResponseWriter, r *http.Request) int {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// w.WriteHeader(http.StatusUnauthorized)
			return -1
		}
		// w.WriteHeader(http.StatusBadRequest)
		return -2
	}
	sessionToken := c.Value

	GMSessionsMu.RLock()
	userSession, exists := GMSessions[sessionToken]
	GMSessionsMu.RUnlock()
	if !exists {
		// w.WriteHeader(http.StatusUnauthorized)
		return -3
	}
	if userSession.IsExpired() {
		GMSessionsMu.Lock()
		delete(GMSessions, sessionToken)
		GMSessionsMu.Unlock()
		// w.WriteHeader(http.StatusUnauthorized)
		return -4
	}
	return 0
}

func GMSessionAuthActive(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return false
		}
		return false
	}
	sessionToken := c.Value

	GMSessionsMu.RLock()
	userSession, exists := GMSessions[sessionToken]
	GMSessionsMu.RUnlock()
	if !exists {
		return false
	}
	if userSession.IsExpired() {
		GMSessionsMu.Lock()
		delete(GMSessions, sessionToken)
		GMSessionsMu.Unlock()
		return false
	}
	return true
}

func GMAuthRefresh(w http.ResponseWriter, r *http.Request) int {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// w.WriteHeader(http.StatusUnauthorized)
			return -1
		}
		// w.WriteHeader(http.StatusBadRequest)
		return -2
	}
	sessionToken := c.Value

	GMSessionsMu.RLock()
	userSession, exists := GMSessions[sessionToken]
	GMSessionsMu.RUnlock()
	if !exists {
		// w.WriteHeader(http.StatusUnauthorized)
		return -3
	}
	if userSession.IsExpired() {
		log.Printf("session expired - token: %s\n", sessionToken)
		GMSessionsMu.Lock()
		delete(GMSessions, sessionToken)
		GMSessionsMu.Unlock()
		// w.WriteHeader(http.StatusUnauthorized)
		return -4
	}

	// rafraîchissement glissant : toute requête authentifiée prolonge la
	// session. Le jeton est conservé tel quel (pas de rotation) et seule
	// l'échéance est repoussée sous le verrou d'écriture ; le cookie de la
	// requête reste donc valide pour les vérifications de session du même
	// traitement, et deux requêtes concurrentes ne se révoquent plus l'une
	// l'autre.
	log.Printf("session active - token: %s\n", sessionToken)

	expiresAt := time.Now().Add(300 * time.Second)

	GMSessionsMu.Lock()
	GMSessions[sessionToken] = GMSession{
		username: userSession.username,
		expiry:   expiresAt,
	}
	GMSessionsMu.Unlock()

	http.SetCookie(w, GMSessionCookie(r, sessionToken, expiresAt))

	return 0
}

func GMLoginCheck(w http.ResponseWriter, r *http.Request) int {
	username := r.FormValue("username")
	password := r.FormValue("password")
	authok := false

	for _, v := range GMConfigV.AuthUsers {
		if v.Username == username {
			if GMCheckPasswords(v.Password, password) {
				authok = true
				break
			}
		}
	}

	if !authok {
		return -1
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(300 * time.Second)

	GMSessionsMu.Lock()
	GMSessions[sessionToken] = GMSession{
		username: username,
		expiry:   expiresAt,
	}
	GMSessionsMu.Unlock()

	http.SetCookie(w, GMSessionCookie(r, sessionToken, expiresAt))

	return 0
}
