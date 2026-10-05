package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Rayten225/wsociety-WWchan/internal/models"
	"github.com/Rayten225/wsociety-WWchan/internal/service"
)

type Handler struct {
	service *service.Service
	mw      *Middleware
}

func NewHandler(srv *service.Service, mw *Middleware) *Handler {
	return &Handler{service: srv, mw: mw}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/register", h.register)
	mux.HandleFunc("POST /api/login", h.login)
	mux.HandleFunc("GET /api/categories", h.getCategories)

	mux.HandleFunc("GET /api/posts", h.getPosts)
	mux.HandleFunc("GET /api/posts/{id}", h.getPostByID)
	
	// Защищенные роуты
	mux.HandleFunc("POST /api/posts", h.mw.Auth(h.createPost))
	
	mux.HandleFunc("GET /api/posts/{id}/comments", h.getComments)
	mux.HandleFunc("POST /api/posts/{id}/comments", h.mw.Auth(h.createComment))

	mux.HandleFunc("POST /api/posts/{id}/reactions", h.mw.Auth(h.reactPost))
	mux.HandleFunc("POST /api/comments/{id}/reactions", h.mw.Auth(h.reactComment))
	
	mux.HandleFunc("POST /api/upload", h.mw.Auth(h.uploadFile))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Register(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	WriteJSON(w, http.StatusOK, models.TokenResponse{Token: token})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), req)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	WriteJSON(w, http.StatusOK, models.TokenResponse{Token: token})
}

func (h *Handler) getCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.service.GetCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, cats)
}

func (h *Handler) getPosts(w http.ResponseWriter, r *http.Request) {
	limit := ParseInt(r.URL.Query().Get("limit"), 20)
	offset := ParseInt(r.URL.Query().Get("offset"), 0)

	posts, err := h.service.GetHotPosts(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, posts)
}

func (h *Handler) getPostByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPostByID(r.Context(), id)
	if err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	WriteJSON(w, http.StatusOK, post)
}

func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postID, err := h.service.CreatePost(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]int{"id": postID})
}

func (h *Handler) getComments(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	comments, err := h.service.GetComments(r.Context(), postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, comments)
}

func (h *Handler) createComment(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	commentID, err := h.service.CreateComment(r.Context(), userID, postID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]int{"id": commentID})
}

func (h *Handler) reactPost(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req models.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.SetReaction(r.Context(), userID, &postID, nil, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) reactComment(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req models.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.SetReaction(r.Context(), userID, nil, &commentID, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse max 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image field required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	
	os.MkdirAll("uploads", os.ModePerm)
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Возвращаем публичный путь
	WriteJSON(w, http.StatusOK, map[string]string{"url": "/uploads/" + filename})
}
