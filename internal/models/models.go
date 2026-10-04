package models

import "time"

// Пользователь
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // не отдаем в JSON
	CreatedAt    time.Time `json:"created_at"`
}

// Категория
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Пост
type Post struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	AuthorName string    `json:"author_name"`
	CategoryID int       `json:"category_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	ImageURL   *string   `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
	Score      int       `json:"score"` // Сумма реакций
}

// Комментарий
type Comment struct {
	ID              int       `json:"id"`
	PostID          int       `json:"post_id"`
	UserID          int       `json:"user_id"`
	AuthorName      string    `json:"author_name"`
	ParentCommentID *int      `json:"parent_comment_id"`
	Content         string    `json:"content"`
	ImageURL        *string   `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	Score           int       `json:"score"`
}

// Запросы/Ответы для API

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type CreatePostRequest struct {
	CategoryID int     `json:"category_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	ImageURL   *string `json:"image_url,omitempty"`
}

type CreateCommentRequest struct {
	Content         string  `json:"content"`
	ParentCommentID *int    `json:"parent_comment_id,omitempty"`
	ImageURL        *string `json:"image_url,omitempty"`
}

type ReactionRequest struct {
	Value int16 `json:"value"` // 1 или -1
}
