package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"grafiki/internal/app"
)

const (
	defaultGalleryDir   = "galery"
	defaultConfigPath   = "config.json"
	defaultDatabaseName = "gallery.db"
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

	if err := app.EnsureDir(dir); err != nil {
		log.Fatalf("ensure gallery dir %q: %v", dir, err)
	}

	configPath, err := filepath.Abs(*configFlag)
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}

	cfg, created, err := app.LoadOrCreateConfig(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if created {
		log.Printf("Created default config at %s (edit to change admin credentials)", configPath)
	}

	dbPath := filepath.Join(filepath.Dir(configPath), defaultDatabaseName)
	db, err := app.OpenDatabase(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	logsPath := filepath.Join(filepath.Dir(configPath), "logs")
	reqLogger, err := app.NewRequestLogger(logsPath)
	if err != nil {
		log.Fatalf("open logs file: %v", err)
	}
	defer reqLogger.Close()

	faviconPath := filepath.Join(filepath.Dir(configPath), "Tsu.ico")
	if _, err := os.Stat(faviconPath); err != nil {
		faviconPath = ""
	}

	tmpl := template.Must(template.New("gallery").Parse(app.PageTemplate))
	srv, err := app.NewServer(app.ServerOptions{
		Dir:      dir,
		Config:   cfg,
		Template: tmpl,
		Sessions: app.NewSessionStore(15 * time.Minute),
		Logger:   reqLogger,
		DB:       db,
		Favicon:  faviconPath,
	})
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	log.Printf(
		"Serving gallery from %s at http://%s (config: %s, logs: %s)",
		dir,
		app.AddressForLog(*addrFlag),
		configPath,
		logsPath,
	)
	if err := http.ListenAndServe(*addrFlag, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
