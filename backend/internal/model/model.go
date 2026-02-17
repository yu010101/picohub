package model

import "time"

type User struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Skill struct {
	ID            int64     `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Version       string    `json:"version"`
	Category      string    `json:"category"`
	AuthorID      int64     `json:"author_id"`
	AuthorName    string    `json:"author_name,omitempty"`
	FilePath      string    `json:"-"`
	FileHash      string    `json:"file_hash"`
	ScanStatus    string    `json:"scan_status"`
	DownloadCount int64     `json:"download_count"`
	IsFeatured    bool      `json:"is_featured"`
	Tags          string    `json:"tags"`
	AvgRating     float64   `json:"avg_rating"`
	ReviewCount   int64     `json:"review_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Review struct {
	ID         int64     `json:"id"`
	SkillID    int64     `json:"skill_id"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username,omitempty"`
	Rating     int       `json:"rating"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Request/Response types

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateReviewRequest struct {
	Rating int    `json:"rating"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type SkillListParams struct {
	Query    string
	Category string
	Sort     string
	Page     int
	PerPage  int
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
