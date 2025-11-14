package app

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type submissionGroupRecord struct {
	ID          int64
	Name        string
	Slug        string
	Path        string
	Visibility  string
	SharedToken sql.NullString
	SharedViews int
}

type submissionGroupView struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Visibility  string `json:"visibility"`
	SharedToken string `json:"sharedToken,omitempty"`
	SharedViews int    `json:"sharedViews"`
	ShareURL    string `json:"shareUrl,omitempty"`
}

type submissionEntryRecord struct {
	ID               int64
	GroupID          int64
	FileName         string
	OriginalName     string
	UploaderName     string
	MimeType         sql.NullString
	SizeBytes        int64
	CreatedAt        time.Time
	ContributorToken string
}

type submissionEntryView struct {
	ID          int64
	Name        string
	Original    string
	URL         string
	DownloadURL string
	MIME        string
	SizeLabel   string
	UploadedBy  string
	UploadedAt  string
	IsImage     bool
	IsPDF       bool
}

func (g submissionGroupRecord) toView(baseURL string) submissionGroupView {
	view := submissionGroupView{
		ID:          g.ID,
		Name:        g.Name,
		Slug:        g.Slug,
		Visibility:  g.Visibility,
		SharedViews: g.SharedViews,
	}
	if g.SharedToken.Valid && g.SharedToken.String != "" {
		view.SharedToken = g.SharedToken.String
	}
	if baseURL != "" {
		switch g.Visibility {
		case visibilityPublic:
			view.ShareURL = fmt.Sprintf("%s/submitted/%s", strings.TrimSuffix(baseURL, "/"), g.Slug)
		case visibilityShared:
			token := view.SharedToken
			if token != "" {
				view.ShareURL = fmt.Sprintf("%s/submitted/shared/%s", strings.TrimSuffix(baseURL, "/"), token)
			}
		}
	}
	return view
}

func (s *Server) submissionsRoot() string {
	return s.submissionsDir
}

func (s *Server) submissionGroupDir(rec *submissionGroupRecord) string {
	return filepath.Join(s.submissionsRoot(), rec.Path)
}

func (s *Server) ensureSubmissionDir(rec *submissionGroupRecord) error {
	return EnsureDir(s.submissionGroupDir(rec))
}

func (s *Server) createSubmissionGroup(name string) (*submissionGroupRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nazwa grupy jest wymagana")
	}

	baseSlug := sanitizeFilename(name)
	if baseSlug == "" {
		baseSlug = sanitizeFilename(strings.ReplaceAll(strings.ToLower(name), " ", "-"))
	}
	if baseSlug == "" {
		return nil, errors.New("nie udalo sie wygenerowac slug")
	}

	slug := baseSlug
	for i := 2; ; i++ {
		taken, err := s.submissionGroupSlugTaken(slug, 0)
		if err != nil {
			return nil, err
		}
		if !taken {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	result, err := s.db.Exec(`INSERT INTO submission_groups (name, slug, path, visibility) VALUES (?, ?, ?, ?)`,
		name, slug, slug, visibilityPrivate)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	group, err := s.getSubmissionGroupByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureSubmissionDir(group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *Server) listSubmissionGroups(loggedIn bool) ([]submissionGroupRecord, error) {
	query := `SELECT id, name, slug, path, visibility, shared_token, shared_views FROM submission_groups`
	var args []any
	if !loggedIn {
		query += ` WHERE visibility = ?`
		args = append(args, visibilityPublic)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []submissionGroupRecord
	for rows.Next() {
		var rec submissionGroupRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews); err != nil {
			return nil, err
		}
		groups = append(groups, rec)
	}
	return groups, rows.Err()
}

func (s *Server) submissionGroupSlugTaken(slug string, excludeID int64) (bool, error) {
	var existingID int64
	err := s.db.QueryRow(`SELECT id FROM submission_groups WHERE slug = ? LIMIT 1`, slug).Scan(&existingID)
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

func (s *Server) getSubmissionGroupBySlug(slug string) (*submissionGroupRecord, error) {
	var rec submissionGroupRecord
	err := s.db.QueryRow(`SELECT id, name, slug, path, visibility, shared_token, shared_views FROM submission_groups WHERE slug = ?`, slug).
		Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Server) getSubmissionGroupByID(id int64) (*submissionGroupRecord, error) {
	var rec submissionGroupRecord
	err := s.db.QueryRow(`SELECT id, name, slug, path, visibility, shared_token, shared_views FROM submission_groups WHERE id = ?`, id).
		Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Server) getSubmissionGroupByToken(token string) (*submissionGroupRecord, error) {
	var rec submissionGroupRecord
	err := s.db.QueryRow(`SELECT id, name, slug, path, visibility, shared_token, shared_views FROM submission_groups WHERE shared_token = ?`, token).
		Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.Path, &rec.Visibility, &rec.SharedToken, &rec.SharedViews)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Server) updateSubmissionGroupVisibility(id int64, visibility string) (*submissionGroupRecord, error) {
	if _, ok := allowedVisibilities[visibility]; !ok {
		return nil, errors.New("nieprawidlowy typ widocznosci")
	}
	if _, err := s.db.Exec(`UPDATE submission_groups SET visibility = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, visibility, id); err != nil {
		return nil, err
	}
	group, err := s.getSubmissionGroupByID(id)
	if err != nil {
		return nil, err
	}
	if visibility == visibilityShared {
		if _, err := s.ensureSubmissionSharedToken(id); err != nil {
			return nil, err
		}
		return s.getSubmissionGroupByID(id)
	}
	return group, nil
}

func (s *Server) ensureSubmissionSharedToken(id int64) (string, error) {
	group, err := s.getSubmissionGroupByID(id)
	if err != nil {
		return "", err
	}
	if group.SharedToken.Valid && group.SharedToken.String != "" {
		return group.SharedToken.String, nil
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	if _, err := s.db.Exec(`UPDATE submission_groups SET shared_token = ?, shared_views = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, token, id); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Server) regenerateSubmissionSharedToken(id int64) (*submissionGroupRecord, error) {
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(`UPDATE submission_groups SET shared_token = ?, shared_views = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, token, id); err != nil {
		return nil, err
	}
	return s.getSubmissionGroupByID(id)
}

func (s *Server) deleteSubmissionGroup(id int64) error {
	group, err := s.getSubmissionGroupByID(id)
	if err != nil {
		return err
	}
	dir := s.submissionGroupDir(group)
	if strings.TrimSpace(group.Path) != "" {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	_, err = s.db.Exec(`DELETE FROM submission_groups WHERE id = ?`, id)
	return err
}

func (s *Server) renameSubmissionGroup(id int64, name string) (*submissionGroupRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nazwa grupy jest wymagana")
	}
	group, err := s.getSubmissionGroupByID(id)
	if err != nil {
		return nil, err
	}

	baseSlug := sanitizeFilename(name)
	if baseSlug == "" {
		baseSlug = sanitizeFilename(strings.ReplaceAll(strings.ToLower(name), " ", "-"))
	}
	if baseSlug == "" {
		return nil, errors.New("nie udalo sie wygenerowac nazwy")
	}

	slug := baseSlug
	for i := 2; ; i++ {
		taken, lookupErr := s.submissionGroupSlugTaken(slug, group.ID)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if !taken {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	oldDir := s.submissionGroupDir(group)
	newRelPath := group.Path
	if slug != group.Slug {
		newRelPath = slug
		newDir := filepath.Join(s.submissionsRoot(), newRelPath)
		if err := EnsureDir(filepath.Dir(newDir)); err != nil {
			return nil, err
		}
		if _, err := os.Stat(oldDir); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err := EnsureDir(oldDir); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		if err := os.Rename(oldDir, newDir); err != nil {
			return nil, err
		}
	}

	if _, err := s.db.Exec(`UPDATE submission_groups SET name = ?, slug = ?, path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, slug, newRelPath, id); err != nil {
		return nil, err
	}

	return s.getSubmissionGroupByID(id)
}

func (s *Server) incrementSubmissionSharedViews(id int64) error {
	_, err := s.db.Exec(`UPDATE submission_groups SET shared_views = shared_views + 1 WHERE id = ?`, id)
	return err
}

func (s *Server) submissionEntriesForGroup(group *submissionGroupRecord, viewerToken string, loggedIn bool) ([]submissionEntryView, error) {
	if group == nil {
		return nil, nil
	}
	query := `SELECT id, group_id, filename, original_name, uploader_name, mime_type, size_bytes, created_at, contributor_token
		FROM submissions WHERE group_id = ?`
	var args []any
	args = append(args, group.ID)
	if !loggedIn && group.Visibility == visibilityShared {
		query += ` AND contributor_token = ?`
		args = append(args, viewerToken)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []submissionEntryView
	for rows.Next() {
		var rec submissionEntryRecord
		if err := rows.Scan(&rec.ID, &rec.GroupID, &rec.FileName, &rec.OriginalName, &rec.UploaderName, &rec.MimeType, &rec.SizeBytes, &rec.CreatedAt, &rec.ContributorToken); err != nil {
			return nil, err
		}
		url := fmt.Sprintf("/submitted/file/%d", rec.ID)
		entry := submissionEntryView{
			ID:          rec.ID,
			Name:        rec.UploaderName,
			Original:    rec.OriginalName,
			URL:         url,
			DownloadURL: url + "?download=1",
			MIME:        rec.MimeType.String,
			SizeLabel:   humanize.Bytes(uint64(rec.SizeBytes)),
			UploadedBy:  rec.UploaderName,
			UploadedAt:  rec.CreatedAt.Format("02.01.2006 15:04"),
			IsImage:     strings.HasPrefix(strings.ToLower(rec.MimeType.String), "image/") || isImageFile(rec.OriginalName),
			IsPDF:       strings.EqualFold(filepath.Ext(rec.OriginalName), ".pdf"),
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *Server) getSubmissionEntry(id int64) (*submissionEntryRecord, *submissionGroupRecord, error) {
	var rec submissionEntryRecord
	err := s.db.QueryRow(`SELECT id, group_id, filename, original_name, uploader_name, mime_type, size_bytes, created_at, contributor_token
		FROM submissions WHERE id = ?`, id).Scan(
		&rec.ID,
		&rec.GroupID,
		&rec.FileName,
		&rec.OriginalName,
		&rec.UploaderName,
		&rec.MimeType,
		&rec.SizeBytes,
		&rec.CreatedAt,
		&rec.ContributorToken,
	)
	if err != nil {
		return nil, nil, err
	}
	group, err := s.getSubmissionGroupByID(rec.GroupID)
	if err != nil {
		return nil, nil, err
	}
	return &rec, group, nil
}

func submissionViewerTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	cookie, err := r.Cookie(submissionViewerCookie)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func (s *Server) ensureSubmissionViewerToken(w http.ResponseWriter, r *http.Request) string {
	token := submissionViewerTokenFromRequest(r)
	if token != "" {
		return token
	}
	newToken, err := randomToken()
	if err != nil {
		newToken = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	http.SetCookie(w, &http.Cookie{
		Name:     submissionViewerCookie,
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	return newToken
}
