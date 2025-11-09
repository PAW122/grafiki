package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

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
