package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Rayten225/wsociety-WWchan/internal/models"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

// === ОПЕРАЦИИ С ПОЛЬЗОВАТЕЛЯМИ ===

func (s *Storage) CreateUser(ctx context.Context, username, passwordHash string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, "INSERT INTO ref_users (username, password_hash) VALUES ($1, $2) RETURNING id", username, passwordHash).Scan(&id)
	return id, err
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx, "SELECT id, username, password_hash, created_at FROM ref_users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// === ОПЕРАЦИИ С КАТЕГОРИЯМИ ===

func (s *Storage) GetCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, name, slug FROM ref_categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// === ОПЕРАЦИИ С ПОСТАМИ ===

func (s *Storage) CreatePost(ctx context.Context, p *models.Post) (int, error) {
	var id int
	// Используем хранимую функцию
	err := s.pool.QueryRow(ctx, "SELECT func_create_post($1, $2, $3, $4, $5)",
		p.UserID, p.CategoryID, p.Title, p.Content, p.ImageURL).Scan(&id)
	return id, err
}

func (s *Storage) GetHotPosts(ctx context.Context, limit, offset int) ([]models.Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, u.username, p.category_id, p.title, p.content, p.image_url, p.created_at,
			COALESCE(SUM(r.reaction_value), 0) as score
		FROM op_posts p
		JOIN ref_users u ON p.user_id = u.id
		LEFT JOIN op_reactions r ON p.id = r.post_id
		GROUP BY p.id, u.username
		ORDER BY score DESC, p.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.AuthorName, &p.CategoryID, &p.Title, &p.Content, &p.ImageURL, &p.CreatedAt, &p.Score); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (s *Storage) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, u.username, p.category_id, p.title, p.content, p.image_url, p.created_at,
			COALESCE(SUM(r.reaction_value), 0) as score
		FROM op_posts p
		JOIN ref_users u ON p.user_id = u.id
		LEFT JOIN op_reactions r ON p.id = r.post_id
		WHERE p.id = $1
		GROUP BY p.id, u.username
	`
	var p models.Post
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.UserID, &p.AuthorName, &p.CategoryID, &p.Title, &p.Content, &p.ImageURL, &p.CreatedAt, &p.Score,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// === ОПЕРАЦИИ С КОММЕНТАРИЯМИ ===

func (s *Storage) CreateComment(ctx context.Context, c *models.Comment) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, "SELECT func_create_comment($1, $2, $3, $4, $5)",
		c.PostID, c.UserID, c.ParentCommentID, c.Content, c.ImageURL).Scan(&id)
	return id, err
}

func (s *Storage) GetCommentsByPostID(ctx context.Context, postID int) ([]models.Comment, error) {
	query := `
		SELECT 
			c.id, c.post_id, c.user_id, u.username, c.parent_comment_id, c.content, c.image_url, c.created_at,
			COALESCE(SUM(r.reaction_value), 0) as score
		FROM op_comments c
		JOIN ref_users u ON c.user_id = u.id
		LEFT JOIN op_reactions r ON c.id = r.comment_id
		WHERE c.post_id = $1
		GROUP BY c.id, u.username
		ORDER BY c.created_at ASC
	`
	rows, err := s.pool.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.AuthorName, &c.ParentCommentID, &c.Content, &c.ImageURL, &c.CreatedAt, &c.Score); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// === ОПЕРАЦИИ С РЕАКЦИЯМИ ===

func (s *Storage) SetReaction(ctx context.Context, userID int, postID *int, commentID *int, value int16) error {
	_, err := s.pool.Exec(ctx, "SELECT func_set_reaction($1, $2, $3, $4)", userID, postID, commentID, value)
	return err
}
