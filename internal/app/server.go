package app

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

type ServerOptions struct {
	Dir      string
	Config   Config
	Template *template.Template
	Sessions *SessionStore
	Logger   *RequestLogger
	DB       *sql.DB
	Favicon  string
}

type imageInfo struct {
	Name string
	URL  string
}

type pageData struct {
	Images                    []imageInfo
	LoggedIn                  bool
	Folders                   []folderView
	ActiveFolder              *folderView
	SharedMode                bool
	AllowFolderManagement     bool
	BaseURL                   string
	View                      string
	SubmissionGroups          []submissionGroupView
	ActiveSubmissionGroup     *submissionGroupView
	SubmissionEntries         []submissionEntryView
	SubmissionSharedMode      bool
	AllowSubmissionManagement bool
	AllowSubmissionUpload     bool
	SubmissionShareLink       string
	SubmissionUploadLimit     int
}

type Server struct {
	dir            string
	submissionsDir string
	cfg            Config
	tmpl           *template.Template
	sessions       *SessionStore
	logger         *RequestLogger
	db             *sql.DB
	favicon        string
}

func NewServer(opts ServerOptions) (*Server, error) {
	submissionsDir := filepath.Join(opts.Dir, "_submitted")
	if err := EnsureDir(submissionsDir); err != nil {
		return nil, err
	}
	return &Server{
		dir:            opts.Dir,
		submissionsDir: submissionsDir,
		cfg:            opts.Config,
		tmpl:           opts.Template,
		sessions:       opts.Sessions,
		logger:         opts.Logger,
		db:             opts.DB,
		favicon:        opts.Favicon,
	}, nil
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/", s)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(s.dir))))
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/api/upload", s.handleUpload)
	mux.HandleFunc("/api/delete", s.handleDelete)
	mux.HandleFunc("/api/images/rename", s.handleRenameImage)
	mux.HandleFunc("/api/folders", s.handleFolders)
	mux.HandleFunc("/api/folders/", s.handleFolderByID)
	mux.HandleFunc("/api/submissions/upload", s.handleSubmissionUpload)
	mux.HandleFunc("/api/submissions/groups", s.handleSubmissionGroups)
	mux.HandleFunc("/api/submissions/groups/", s.handleSubmissionGroupByID)
	mux.HandleFunc("/shared/", s.handleSharedFolder)
	mux.HandleFunc("/submitted/", s.handleSubmittedRoutes)
	mux.HandleFunc("/submitted/file/", s.handleSubmissionFile)
	mux.HandleFunc("/favicon.ico", s.handleFavicon)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathSlug := strings.Trim(r.URL.Path, "/")
	viewParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("view")))
	if pathSlug == "" && viewParam == "submitted" {
		if !s.sessions.authenticated(r) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		s.renderSubmittedDashboard(w, r)
		return
	}

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
		View:                  "gallery",
		SubmissionUploadLimit: int(submissionUploadMaxSize >> 20),
	}

	s.renderPage(w, data)
}

func (s *Server) renderSubmittedDashboard(w http.ResponseWriter, r *http.Request) {
	loggedIn := s.sessions.authenticated(r)
	if !loggedIn {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	viewerToken := s.ensureSubmissionViewerToken(w, r)
	baseURL := requestBaseURL(r)

	groups, err := s.listSubmissionGroups(loggedIn)
	if err != nil {
		log.Printf("list submission groups: %v", err)
		http.Error(w, "failed to load groups", http.StatusInternalServerError)
		return
	}

	groupSlug := strings.TrimSpace(r.URL.Query().Get("group"))
	var activeRecord *submissionGroupRecord
	groupViews := make([]submissionGroupView, 0, len(groups))
	for i := range groups {
		rec := groups[i]
		if activeRecord == nil {
			if groupSlug == "" && i == 0 {
				activeRecord = &groups[i]
			} else if groupSlug != "" && rec.Slug == sanitizeFilename(groupSlug) {
				activeRecord = &groups[i]
			}
		}
		groupViews = append(groupViews, rec.toView(baseURL))
	}

	var activeView *submissionGroupView
	var entries []submissionEntryView
	shareLink := ""
	allowUpload := false

	if activeRecord != nil {
		if activeRecord.Visibility == visibilityShared {
			if _, err := s.ensureSubmissionSharedToken(activeRecord.ID); err == nil {
				if refreshed, err := s.getSubmissionGroupByID(activeRecord.ID); err == nil {
					activeRecord = refreshed
				}
			}
		}
		view := activeRecord.toView(baseURL)
		activeView = &view
		shareLink = view.ShareURL
		allowUpload = loggedIn || activeRecord.Visibility != visibilityPrivate
		entries, err = s.submissionEntriesForGroup(activeRecord, viewerToken, loggedIn)
		if err != nil {
			log.Printf("list submissions: %v", err)
			http.Error(w, "failed to load submissions", http.StatusInternalServerError)
			return
		}
	}

	data := pageData{
		LoggedIn:                  loggedIn,
		View:                      "submitted",
		BaseURL:                   baseURL,
		SubmissionGroups:          groupViews,
		ActiveSubmissionGroup:     activeView,
		SubmissionEntries:         entries,
		SubmissionSharedMode:      false,
		AllowSubmissionManagement: loggedIn,
		AllowSubmissionUpload:     allowUpload,
		SubmissionShareLink:       shareLink,
		SubmissionUploadLimit:     int(submissionUploadMaxSize >> 20),
		AllowFolderManagement:     loggedIn,
	}

	s.renderPage(w, data)
}

func (s *Server) renderPage(w http.ResponseWriter, data pageData) {
	if data.View == "" {
		data.View = "gallery"
	}
	if data.SubmissionUploadLimit == 0 {
		data.SubmissionUploadLimit = int(submissionUploadMaxSize >> 20)
	}
	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template execute: %v", err)
	}
}

func (s *Server) canAccessFolder(rec *folderRecord, loggedIn bool) bool {
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

func (s *Server) imagesForFolder(rec *folderRecord) ([]imageInfo, error) {
	dir := s.dir
	urlPrefix := "/images/"
	if rec != nil && rec.Path != "" {
		dir = filepath.Join(s.dir, rec.Path)
		urlPrefix = "/images/" + folderURLPrefix(rec.Path) + "/"
	}
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}
	return listImages(dir, urlPrefix)
}

func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
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
