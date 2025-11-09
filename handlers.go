package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe dane logowania")
		return
	}

	if creds.Username != s.cfg.Username || creds.Password != s.cfg.Password {
		writeJSONError(w, http.StatusUnauthorized, "Bledny login lub haslo")
		return
	}

	if err := s.sessions.start(w); err != nil {
		log.Printf("start session: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie utworzyc sesji")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
		return
	}

	s.sessions.clear(w, r)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
		return
	}
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, uploadMaxSize)
	if err := r.ParseMultipartForm(uploadMaxSize); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nie udalo sie odczytac pliku")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nie znaleziono pliku w formularzu")
		return
	}
	defer file.Close()

	override := strings.TrimSpace(r.FormValue("name"))
	filename := sanitizeFilename(header.Filename)
	if override != "" {
		filename = sanitizeFilename(override)
	}
	if filename == "" {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowa nazwa pliku")
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(filename))
	}
	if ext == "" {
		writeJSONError(w, http.StatusBadRequest, "Plik musi miec rozszerzenie")
		return
	}
	if filepath.Ext(filename) == "" {
		filename += ext
	}

	if !isImageFile(filename) {
		writeJSONError(w, http.StatusBadRequest, "Nieobslugiwany typ pliku")
		return
	}

	target, err := uniqueFilename(s.dir, filename)
	if err != nil {
		log.Printf("uniqueFilename: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Blad podczas zapisu pliku")
		return
	}

	dst, err := os.Create(target)
	if err != nil {
		log.Printf("create file: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie zapisac pliku")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("copy file: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie zapisac pliku")
		return
	}

	if s.logger != nil {
		s.logger.Log(r, "dodajzdj")
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"name":   filepath.Base(target),
	})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
		return
	}
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowy plik do usuniecia")
		return
	}

	filename := filepath.Base(req.Name)
	if !isImageFile(filename) {
		writeJSONError(w, http.StatusBadRequest, "Nieobslugiwany plik")
		return
	}

	target := filepath.Join(s.dir, filename)
	cleanDir := filepath.Clean(s.dir)
	cleanTarget := filepath.Clean(target)
	if cleanTarget != cleanDir && !strings.HasPrefix(cleanTarget, cleanDir+string(os.PathSeparator)) {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowy plik")
		return
	}

	if err := os.Remove(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeJSONError(w, http.StatusNotFound, "Plik nie istnieje")
			return
		}
		log.Printf("remove file: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie usunac pliku")
		return
	}

	if s.logger != nil {
		s.logger.Log(r, "usunzdj")
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
