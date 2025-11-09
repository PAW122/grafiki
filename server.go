package main

import (
	"html/template"
	"log"
	"net/http"
)

type imageInfo struct {
	Name string
	URL  string
}

type pageData struct {
	Dir      string
	Images   []imageInfo
	LoggedIn bool
}

type server struct {
	dir      string
	cfg      appConfig
	tmpl     *template.Template
	sessions *sessionStore
	logger   *requestLogger
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.logger != nil {
		s.logger.Log(r, "wyswietl")
	}

	images, err := listImages(s.dir)
	if err != nil {
		log.Printf("listImages: %v", err)
		http.Error(w, "failed to load images", http.StatusInternalServerError)
		return
	}

	data := pageData{
		Dir:      s.dir,
		Images:   images,
		LoggedIn: s.sessions.authenticated(r),
	}

	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template execute: %v", err)
	}
}
