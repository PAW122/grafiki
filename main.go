package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

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

	dbPath := filepath.Join(filepath.Dir(configPath), defaultDatabaseName)
	db, err := openDatabase(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

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
		db:       db,
	}

	mux := http.NewServeMux()
	mux.Handle("/", srv)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(dir))))
	mux.HandleFunc("/api/login", srv.handleLogin)
	mux.HandleFunc("/api/logout", srv.handleLogout)
	mux.HandleFunc("/api/upload", srv.handleUpload)
	mux.HandleFunc("/api/delete", srv.handleDelete)
	mux.HandleFunc("/api/folders", srv.handleFolders)
	mux.HandleFunc("/api/folders/", srv.handleFolderByID)
	mux.HandleFunc("/shared/", srv.handleSharedFolder)

	log.Printf("Serving gallery from %s at http://%s (config: %s, logs: %s)", dir, addressForLog(*addrFlag), configPath, logsPath)
	if err := http.ListenAndServe(*addrFlag, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
