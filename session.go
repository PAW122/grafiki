package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type sessionStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
	ttl    time.Duration
}

func newSessionStore(ttl time.Duration) *sessionStore {
	return &sessionStore{
		tokens: make(map[string]time.Time),
		ttl:    ttl,
	}
}

func (s *sessionStore) start(w http.ResponseWriter) error {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)
	expires := time.Now().Add(s.ttl)

	s.mu.Lock()
	s.tokens[token] = expires
	s.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.ttl.Seconds()),
		Expires:  expires,
	})

	return nil
}

func (s *sessionStore) authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expiry, ok := s.tokens[cookie.Value]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(s.tokens, cookie.Value)
		return false
	}
	return true
}

func (s *sessionStore) clear(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		s.mu.Lock()
		delete(s.tokens, cookie.Value)
		s.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
