package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	visibilityPublic  = "public"
	visibilityShared  = "shared"
	visibilityPrivate = "private"
)

var allowedVisibilities = map[string]struct{}{
	visibilityPublic:  {},
	visibilityShared:  {},
	visibilityPrivate: {},
}

var (
	errFolderProtected    = errors.New("nie mozna usunac folderu glownego")
	errFolderPathInvalid  = errors.New("nieprawidlowy katalog folderu")
	errFolderNameExists   = errors.New("folder o takiej nazwie juz istnieje")
	errFolderRenameFailed = errors.New("Nie udalo sie zmienic nazwy folderu")
)

type folderRecord struct {
	ID          int64
	Name        string
	Slug        string
	Path        string
	Visibility  string
	SharedToken sql.NullString
	SharedViews int
}

type folderView struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Visibility  string `json:"visibility"`
	SharedToken string `json:"sharedToken,omitempty"`
	SharedViews int    `json:"sharedViews"`
	ShareURL    string `json:"shareUrl,omitempty"`
}

func (f folderRecord) toView(baseURL string) folderView {
	view := folderView{
		ID:          f.ID,
		Name:        f.Name,
		Slug:        f.Slug,
		Visibility:  f.Visibility,
		SharedViews: f.SharedViews,
	}
	if f.SharedToken.Valid && f.SharedToken.String != "" {
		view.SharedToken = f.SharedToken.String
		if baseURL != "" {
			view.ShareURL = fmt.Sprintf("%s/shared/%s", strings.TrimSuffix(baseURL, "/"), f.SharedToken.String)
		}
	}
	return view
}

func (s *Server) folderSlugExists(slug string) (bool, error) {
	return s.folderSlugTaken(slug, 0)
}

func (s *Server) folderSlugTaken(slug string, excludeID int64) (bool, error) {
	var existingID int64
	err := s.db.QueryRow(`SELECT id FROM folders WHERE slug = ? LIMIT 1`, slug).Scan(&existingID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if excludeID != 0 && existingID == excludeID {
		return false, nil
	}
	return true, nil
}

func (s *Server) createFolder(name string) (*folderRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nazwa folderu jest wymagana")
	}

	baseSlug := sanitizeFilename(name)
	if baseSlug == "" {
		baseSlug = sanitizeFilename(strings.ReplaceAll(strings.ToLower(name), " ", "-"))
	}
	if baseSlug == "" {
		return nil, errors.New("nie udalo sie wygenerowac nazwy folderu")
	}

	slug := baseSlug
	for i := 2; ; i++ {
		exists, err := s.folderSlugExists(slug)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	relPath := slug
	fullPath := filepath.Join(s.dir, relPath)
	if err := EnsureDir(fullPath); err != nil {
		return nil, err
	}

	result, err := s.db.Exec(`INSERT INTO folders (name, slug, path, visibility) VALUES (?, ?, ?, ?)`,
		name, slug, relPath, visibilityPrivate)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.getFolderByID(id)
}

func (s *Server) listFolders(loggedIn bool) ([]folderRecord, error) {
	query := `SELECT id, name, slug, path, visibility, shared_token, shared_views FROM folders`
	var args []any
	if !loggedIn {
		query += ` WHERE visibility = ?`
		args = append(args, visibilityPublic)
	}
	query += ` ORDER BY name`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []folderRecord
	for rows.Next() {
		var rec folderRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews); err != nil {
			return nil, err
		}
		folders = append(folders, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return folders, nil
}

func (s *Server) getFolderBySlug(slug string) (*folderRecord, error) {
	var rec folderRecord
	err := s.db.QueryRow(`SELECT id, name, slug, path, visibility, shared_token, shared_views FROM folders WHERE slug = ?`, slug).
		Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Server) getFolderByID(id int64) (*folderRecord, error) {
	var rec folderRecord
	err := s.db.QueryRow(`SELECT id, name, slug, path, visibility, shared_token, shared_views FROM folders WHERE id = ?`, id).
		Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Server) getFolderByToken(token string) (*folderRecord, error) {
	var rec folderRecord
	err := s.db.QueryRow(`SELECT id, name, slug, path, visibility, shared_token, shared_views FROM folders WHERE shared_token = ?`, token).
		Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Server) updateFolderVisibility(id int64, visibility string) (*folderRecord, error) {
	if _, ok := allowedVisibilities[visibility]; !ok {
		return nil, errors.New("nieprawidlowy typ widocznosci")
	}

	if _, err := s.db.Exec(`UPDATE folders SET visibility = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, visibility, id); err != nil {
		return nil, err
	}

	rec, err := s.getFolderByID(id)
	if err != nil {
		return nil, err
	}
	if visibility == visibilityShared {
		if _, err := s.ensureSharedToken(rec.ID); err != nil {
			return nil, err
		}
		rec, err = s.getFolderByID(id)
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func (s *Server) ensureSharedToken(id int64) (string, error) {
	rec, err := s.getFolderByID(id)
	if err != nil {
		return "", err
	}
	if rec.SharedToken.Valid && rec.SharedToken.String != "" {
		return rec.SharedToken.String, nil
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	if _, err := s.db.Exec(`UPDATE folders SET shared_token = ?, shared_views = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, token, id); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Server) regenerateSharedToken(id int64) (*folderRecord, error) {
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(`UPDATE folders SET shared_token = ?, shared_views = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, token, id); err != nil {
		return nil, err
	}
	return s.getFolderByID(id)
}

func (s *Server) incrementSharedViews(id int64) error {
	_, err := s.db.Exec(`UPDATE folders SET shared_views = shared_views + 1 WHERE id = ?`, id)
	return err
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func folderURLPrefix(path string) string {
	if path == "" {
		return ""
	}
	return strings.Trim(strings.ReplaceAll(filepath.ToSlash(path), "//", "/"), "/")
}

func (s *Server) renameFolder(id int64, name string) (*folderRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nazwa folderu jest wymagana")
	}

	folder, err := s.getFolderByID(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(folder.Path) == "" {
		return nil, errFolderProtected
	}

	baseSlug := sanitizeFilename(name)
	if baseSlug == "" {
		baseSlug = sanitizeFilename(strings.ReplaceAll(strings.ToLower(name), " ", "-"))
	}
	if baseSlug == "" {
		return nil, errors.New("nie udalo sie wygenerowac nazwy folderu")
	}

	slug := baseSlug
	for i := 2; ; i++ {
		taken, lookupErr := s.folderSlugTaken(slug, id)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if !taken {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	slugChanged := slug != folder.Slug
	nameChanged := name != folder.Name
	if !slugChanged && !nameChanged {
		return folder, nil
	}

	newRelPath := folder.Path
	baseDir := filepath.Clean(s.dir)
	oldPath := filepath.Join(baseDir, folder.Path)
	newPath := filepath.Join(baseDir, slug)

	if slugChanged {
		if oldPath == baseDir || !strings.HasPrefix(oldPath, baseDir+string(os.PathSeparator)) {
			return nil, errFolderPathInvalid
		}
		if !strings.HasPrefix(newPath, baseDir+string(os.PathSeparator)) {
			return nil, errFolderPathInvalid
		}
		if _, err := os.Stat(newPath); err == nil {
			return nil, errFolderNameExists
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return nil, fmt.Errorf("%w: %v", errFolderRenameFailed, err)
		}
		newRelPath = slug
	}

	_, err = s.db.Exec(`UPDATE folders SET name = ?, slug = ?, path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, slug, newRelPath, id)
	if err != nil {
		if slugChanged {
			if revertErr := os.Rename(newPath, oldPath); revertErr != nil {
				return nil, fmt.Errorf("%w: %v (rollback: %v)", errFolderRenameFailed, err, revertErr)
			}
		}
		return nil, err
	}

	return s.getFolderByID(id)
}

func (s *Server) deleteFolder(id int64) error {
	folder, err := s.getFolderByID(id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(folder.Path) == "" {
		return errFolderProtected
	}

	baseDir := filepath.Clean(s.dir)
	targetDir := filepath.Join(baseDir, folder.Path)
	cleanTarget := filepath.Clean(targetDir)

	if cleanTarget == baseDir || !strings.HasPrefix(cleanTarget, baseDir+string(os.PathSeparator)) {
		return errFolderPathInvalid
	}

	if err := os.RemoveAll(cleanTarget); err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM folders WHERE id = ?`, id)
	return err
}
