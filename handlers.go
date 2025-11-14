package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/skip2/go-qrcode"
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

	folderSlug := strings.TrimSpace(r.FormValue("folder"))
	if folderSlug == "" {
		writeJSONError(w, http.StatusBadRequest, "Wybierz folder docelowy")
		return
	}
	folder, err := s.getFolderBySlug(folderSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusBadRequest, "Folder nie istnieje")
			return
		}
		log.Printf("folder lookup: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie sprawdzic folderu")
		return
	}

	targetDir := s.dir
	if folder.Path != "" {
		targetDir = filepath.Join(targetDir, folder.Path)
	}
	if err := ensureDir(targetDir); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie przygotowac folderu")
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

	target, err := uniqueFilename(targetDir, filename)
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
		"folder": folderSlug,
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
		Name   string `json:"name"`
		Folder string `json:"folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Folder) == "" {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe dane")
		return
	}

	folder, err := s.getFolderBySlug(strings.TrimSpace(req.Folder))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusBadRequest, "Folder nie istnieje")
			return
		}
		log.Printf("folder lookup: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie sprawdzic folderu")
		return
	}

	filename := filepath.Base(req.Name)
	if !isImageFile(filename) {
		writeJSONError(w, http.StatusBadRequest, "Nieobslugiwany plik")
		return
	}

	targetDir := s.dir
	if folder.Path != "" {
		targetDir = filepath.Join(targetDir, folder.Path)
	}
	target := filepath.Join(targetDir, filename)
	cleanDir := filepath.Clean(targetDir)
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

func (s *server) handleFolders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListFoldersAPI(w, r)
	case http.MethodPost:
		s.handleCreateFolderAPI(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
	}
}

func (s *server) handleFolderByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/folders/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(path, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowy folder")
		return
	}

	if len(parts) == 2 && parts[1] == "qr" {
		s.handleFolderQR(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetFolderAPI(w, r, id)
	case http.MethodPatch:
		s.handleUpdateFolderAPI(w, r, id)
	case http.MethodDelete:
		s.handleDeleteFolderAPI(w, r, id)
	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
	}
}

func (s *server) handleCreateFolderAPI(w http.ResponseWriter, r *http.Request) {
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe dane")
		return
	}
	folder, err := s.createFolder(req.Name)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	view := folder.toView(requestBaseURL(r))
	writeJSON(w, http.StatusCreated, view)
}

func (s *server) handleListFoldersAPI(w http.ResponseWriter, r *http.Request) {
	loggedIn := s.sessions.authenticated(r)
	folders, err := s.listFolders(loggedIn)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac folderow")
		return
	}

	baseURL := requestBaseURL(r)
	var views []folderView
	for _, folder := range folders {
		views = append(views, folder.toView(baseURL))
	}
	writeJSON(w, http.StatusOK, map[string]any{"folders": views})
}

func (s *server) handleGetFolderAPI(w http.ResponseWriter, r *http.Request, id int64) {
	folder, err := s.getFolderByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "Folder nie istnieje")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac folderu")
		return
	}
	loggedIn := s.sessions.authenticated(r)
	if !s.canAccessFolder(folder, loggedIn) {
		writeJSONError(w, http.StatusForbidden, "Brak dostepu")
		return
	}
	writeJSON(w, http.StatusOK, folder.toView(requestBaseURL(r)))
}

func (s *server) handleUpdateFolderAPI(w http.ResponseWriter, r *http.Request, id int64) {
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}

	var req struct {
		Visibility     string `json:"visibility"`
		RegenerateLink bool   `json:"regenerateLink"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe dane")
		return
	}

	var folder *folderRecord
	var err error

	if req.Visibility != "" {
		folder, err = s.updateFolderVisibility(id, req.Visibility)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		folder, err = s.getFolderByID(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, http.StatusNotFound, "Folder nie istnieje")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac folderu")
			return
		}
	}

	if req.RegenerateLink {
		if folder.Visibility != visibilityShared {
			writeJSONError(w, http.StatusBadRequest, "Folder nie jest ustawiony jako udostepniony")
			return
		}
		folder, err = s.regenerateSharedToken(id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie odswiezyc linku")
			return
		}
	}

	writeJSON(w, http.StatusOK, folder.toView(requestBaseURL(r)))
}

func (s *server) handleDeleteFolderAPI(w http.ResponseWriter, r *http.Request, id int64) {
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}

	if err := s.deleteFolder(id); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			writeJSONError(w, http.StatusNotFound, "Folder nie istnieje")
		case errors.Is(err, errFolderProtected):
			writeJSONError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, errFolderPathInvalid):
			writeJSONError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("delete folder: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie usunac folderu")
		}
		return
	}

	if s.logger != nil {
		s.logger.Log(r, "usunfold")
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleFolderQR(w http.ResponseWriter, r *http.Request, id int64) {
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}
	folder, err := s.getFolderByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "Folder nie istnieje")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac folderu")
		return
	}
	if folder.Visibility != visibilityShared {
		writeJSONError(w, http.StatusBadRequest, "Folder nie ma udostepnionego linku")
		return
	}
	token := ""
	if folder.SharedToken.Valid {
		token = folder.SharedToken.String
	}
	if token == "" {
		token, err = s.ensureSharedToken(folder.ID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie przygotowac linku")
			return
		}
		folder, _ = s.getFolderByID(folder.ID)
	}

	link := requestBaseURL(r) + "/shared/" + token
	png, err := qrcode.Encode(link, qrcode.Medium, 256)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie wygenerowac QR")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\"folder-"+folder.Slug+"-qr.png\"")
	if _, err := w.Write(png); err != nil {
		log.Printf("write qr: %v", err)
	}
}

func (s *server) handleSharedFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/shared/")
	token = strings.Trim(token, "/")
	if token == "" {
		http.NotFound(w, r)
		return
	}

	folder, err := s.getFolderByToken(token)
	if err != nil || folder.Visibility != visibilityShared {
		http.NotFound(w, r)
		return
	}

	if err := s.incrementSharedViews(folder.ID); err != nil {
		log.Printf("shared view: %v", err)
	} else {
		folder.SharedViews++
	}

	baseURL := requestBaseURL(r)
	view := folder.toView(baseURL)

	images, err := s.imagesForFolder(folder)
	if err != nil {
		log.Printf("listImages: %v", err)
		http.Error(w, "failed to load images", http.StatusInternalServerError)
		return
	}

	data := pageData{
		Images:                images,
		LoggedIn:              s.sessions.authenticated(r),
		Folders:               nil,
		ActiveFolder:          &view,
		SharedMode:            true,
		BaseURL:               baseURL,
		AllowFolderManagement: false,
	}

	if s.logger != nil {
		s.logger.Log(r, "folderlink")
	}

	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template execute: %v", err)
	}
}
