package app

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type SessionStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
	ttl    time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		tokens: make(map[string]time.Time),
		ttl:    ttl,
	}
}

func (s *SessionStore) start(w http.ResponseWriter) error {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)
	expires := time.Now().Add(s.ttl)

	s.mu.Lock()
	s.tokens[token] = expires
	s.mu.Unlock()

	setSessionCookie(w, token, expires, s.ttl)

	return nil
}

func (s *SessionStore) authenticated(w http.ResponseWriter, r *http.Request) bool {
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

	newExpiry := time.Now().Add(s.ttl)
	s.tokens[cookie.Value] = newExpiry

	if w != nil {
		setSessionCookie(w, cookie.Value, newExpiry, s.ttl)
	}
	return true
}

func (s *SessionStore) clear(w http.ResponseWriter, r *http.Request) {
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

func setSessionCookie(w http.ResponseWriter, token string, expires time.Time, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
		Expires:  expires,
	})
}
