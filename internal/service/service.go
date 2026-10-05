package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/Rayten225/wsociety-WWchan/internal/models"
	"github.com/Rayten225/wsociety-WWchan/internal/storage"
	"github.com/Rayten225/wsociety-WWchan/pkg/jwt"
)

type Service struct {
	storage *storage.Storage
	jwtMgr  *jwt.TokenManager
}

func NewService(store *storage.Storage, jwtMgr *jwt.TokenManager) *Service {
	return &Service{
		storage: store,
		jwtMgr:  jwtMgr,
	}
}

// === AUTH ===

func (s *Service) Register(ctx context.Context, req models.RegisterRequest) (string, error) {
	if req.Username == "" || req.Password == "" {
		return "", errors.New("empty credentials")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	userID, err := s.storage.CreateUser(ctx, req.Username, string(hash))
	if err != nil {
		return "", err // probably user exists
	}

	return s.jwtMgr.GenerateToken(userID)
}

func (s *Service) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	user, err := s.storage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.jwtMgr.GenerateToken(user.ID)
}

// === CATEGORIES ===

func (s *Service) GetCategories(ctx context.Context) ([]models.Category, error) {
	return s.storage.GetCategories(ctx)
}

// === POSTS ===

func (s *Service) CreatePost(ctx context.Context, userID int, req models.CreatePostRequest) (int, error) {
	if req.Title == "" || req.Content == "" {
		return 0, errors.New("title and content are required")
	}
	post := &models.Post{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Content:    req.Content,
		ImageURL:   req.ImageURL,
	}
	return s.storage.CreatePost(ctx, post)
}

func (s *Service) GetHotPosts(ctx context.Context, limit, offset int) ([]models.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.storage.GetHotPosts(ctx, limit, offset)
}

func (s *Service) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	return s.storage.GetPostByID(ctx, id)
}

// === COMMENTS ===

func (s *Service) CreateComment(ctx context.Context, userID, postID int, req models.CreateCommentRequest) (int, error) {
	if req.Content == "" && req.ImageURL == nil {
		return 0, errors.New("content or image is required")
	}
	c := &models.Comment{
		PostID:          postID,
		UserID:          userID,
		ParentCommentID: req.ParentCommentID,
		Content:         req.Content,
		ImageURL:        req.ImageURL,
	}
	return s.storage.CreateComment(ctx, c)
}

func (s *Service) GetComments(ctx context.Context, postID int) ([]models.Comment, error) {
	return s.storage.GetCommentsByPostID(ctx, postID)
}

// === REACTIONS ===

func (s *Service) SetReaction(ctx context.Context, userID int, postID *int, commentID *int, value int16) error {
	if value != 1 && value != -1 {
		return errors.New("invalid reaction value")
	}
	if (postID == nil && commentID == nil) || (postID != nil && commentID != nil) {
		return errors.New("target must be exactly one: post or comment")
	}
	return s.storage.SetReaction(ctx, userID, postID, commentID, value)
}
