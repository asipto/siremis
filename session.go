package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GMSession struct {
	username string
	expiry   time.Time
}

func (s GMSession) IsExpired() bool {
	return s.expiry.Before(time.Now())
}

var GMSessions = map[string]GMSession{}

func GMCheckPasswords(cfgPassword string, valPassword string) bool {
	if strings.HasPrefix(cfgPassword, "sha256:") {
		v := strings.TrimPrefix(cfgPassword, "sha256:")
		hash := sha256.Sum256([]byte(v))
		return (valPassword == hex.EncodeToString(hash[:]))
	}
	if strings.HasPrefix(cfgPassword, "sha1:") {
		v := strings.TrimPrefix(cfgPassword, "sha1:")
		hash := sha1.Sum([]byte(v))
		return (valPassword == hex.EncodeToString(hash[:]))
	}
	if strings.HasPrefix(cfgPassword, "text:") {
		return (valPassword == strings.TrimPrefix(cfgPassword, "text:"))
	}
	if strings.HasPrefix(cfgPassword, "md5:") {
		v := strings.TrimPrefix(cfgPassword, "md5:")
		hash := md5.Sum([]byte(v))
		return (valPassword == hex.EncodeToString(hash[:]))
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

	userSession, exists := GMSessions[sessionToken]
	if !exists {
		// w.WriteHeader(http.StatusUnauthorized)
		return -3
	}
	if userSession.IsExpired() {
		delete(GMSessions, sessionToken)
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

	userSession, exists := GMSessions[sessionToken]
	if !exists {
		return false
	}
	if userSession.IsExpired() {
		delete(GMSessions, sessionToken)
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

	userSession, exists := GMSessions[sessionToken]
	if !exists {
		// w.WriteHeader(http.StatusUnauthorized)
		return -3
	}
	if userSession.IsExpired() {
		log.Printf("session expired - token: %s\n", sessionToken)
		delete(GMSessions, sessionToken)
		// w.WriteHeader(http.StatusUnauthorized)
		return -4
	}

	if userSession.expiry.After(time.Now().Add(-60 * time.Second)) {
		// not yet close to expire
		log.Printf("session active - token: %s\n", sessionToken)
		return 1
	}

	newSessionToken := uuid.NewString()
	log.Printf("new session token: %s\n", newSessionToken)

	expiresAt := time.Now().Add(300 * time.Second)

	GMSessions[newSessionToken] = GMSession{
		username: userSession.username,
		expiry:   expiresAt,
	}

	delete(GMSessions, sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSessionToken,
		Expires: expiresAt,
		Path:    "/",
	})

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

	GMSessions[sessionToken] = GMSession{
		username: username,
		expiry:   expiresAt,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
		Path:    "/",
	})

	return 0
}
