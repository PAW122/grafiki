package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
