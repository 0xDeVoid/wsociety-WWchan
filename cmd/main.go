package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/Rayten225/wsociety-WWchan/internal/api"
	"github.com/Rayten225/wsociety-WWchan/internal/service"
	"github.com/Rayten225/wsociety-WWchan/internal/storage"
	"github.com/Rayten225/wsociety-WWchan/pkg/jwt"
)

func main() {
	log.Println("Starting WWchan Server...")

	// 1. Подключение к БД
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://wwchan:secretpassword@localhost:5432/wwchan_db?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Connected to PostgreSQL successfully!")

	// 2. Инициализация слоёв
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretjwtkey_please_change"
	}
	jwtMgr := jwt.NewTokenManager(jwtSecret)
	
	store := storage.NewStorage(pool)
	srv := service.NewService(store, jwtMgr)
	mw := api.NewMiddleware(jwtMgr)
	handler := api.NewHandler(srv, mw)

	// 3. Настройка маршрутизации (net/http 1.22+)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// API маршруты
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Раздача загруженных файлов
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads")))
	mux.Handle("GET /uploads/", fileServer)

	// Раздача статики React-фронтенда
	distDir := "./frontend/dist"
	fs := http.FileServer(http.Dir(distDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Если запрашивают файл, который есть (например, .js, .css), отдаём его
		path := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Иначе отдаём index.html для роутинга на стороне клиента (React Router)
			if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/uploads/") {
				http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
				return
			}
		}
		fs.ServeHTTP(w, r)
	})

	// 4. Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
