package main

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type imageInfo struct {
	Name string
	URL  string
}

type pageData struct {
	Images                []imageInfo
	LoggedIn              bool
	Folders               []folderView
	ActiveFolder          *folderView
	SharedMode            bool
	AllowFolderManagement bool
	BaseURL               string
}

type server struct {
	dir      string
	cfg      appConfig
	tmpl     *template.Template
	sessions *sessionStore
	logger   *requestLogger
	db       *sql.DB
	favicon  string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathSlug := strings.Trim(r.URL.Path, "/")
	var rawSlug string
	if pathSlug != "" {
		if strings.Contains(pathSlug, "/") {
			http.NotFound(w, r)
			return
		}
		rawSlug = pathSlug
	} else {
		rawSlug = strings.TrimSpace(r.URL.Query().Get("folder"))
	}

	var folderSlug string
	if rawSlug != "" {
		folderSlug = sanitizeFilename(rawSlug)
		if folderSlug == "" {
			http.NotFound(w, r)
			return
		}
	}

	if s.logger != nil {
		s.logger.Log(r, "wyswietl")
	}

	loggedIn := s.sessions.authenticated(r)
	baseURL := requestBaseURL(r)

	folders, err := s.listFolders(loggedIn)
	if err != nil {
		log.Printf("list folders: %v", err)
		http.Error(w, "failed to load folders", http.StatusInternalServerError)
		return
	}

	var folderViews []folderView
	for _, f := range folders {
		folderViews = append(folderViews, f.toView(baseURL))
	}

	var activeFolder *folderView
	var images []imageInfo

	if folderSlug != "" {
		rec, err := s.getFolderBySlug(folderSlug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			log.Printf("get folder: %v", err)
			http.Error(w, "folder unavailable", http.StatusInternalServerError)
			return
		}
		if !s.canAccessFolder(rec, loggedIn) {
			http.NotFound(w, r)
			return
		}

		view := rec.toView(baseURL)
		activeFolder = &view

		images, err = s.imagesForFolder(rec)
		if err != nil {
			log.Printf("listImages: %v", err)
			http.Error(w, "failed to load images", http.StatusInternalServerError)
			return
		}
	}

	data := pageData{
		Images:                images,
		LoggedIn:              loggedIn,
		Folders:               folderViews,
		ActiveFolder:          activeFolder,
		BaseURL:               baseURL,
		AllowFolderManagement: loggedIn,
	}

	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template execute: %v", err)
	}
}

func (s *server) canAccessFolder(rec *folderRecord, loggedIn bool) bool {
	switch rec.Visibility {
	case visibilityPublic:
		return true
	case visibilityShared:
		return loggedIn
	case visibilityPrivate:
		return loggedIn
	default:
		return false
	}
}

func (s *server) imagesForFolder(rec *folderRecord) ([]imageInfo, error) {
	dir := s.dir
	urlPrefix := "/images/"
	if rec != nil && rec.Path != "" {
		dir = filepath.Join(s.dir, rec.Path)
		urlPrefix = "/images/" + folderURLPrefix(rec.Path) + "/"
	}
	if err := ensureDir(dir); err != nil {
		return nil, err
	}
	return listImages(dir, urlPrefix)
}

func (s *server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if strings.TrimSpace(s.favicon) == "" {
		http.NotFound(w, r)
		return
	}
	if _, err := os.Stat(s.favicon); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, s.favicon)
}
