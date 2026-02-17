package repository

import (
	"database/sql"

	"github.com/yu01/picohub/internal/model"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(rev *model.Review) error {
	res, err := r.db.Exec(
		`INSERT INTO reviews (skill_id, user_id, rating, title, body) VALUES (?, ?, ?, ?, ?)`,
		rev.SkillID, rev.UserID, rev.Rating, rev.Title, rev.Body,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	rev.ID = id
	return nil
}

func (r *ReviewRepository) ListBySkillID(skillID int64) ([]model.Review, error) {
	rows, err := r.db.Query(`
		SELECT r.id, r.skill_id, r.user_id, u.username, r.rating, r.title, r.body, r.created_at, r.updated_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		WHERE r.skill_id = ?
		ORDER BY r.created_at DESC`, skillID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var rev model.Review
		if err := rows.Scan(&rev.ID, &rev.SkillID, &rev.UserID, &rev.Username,
			&rev.Rating, &rev.Title, &rev.Body, &rev.CreatedAt, &rev.UpdatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}
