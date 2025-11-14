package app

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
)

func (s *Server) handleSubmittedRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/submitted/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(path, "shared/") {
		token := strings.Trim(strings.TrimPrefix(path, "shared/"), "/")
		if token == "" {
			http.NotFound(w, r)
			return
		}
		s.renderSubmissionShared(w, r, token)
		return
	}

	slug := sanitizeFilename(path)
	if slug == "" {
		http.NotFound(w, r)
		return
	}
	s.renderSubmissionPublic(w, r, slug)
}

func (s *Server) renderSubmissionPublic(w http.ResponseWriter, r *http.Request, slug string) {
	group, err := s.getSubmissionGroupBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	loggedIn := s.sessions.authenticated(r)
	if group.Visibility == visibilityPrivate && !loggedIn {
		http.NotFound(w, r)
		return
	}

	viewerToken := s.ensureSubmissionViewerToken(w, r)
	baseURL := requestBaseURL(r)

	entries, err := s.submissionEntriesForGroup(group, viewerToken, loggedIn)
	if err != nil {
		log.Printf("list submissions: %v", err)
		http.Error(w, "failed to load submissions", http.StatusInternalServerError)
		return
	}

	view := group.toView(baseURL)
	data := pageData{
		LoggedIn:                  loggedIn,
		View:                      "submitted",
		BaseURL:                   baseURL,
		ActiveSubmissionGroup:     &view,
		SubmissionEntries:         entries,
		SubmissionSharedMode:      false,
		AllowSubmissionManagement: loggedIn,
		AllowSubmissionUpload:     loggedIn || group.Visibility == visibilityPublic,
		SubmissionShareLink:       view.ShareURL,
		SubmissionUploadLimit:     int(submissionUploadMaxSize >> 20),
	}

	s.renderPage(w, data)
}

func (s *Server) renderSubmissionShared(w http.ResponseWriter, r *http.Request, token string) {
	group, err := s.getSubmissionGroupByToken(token)
	if err != nil || group.Visibility != visibilityShared {
		http.NotFound(w, r)
		return
	}

	loggedIn := s.sessions.authenticated(r)
	viewerToken := s.ensureSubmissionViewerToken(w, r)
	baseURL := requestBaseURL(r)

	if err := s.incrementSubmissionSharedViews(group.ID); err != nil {
		log.Printf("submission shared view: %v", err)
	}

	entries, err := s.submissionEntriesForGroup(group, viewerToken, loggedIn)
	if err != nil {
		log.Printf("list submissions: %v", err)
		http.Error(w, "failed to load submissions", http.StatusInternalServerError)
		return
	}

	view := group.toView(baseURL)
	data := pageData{
		View:                      "submitted",
		BaseURL:                   baseURL,
		ActiveSubmissionGroup:     &view,
		SubmissionEntries:         entries,
		SubmissionSharedMode:      true,
		AllowSubmissionManagement: loggedIn,
		AllowSubmissionUpload:     true,
		SubmissionShareLink:       view.ShareURL,
		SubmissionUploadLimit:     int(submissionUploadMaxSize >> 20),
	}

	s.renderPage(w, data)
}

func (s *Server) handleSubmissionUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
		return
	}

	viewerToken := s.ensureSubmissionViewerToken(w, r)
	loggedIn := s.sessions.authenticated(r)

	r.Body = http.MaxBytesReader(w, r.Body, submissionUploadMaxSize)
	if err := r.ParseMultipartForm(submissionUploadMaxSize); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nie udalo sie odczytac pliku")
		return
	}

	groupSlug := sanitizeFilename(r.FormValue("group"))
	uploader := strings.TrimSpace(r.FormValue("name"))
	token := strings.TrimSpace(r.FormValue("token"))

	if groupSlug == "" || uploader == "" {
		writeJSONError(w, http.StatusBadRequest, "Podaj nazwe grupy i swoje imie")
		return
	}

	group, err := s.getSubmissionGroupBySlug(groupSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusBadRequest, "Grupa nie istnieje")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac grupy")
		return
	}

	if group.Visibility == visibilityPrivate && !loggedIn {
		writeJSONError(w, http.StatusForbidden, "Ta grupa jest prywatna")
		return
	}
	if group.Visibility == visibilityShared && !loggedIn {
		if !group.SharedToken.Valid || token == "" || token != group.SharedToken.String {
			writeJSONError(w, http.StatusForbidden, "Ten link nie jest aktywny")
			return
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Brak pliku w formularzu")
		return
	}
	defer file.Close()

	filename := sanitizeFilename(header.Filename)
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

	if !isSubmissionFile(filename) {
		writeJSONError(w, http.StatusBadRequest, "Dozwolone sa tylko obrazy lub PDF")
		return
	}

	if err := s.ensureSubmissionDir(group); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie przygotowac katalogu")
		return
	}

	targetDir := s.submissionGroupDir(group)
	target, err := uniqueFilename(targetDir, filename)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie zapisac pliku")
		return
	}

	dst, err := os.Create(target)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie zapisac pliku")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie zapisac pliku")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	result, err := s.db.Exec(`INSERT INTO submissions (group_id, uploader_name, contributor_token, filename, original_name, mime_type, size_bytes) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		group.ID, uploader, viewerToken, filepath.Base(target), header.Filename, mimeType, written)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie zapisac zgłoszenia")
		return
	}
	id, _ := result.LastInsertId()

	if s.logger != nil {
		s.logger.Log(r, "przeslane")
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status": "ok",
		"id":     id,
	})
}

func (s *Server) handleSubmissionGroups(w http.ResponseWriter, r *http.Request) {
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}

	switch r.Method {
	case http.MethodGet:
		groups, err := s.listSubmissionGroups(true)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac grup")
			return
		}
		var views []submissionGroupView
		baseURL := requestBaseURL(r)
		for _, g := range groups {
			views = append(views, g.toView(baseURL))
		}
		writeJSON(w, http.StatusOK, views)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe dane")
			return
		}
		group, err := s.createSubmissionGroup(req.Name)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, group.toView(requestBaseURL(r)))
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
	}
}

func (s *Server) handleSubmissionGroupByID(w http.ResponseWriter, r *http.Request) {
	if !s.sessions.authenticated(r) {
		writeJSONError(w, http.StatusUnauthorized, "Wymagane logowanie")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/submissions/groups/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe ID grupy")
		return
	}

	switch r.Method {
	case http.MethodGet:
		group, err := s.getSubmissionGroupByID(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, http.StatusNotFound, "Grupa nie istnieje")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac grupy")
			return
		}
		writeJSON(w, http.StatusOK, group.toView(requestBaseURL(r)))
	case http.MethodPatch:
		var req struct {
			Name           string `json:"name"`
			Visibility     string `json:"visibility"`
			RegenerateLink bool   `json:"regenerateLink"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Nieprawidlowe dane")
			return
		}

		var group *submissionGroupRecord
		if strings.TrimSpace(req.Name) != "" {
			group, err = s.renameSubmissionGroup(id, req.Name)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if req.Visibility != "" {
			group, err = s.updateSubmissionGroupVisibility(id, req.Visibility)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if group == nil {
			group, err = s.getSubmissionGroupByID(id)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie pobrac grupy")
				return
			}
		}

		if req.RegenerateLink {
			group, err = s.regenerateSubmissionSharedToken(id)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie odswiezyc linku")
				return
			}
		}

		writeJSON(w, http.StatusOK, group.toView(requestBaseURL(r)))
	case http.MethodDelete:
		if err := s.deleteSubmissionGroup(id); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Nie udalo sie usunac grupy")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		writeJSONError(w, http.StatusMethodNotAllowed, "Metoda niedozwolona")
	}
}

func (s *Server) handleSubmissionFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/submitted/file/")
	idStr = strings.Trim(idStr, "/")
	if idStr == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	entry, group, err := s.getSubmissionEntry(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	loggedIn := s.sessions.authenticated(r)
	viewerToken := submissionViewerTokenFromRequest(r)

	switch group.Visibility {
	case visibilityPrivate:
		if !loggedIn {
			http.NotFound(w, r)
			return
		}
	case visibilityShared:
		if !loggedIn && viewerToken != entry.ContributorToken {
			http.NotFound(w, r)
			return
		}
	}

	dir := s.submissionGroupDir(group)
	target := filepath.Join(dir, entry.FileName)
	cleanDir := filepath.Clean(dir)
	cleanTarget := filepath.Clean(target)
	if cleanTarget != cleanDir && !strings.HasPrefix(cleanTarget, cleanDir+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return
	}

	if _, err := os.Stat(target); err != nil {
		http.NotFound(w, r)
		return
	}

	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+entry.OriginalName+"\"")
	} else {
		w.Header().Set("Content-Disposition", "inline; filename=\""+entry.OriginalName+"\"")
	}
	if entry.MimeType.Valid && entry.MimeType.String != "" {
		w.Header().Set("Content-Type", entry.MimeType.String)
	}
	http.ServeFile(w, r, target)
}
