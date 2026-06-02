package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"conversorj/internal/converter"
	"conversorj/internal/handler"
	"conversorj/internal/model"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "conversorj.db"
	}
	tmpDir := os.Getenv("TMP_DIR")
	if tmpDir == "" {
		tmpDir = "./tmp"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("abrir banco: %v", err)
	}
	if err := db.AutoMigrate(&model.Conversion{}); err != nil {
		log.Fatalf("migrar banco: %v", err)
	}

	conv, err := converter.New(tmpDir)
	if err != nil {
		log.Fatalf("criar converter: %v", err)
	}

	h := handler.New(db, conv)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)
	h.Routes(r)

	log.Println("servidor iniciado em :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
