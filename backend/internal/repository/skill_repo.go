package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/yu01/picohub/internal/model"
)

type SkillRepository struct {
	db *sql.DB
}

func NewSkillRepository(db *sql.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) Create(s *model.Skill) error {
	res, err := r.db.Exec(
		`INSERT INTO skills (slug, name, description, version, category, author_id, file_path, file_hash, scan_status, is_featured, tags)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Slug, s.Name, s.Description, s.Version, s.Category, s.AuthorID,
		s.FilePath, s.FileHash, s.ScanStatus, s.IsFeatured, s.Tags,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	s.ID = id
	return nil
}

func (r *SkillRepository) FindBySlug(slug string) (*model.Skill, error) {
	s := &model.Skill{}
	err := r.db.QueryRow(`
		SELECT s.id, s.slug, s.name, s.description, s.version, s.category,
		       s.author_id, u.username, s.file_path, s.file_hash, s.scan_status,
		       s.download_count, s.is_featured, s.tags,
		       COALESCE(AVG(r.rating), 0), COUNT(r.id),
		       s.created_at, s.updated_at
		FROM skills s
		JOIN users u ON u.id = s.author_id
		LEFT JOIN reviews r ON r.skill_id = s.id
		WHERE s.slug = ?
		GROUP BY s.id`, slug,
	).Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Version, &s.Category,
		&s.AuthorID, &s.AuthorName, &s.FilePath, &s.FileHash, &s.ScanStatus,
		&s.DownloadCount, &s.IsFeatured, &s.Tags,
		&s.AvgRating, &s.ReviewCount,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SkillRepository) List(p model.SkillListParams) ([]model.Skill, int64, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 50 {
		p.PerPage = 12
	}

	var where []string
	var args []interface{}

	if p.Category != "" {
		where = append(where, "s.category = ?")
		args = append(args, p.Category)
	}

	if p.Query != "" {
		where = append(where, "s.id IN (SELECT rowid FROM skills_fts WHERE skills_fts MATCH ?)")
		args = append(args, p.Query)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	orderBy := "s.created_at DESC"
	switch p.Sort {
	case "downloads":
		orderBy = "s.download_count DESC"
	case "rating":
		orderBy = "avg_rating DESC"
	case "name":
		orderBy = "s.name ASC"
	case "newest":
		orderBy = "s.created_at DESC"
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM skills s %s", whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (p.Page - 1) * p.PerPage
	query := fmt.Sprintf(`
		SELECT s.id, s.slug, s.name, s.description, s.version, s.category,
		       s.author_id, u.username, s.file_path, s.file_hash, s.scan_status,
		       s.download_count, s.is_featured, s.tags,
		       COALESCE(AVG(r.rating), 0) as avg_rating, COUNT(r.id) as review_count,
		       s.created_at, s.updated_at
		FROM skills s
		JOIN users u ON u.id = s.author_id
		LEFT JOIN reviews r ON r.skill_id = s.id
		%s
		GROUP BY s.id
		ORDER BY %s
		LIMIT ? OFFSET ?`, whereClause, orderBy)

	args = append(args, p.PerPage, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var skills []model.Skill
	for rows.Next() {
		var s model.Skill
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Version, &s.Category,
			&s.AuthorID, &s.AuthorName, &s.FilePath, &s.FileHash, &s.ScanStatus,
			&s.DownloadCount, &s.IsFeatured, &s.Tags,
			&s.AvgRating, &s.ReviewCount,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		skills = append(skills, s)
	}
	return skills, total, nil
}

func (r *SkillRepository) Featured() ([]model.Skill, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.slug, s.name, s.description, s.version, s.category,
		       s.author_id, u.username, s.file_path, s.file_hash, s.scan_status,
		       s.download_count, s.is_featured, s.tags,
		       COALESCE(AVG(r.rating), 0), COUNT(r.id),
		       s.created_at, s.updated_at
		FROM skills s
		JOIN users u ON u.id = s.author_id
		LEFT JOIN reviews r ON r.skill_id = s.id
		WHERE s.is_featured = 1
		GROUP BY s.id
		ORDER BY s.download_count DESC
		LIMIT 6`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []model.Skill
	for rows.Next() {
		var s model.Skill
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Version, &s.Category,
			&s.AuthorID, &s.AuthorName, &s.FilePath, &s.FileHash, &s.ScanStatus,
			&s.DownloadCount, &s.IsFeatured, &s.Tags,
			&s.AvgRating, &s.ReviewCount,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *SkillRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM skills WHERE id = ?", id)
	return err
}

func (r *SkillRepository) Update(s *model.Skill) error {
	_, err := r.db.Exec(
		`UPDATE skills SET description = ?, is_featured = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		s.Description, s.IsFeatured, s.ID,
	)
	return err
}

func (r *SkillRepository) FindByAuthorID(authorID int64) ([]model.Skill, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.slug, s.name, s.description, s.version, s.category,
		       s.author_id, u.username, s.file_path, s.file_hash, s.scan_status,
		       s.download_count, s.is_featured, s.tags,
		       COALESCE(AVG(r.rating), 0), COUNT(r.id),
		       s.created_at, s.updated_at
		FROM skills s
		JOIN users u ON u.id = s.author_id
		LEFT JOIN reviews r ON r.skill_id = s.id
		WHERE s.author_id = ?
		GROUP BY s.id
		ORDER BY s.created_at DESC`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []model.Skill
	for rows.Next() {
		var s model.Skill
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Version, &s.Category,
			&s.AuthorID, &s.AuthorName, &s.FilePath, &s.FileHash, &s.ScanStatus,
			&s.DownloadCount, &s.IsFeatured, &s.Tags,
			&s.AvgRating, &s.ReviewCount,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *SkillRepository) IncrementDownload(slug string) error {
	_, err := r.db.Exec("UPDATE skills SET download_count = download_count + 1 WHERE slug = ?", slug)
	return err
}

func (r *SkillRepository) Categories() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT category FROM skills ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}
