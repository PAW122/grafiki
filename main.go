package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultGalleryDir       = "galery"
	defaultConfigPath       = "config.json"
	sessionCookieName       = "gallery_session"
	uploadMaxSize     int64 = 32 << 20 // 32 MB
)

type appConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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

type sessionStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
	ttl    time.Duration
}

type requestLogger struct {
	mu   sync.Mutex
	file *os.File
}

func newRequestLogger(path string) (*requestLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &requestLogger{file: f}, nil
}

func (l *requestLogger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *requestLogger) Log(r *http.Request, action string) {
	if l == nil || l.file == nil {
		return
	}
	now := time.Now()
	entry := fmt.Sprintf("[%s] [%s] [%s] [%s]\n",
		now.Format("2006-01-02"),
		now.Format("15:04:05"),
		clientIP(r),
		action,
	)

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.file.WriteString(entry); err != nil {
		log.Printf("write log: %v", err)
	}
}

var (
	dirFlag    = flag.String("dir", defaultGalleryDir, "Directory containing images")
	addrFlag   = flag.String("addr", ":3051", "Server listen address")
	configFlag = flag.String("config", defaultConfigPath, "Path to configuration file")
)

func main() {
	flag.Parse()

	dir, err := filepath.Abs(*dirFlag)
	if err != nil {
		log.Fatalf("resolve dir: %v", err)
	}

	if err := ensureDir(dir); err != nil {
		log.Fatalf("ensure gallery dir %q: %v", dir, err)
	}

	configPath, err := filepath.Abs(*configFlag)
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}

	cfg, created, err := loadOrCreateConfig(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if created {
		log.Printf("Created default config at %s (edit to change admin credentials)", configPath)
	}

	logsPath := filepath.Join(filepath.Dir(configPath), "logs")
	reqLogger, err := newRequestLogger(logsPath)
	if err != nil {
		log.Fatalf("open logs file: %v", err)
	}
	defer reqLogger.Close()

	tmpl := template.Must(template.New("gallery").Parse(pageTemplate))
	srv := &server{
		dir:      dir,
		cfg:      cfg,
		tmpl:     tmpl,
		sessions: newSessionStore(24 * time.Hour),
		logger:   reqLogger,
	}

	mux := http.NewServeMux()
	mux.Handle("/", srv)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(dir))))
	mux.HandleFunc("/api/login", srv.handleLogin)
	mux.HandleFunc("/api/logout", srv.handleLogout)
	mux.HandleFunc("/api/upload", srv.handleUpload)
	mux.HandleFunc("/api/delete", srv.handleDelete)

	log.Printf("Serving gallery from %s at http://%s (config: %s, logs: %s)", dir, addressForLog(*addrFlag), configPath, logsPath)
	if err := http.ListenAndServe(*addrFlag, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
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

func listImages(dir string) ([]imageInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var images []imageInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isImageFile(entry.Name()) {
			continue
		}
		escaped := url.PathEscape(entry.Name())
		images = append(images, imageInfo{
			Name: entry.Name(),
			URL:  "/images/" + escaped,
		})
	}

	sort.Slice(images, func(i, j int) bool {
		return strings.ToLower(images[i].Name) < strings.ToLower(images[j].Name)
	})

	return images, nil
}

func isImageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".avif":
		return true
	default:
		return false
	}
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func loadOrCreateConfig(path string) (appConfig, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := appConfig{
				Username: "admin",
				Password: "admin123",
			}
			if err := writeConfig(path, cfg); err != nil {
				return appConfig{}, false, err
			}
			return cfg, true, nil
		}
		return appConfig{}, false, err
	}

	// Allow UTF-8 BOM that some editors add.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{}, false, fmt.Errorf("parse config: %w", err)
	}

	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.Password = strings.TrimSpace(cfg.Password)
	if cfg.Username == "" || cfg.Password == "" {
		return appConfig{}, false, errors.New("config requires non-empty username and password")
	}

	return cfg, false, nil
}

func writeConfig(path string, cfg appConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
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

func uniqueFilename(dir, name string) (string, error) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	ext := filepath.Ext(name)
	target := filepath.Join(dir, name)

	for i := 1; ; i++ {
		_, err := os.Stat(target)
		if err == nil {
			target = filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
			continue
		}
		if errors.Is(err, os.ErrNotExist) {
			return target, nil
		}
		return "", err
	}
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '.', r == '-', r == '_':
			builder.WriteRune(r)
		case r == ' ':
			builder.WriteRune('-')
		}
	}

	result := builder.String()
	result = strings.Trim(result, ".-")
	if result == "" {
		return ""
	}
	return result
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		if comma := strings.Index(ip, ","); comma >= 0 {
			ip = ip[:comma]
		}
		return strings.TrimSpace(ip)
	}
	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return "-"
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil && parsed != "" {
		return parsed
	}
	return host
}

func addressForLog(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}

const pageTemplate = `<!DOCTYPE html>
<html lang="pl">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Galeria zdjec</title>
  <style>
    :root {
      color-scheme: light dark;
      font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
    }
    body {
      margin: 0;
      background: #f3f4f8;
      color: #1f2933;
    }
    .topbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 2rem;
      background: linear-gradient(135deg, #3a7bd5, #00d2ff);
      color: #fff;
      box-shadow: 0 8px 24px rgba(58, 123, 213, 0.35);
    }
    .brand {
      font-weight: 600;
      font-size: 1.25rem;
      letter-spacing: 0.04em;
    }
    .top-actions {
      display: flex;
      gap: 0.75rem;
    }
    .btn {
      border: none;
      border-radius: 999px;
      padding: 0.55rem 1.25rem;
      font-size: 0.95rem;
      font-weight: 500;
      cursor: pointer;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .btn-primary {
      background: rgba(255, 255, 255, 0.16);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.35);
    }
    .btn-secondary {
      background: rgba(15, 23, 42, 0.12);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.25);
    }
    .btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 30px rgba(15, 23, 42, 0.25);
    }
    .hidden-input {
      display: none;
    }
    .page {
      padding: 2rem clamp(1rem, 4vw, 3rem);
      max-width: 1200px;
      margin: 0 auto;
    }
    .info-panel {
      background: #ffffff;
      border-radius: 18px;
      padding: 1.5rem;
      box-shadow: 0 12px 40px rgba(15, 23, 42, 0.08);
      margin-bottom: 1.5rem;
    }
    .info-panel p {
      margin: 0 0 0.4rem;
      color: #52606d;
      font-size: 0.95rem;
    }
    .info-panel code {
      background: rgba(82, 96, 109, 0.08);
      padding: 0.25rem 0.45rem;
      border-radius: 6px;
      font-family: 'JetBrains Mono', 'SFMono-Regular', ui-monospace, monospace;
      font-size: 0.85rem;
    }
    .upload-panel {
      display: grid;
      gap: 1rem;
      margin-top: 1.25rem;
    }
    .upload-panel form {
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      align-items: center;
    }
    .upload-panel input[type="file"],
    .upload-panel input[type="text"] {
      flex: 1 1 240px;
      max-width: 320px;
      padding: 0.65rem 0.8rem;
      border-radius: 10px;
      border: 1px solid rgba(82, 96, 109, 0.2);
      font-size: 0.95rem;
    }
    .upload-panel .submit-btn {
      flex: 0 0 auto;
      background: #2563eb;
      color: #fff;
      padding: 0.65rem 1.4rem;
      border-radius: 999px;
      border: none;
      cursor: pointer;
      font-weight: 500;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .upload-panel .submit-btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 30px rgba(37, 99, 235, 0.3);
    }
    .gallery {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
      gap: 1.25rem;
    }
    .tile {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }
    .thumb {
      position: relative;
      border: none;
      border-radius: 16px;
      overflow: hidden;
      cursor: zoom-in;
      padding: 0;
      background: #131722;
      box-shadow: 0 12px 40px rgba(15, 23, 42, 0.25);
      transition: transform 0.2s ease, box-shadow 0.2s ease;
      min-height: 200px;
      display: block;
    }
    .thumb:hover {
      transform: translateY(-4px);
      box-shadow: 0 18px 55px rgba(15, 23, 42, 0.28);
    }
    .thumb img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
      transition: transform 0.25s ease;
    }
    .tile-meta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 0.75rem;
      font-size: 0.9rem;
      color: #364152;
    }
    .filename {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .delete-btn {
      border: none;
      border-radius: 8px;
      padding: 0.4rem 0.75rem;
      background: rgba(239, 68, 68, 0.16);
      color: #dc2626;
      font-weight: 500;
      cursor: pointer;
      transition: background 0.18s ease, transform 0.18s ease;
    }
    .delete-btn:hover {
      background: rgba(239, 68, 68, 0.28);
      transform: translateY(-2px);
    }
    .empty {
      text-align: center;
      font-size: 1.1rem;
      color: #52606d;
      padding: 3rem 0;
    }
    .fullscreen-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.92);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1000;
      padding: 2rem;
    }
    .fullscreen-backdrop.active {
      display: flex;
    }
    .fullscreen-backdrop img {
      max-width: 95vw;
      max-height: 95vh;
      border-radius: 20px;
      box-shadow: 0 20px 45px rgba(15, 23, 42, 0.55);
      transition: transform 0.2s ease;
      transform-origin: center;
    }
    .fullscreen-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1.5rem;
      width: min(900px, 95vw);
    }
    .zoom-controls {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      width: min(420px, 90vw);
      padding: 0.65rem 1rem;
      border-radius: 999px;
      background: rgba(12, 18, 31, 0.75);
      box-shadow: 0 15px 40px rgba(0, 0, 0, 0.35);
      color: #e2e8f0;
      position: fixed;
      bottom: 2rem;
      left: 50%;
      transform: translateX(-50%);
      z-index: 1105;
      backdrop-filter: blur(12px);
      border: 1px solid rgba(148, 163, 184, 0.35);
    }
    .zoom-controls label {
      font-size: 0.8rem;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      color: #cbd5f5;
      white-space: nowrap;
    }
    .zoom-controls input[type="range"] {
      flex: 1;
      accent-color: #38bdf8;
      cursor: pointer;
    }
    .zoom-value {
      font-variant-numeric: tabular-nums;
      min-width: 3ch;
      text-align: right;
    }
    .modal-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.6);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1100;
      padding: 1.5rem;
    }
    .modal-backdrop.active {
      display: flex;
    }
    .modal {
      background: #ffffff;
      border-radius: 20px;
      padding: 2rem;
      width: min(360px, 90vw);
      box-shadow: 0 20px 50px rgba(15, 23, 42, 0.35);
      display: grid;
      gap: 1rem;
    }
    .modal h2 {
      margin: 0;
      font-size: 1.25rem;
      color: #111827;
      text-align: center;
    }
    .modal label {
      display: grid;
      gap: 0.35rem;
      font-size: 0.9rem;
      color: #364152;
    }
    .modal input {
      padding: 0.65rem 0.8rem;
      border-radius: 10px;
      border: 1px solid rgba(82, 96, 109, 0.2);
      font-size: 0.95rem;
      background: #f8fafc;
      color: #111827;
      transition: border 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
    }
    .modal input:focus {
      outline: none;
      border-color: rgba(37, 99, 235, 0.6);
      box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
      background: #ffffff;
    }
    .modal-actions {
      display: flex;
      gap: 0.75rem;
      justify-content: center;
    }
    .modal .primary {
      flex: 1;
      background: #2563eb;
      color: #fff;
      border: none;
      border-radius: 12px;
      padding: 0.65rem;
      font-weight: 600;
      cursor: pointer;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .modal .primary:hover {
      transform: translateY(-2px);
      box-shadow: 0 12px 32px rgba(37, 99, 235, 0.35);
    }
    .modal .ghost {
      flex: 1;
      background: #f8fafc;
      border: 1px solid rgba(148, 163, 184, 0.6);
      border-radius: 12px;
      padding: 0.65rem;
      font-weight: 500;
      cursor: pointer;
      color: #364152;
      transition: transform 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
    }
    .modal .ghost:hover {
      transform: translateY(-2px);
      box-shadow: 0 12px 32px rgba(15, 23, 42, 0.12);
      background: #eef2f8;
    }
    .toast {
      position: fixed;
      bottom: 2rem;
      right: 2rem;
      background: #111827;
      color: #fff;
      padding: 0.85rem 1.25rem;
      border-radius: 12px;
      box-shadow: 0 15px 40px rgba(17, 24, 39, 0.35);
      opacity: 0;
      transform: translateY(20px);
      pointer-events: none;
      transition: opacity 0.2s ease, transform 0.2s ease;
      z-index: 1200;
      font-size: 0.95rem;
    }
    .toast.visible {
      opacity: 1;
      transform: translateY(0);
    }
    .toast[data-type="error"] {
      background: #dc2626;
      box-shadow: 0 15px 40px rgba(220, 38, 38, 0.35);
    }
    @media (max-width: 720px) {
      .topbar {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
      }
      .upload-panel form {
        flex-direction: column;
        align-items: stretch;
      }
      .toast {
        left: 1rem;
        right: 1rem;
      }
      .zoom-controls {
        flex-direction: column;
        align-items: stretch;
        border-radius: 16px;
        gap: 0.5rem;
        width: calc(100% - 2rem);
        bottom: 1rem;
        padding: 0.75rem 1rem;
      }
    }
  </style>
</head>
<body>
  <header class="topbar">
    <div class="brand">Galeria zdjec</div>
    <div class="top-actions">
      {{if .LoggedIn}}
      <label for="quickUploadInput" class="btn btn-primary" id="quickUploadTrigger">Dodaj zdjecie</label>
      <input type="file" id="quickUploadInput" class="hidden-input" accept=".jpg,.jpeg,.png,.gif,.bmp,.svg,.webp,.avif">
      <button id="logoutButton" class="btn btn-secondary" type="button">Wyloguj</button>
      {{else}}
      <button id="loginButton" class="btn btn-primary" type="button">Zaloguj</button>
      {{end}}
    </div>
  </header>
  <main class="page">
    {{if .Images}}
    <section class="gallery">
      {{range .Images}}
      <div class="tile" data-name="{{.Name}}">
        <button type="button" class="thumb" data-src="{{.URL}}" aria-label="Zobacz {{.Name}}">
          <img src="{{.URL}}" alt="{{.Name}}">
        </button>
        <div class="tile-meta">
          <span class="filename" title="{{.Name}}">{{.Name}}</span>
          {{if $.LoggedIn}}
          <button type="button" class="delete-btn" data-name="{{.Name}}">Usun</button>
          {{end}}
        </div>
      </div>
      {{end}}
    </section>
    {{else}}
    <p class="empty">Brak obrazow w katalogu.</p>
    {{end}}
  </main>

  <div class="fullscreen-backdrop" id="backdrop" role="dialog" aria-modal="true">
    <div class="fullscreen-content">
      <img id="fullImage" alt="">
      <div class="zoom-controls" id="zoomControls" hidden>
        <label for="zoomSlider">Powiekszenie</label>
        <input type="range" id="zoomSlider" min="100" max="250" step="10" value="100">
        <span class="zoom-value" id="zoomValue">100%</span>
      </div>
    </div>
  </div>

  <div class="modal-backdrop" id="loginModal">
    <form class="modal" id="loginForm" autocomplete="off">
      <h2>Panel administratora</h2>
      <label>
        Login
        <input type="text" name="username" autocomplete="off" autocapitalize="none" spellcheck="false" required>
      </label>
      <label>
        Haslo
        <input type="password" name="password" autocomplete="off" required>
      </label>
      <div class="modal-actions">
        <button class="primary" type="submit">Zaloguj</button>
        <button class="ghost" type="button" id="loginCancel">Anuluj</button>
      </div>
    </form>
  </div>

  <div class="toast" id="statusMessage" role="status" aria-live="polite"></div>

  <script>
    const backdrop = document.getElementById('backdrop');
    const fullImage = document.getElementById('fullImage');
    const loginModal = document.getElementById('loginModal');
    const loginButton = document.getElementById('loginButton');
    const logoutButton = document.getElementById('logoutButton');
    const loginForm = document.getElementById('loginForm');
    const loginCancel = document.getElementById('loginCancel');
    const uploadForm = document.getElementById('uploadForm');
    const quickUploadInput = document.getElementById('quickUploadInput');
    const messageEl = document.getElementById('statusMessage');
    const zoomSlider = document.getElementById('zoomSlider');
    const zoomControls = document.getElementById('zoomControls');
    const zoomValue = document.getElementById('zoomValue');
    let hideToast;

    function showMessage(text, type = 'info') {
      if (!messageEl) return;
      messageEl.textContent = text;
      messageEl.dataset.type = type;
      messageEl.classList.add('visible');
      clearTimeout(hideToast);
      hideToast = setTimeout(() => {
        messageEl.classList.remove('visible');
      }, 4000);
    }

    function setZoom(value) {
      if (!fullImage) return;
      const scale = value / 100;
      fullImage.style.transform = 'scale(' + scale + ')';
      if (zoomValue) {
        zoomValue.textContent = value + '%';
      }
    }

    function resetZoom() {
      if (zoomSlider) {
        zoomSlider.value = '100';
      }
      setZoom(100);
    }

    function openFullscreen(src, alt) {
      fullImage.src = src;
      fullImage.alt = alt;
      resetZoom();
      if (zoomControls) {
        zoomControls.hidden = false;
      }
      backdrop.classList.add('active');
    }

    function closeFullscreen() {
      backdrop.classList.remove('active');
      if (zoomControls) {
        zoomControls.hidden = true;
      }
      resetZoom();
      fullImage.src = '';
      fullImage.alt = '';
    }

    document.querySelectorAll('.thumb').forEach(btn => {
      btn.addEventListener('click', () => {
        const src = btn.dataset.src;
        const alt = btn.closest('.tile')?.dataset.name || '';
        if (backdrop.classList.contains('active') && fullImage.src.endsWith(src)) {
          closeFullscreen();
        } else {
          openFullscreen(src, alt);
        }
      });
    });

    backdrop.addEventListener('click', closeFullscreen);
    document.addEventListener('keydown', event => {
      if (event.key === 'Escape') {
        if (backdrop.classList.contains('active')) {
          closeFullscreen();
        }
        if (loginModal.classList.contains('active')) {
          loginModal.classList.remove('active');
        }
      }
    });

    if (loginButton) {
      loginButton.addEventListener('click', () => {
        loginModal.classList.add('active');
        loginForm.reset();
        loginForm.querySelector('input[name="username"]').focus();
      });
    }

    if (loginCancel) {
      loginCancel.addEventListener('click', () => {
        loginModal.classList.remove('active');
      });
    }

    if (loginModal) {
      loginModal.addEventListener('click', event => {
        if (event.target === loginModal) {
          loginModal.classList.remove('active');
        }
      });
    }

    async function fetchJSON(url, options = {}) {
      const response = await fetch(url, options);
      const text = await response.text();
      let data;
      try {
        data = text ? JSON.parse(text) : {};
      } catch (_) {
        data = {};
      }
      if (!response.ok) {
        const message = data.error || text || 'Wystapil blad';
        throw new Error(message);
      }
      return data;
    }

    if (loginForm) {
      loginForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(loginForm);
        const payload = {
          username: formData.get('username'),
          password: formData.get('password')
        };
        try {
          await fetchJSON('/api/login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    if (logoutButton) {
      logoutButton.addEventListener('click', async () => {
        try {
          await fetchJSON('/api/logout', { method: 'POST' });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    if (uploadForm) {
      uploadForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(uploadForm);
        try {
          await fetchJSON('/api/upload', {
            method: 'POST',
            body: formData
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    if (quickUploadInput) {
      quickUploadInput.addEventListener('change', async () => {
        if (!quickUploadInput.files.length) {
          return;
        }
        const formData = new FormData();
        formData.append('file', quickUploadInput.files[0]);
        try {
          await fetchJSON('/api/upload', {
            method: 'POST',
            body: formData
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        } finally {
          quickUploadInput.value = '';
        }
      });
    }

    if (zoomSlider) {
      zoomSlider.addEventListener('input', () => {
        const value = Number(zoomSlider.value) || 100;
        setZoom(value);
      });
    }

    if (zoomControls) {
      ['click', 'mousedown', 'pointerdown', 'touchstart'].forEach(type => {
        zoomControls.addEventListener(type, event => {
          event.stopPropagation();
        });
      });
    }

    document.querySelectorAll('.delete-btn').forEach(btn => {
      btn.addEventListener('click', async event => {
        event.stopPropagation();
        const name = btn.dataset.name;
        if (!name) return;
        const confirmed = confirm('Czy na pewno chcesz usunac plik "' + name + '"?');
        if (!confirmed) return;
        try {
          await fetchJSON('/api/delete', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name})
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    });
  </script>
</body>
</html>
`
