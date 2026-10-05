package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"github.com/Rayten225/wsociety-WWchan/pkg/jwt"
)

type Middleware struct {
	jwtMgr *jwt.TokenManager
}

func NewMiddleware(jwtMgr *jwt.TokenManager) *Middleware {
	return &Middleware{jwtMgr: jwtMgr}
}

func (m *Middleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing auth token", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}

		userID, err := m.jwtMgr.ParseToken(parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserID(ctx context.Context) int {
	id, ok := ctx.Value("user_id").(int)
	if !ok {
		return 0
	}
	return id
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func ParseInt(str string, def int) int {
	v, err := strconv.Atoi(str)
	if err != nil {
		return def
	}
	return v
}
