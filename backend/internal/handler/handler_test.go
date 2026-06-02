package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"conversorj/internal/converter"
	"conversorj/internal/model"
)

func setupTestHandler(t *testing.T) (*Handler, *chi.Mux) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir banco de teste: %v", err)
	}
	db.AutoMigrate(&model.Conversion{})

	conv, _ := converter.New(t.TempDir())
	h := New(db, conv)

	r := chi.NewRouter()
	h.Routes(r)
	return h, r
}

func post(t *testing.T, r http.Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/convert", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	_, r := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, got %d", w.Code)
	}
}

func TestConvert_InvalidJSON(t *testing.T) {
	_, r := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/convert", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, got %d", w.Code)
	}
}

func TestConvert_InvalidPlatform(t *testing.T) {
	_, r := setupTestHandler(t)
	w := post(t, r, map[string]string{"platform": "instagram", "url": "https://x.com/u/status/1", "format": "mp3"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, got %d", w.Code)
	}
	var resp errResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "invalid_platform" {
		t.Errorf("esperava invalid_platform, got %q", resp.Error)
	}
}

func TestConvert_InvalidFormat(t *testing.T) {
	_, r := setupTestHandler(t)
	w := post(t, r, map[string]string{"platform": "youtube", "url": "https://youtube.com/watch?v=abc", "format": "ogg"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, got %d", w.Code)
	}
	var resp errResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "invalid_format" {
		t.Errorf("esperava invalid_format, got %q", resp.Error)
	}
}

func TestConvert_InvalidURL(t *testing.T) {
	_, r := setupTestHandler(t)
	w := post(t, r, map[string]string{"platform": "youtube", "url": "https://vimeo.com/123", "format": "mp3"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, got %d", w.Code)
	}
	var resp errResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "invalid_url" {
		t.Errorf("esperava invalid_url, got %q", resp.Error)
	}
}

func TestConvert_RateLimit(t *testing.T) {
	_, r := setupTestHandler(t)
	payload := map[string]string{
		"platform": "youtube",
		"url":      "https://youtube.com/watch?v=dQw4w9WgXcQ",
		"format":   "mp3",
	}
	// Esgotar o rate limit (chega até a tentativa de chamar yt-dlp, que falha com 502).
	// O que nos importa é que após 10 tentativas válidas o 11º retorna 429.
	var last *httptest.ResponseRecorder
	for i := 0; i < 11; i++ {
		last = post(t, r, payload)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Errorf("esperava 429 no 11º request, got %d", last.Code)
	}
	var resp errResponse
	json.NewDecoder(last.Body).Decode(&resp)
	if resp.Error != "rate_limit_exceeded" {
		t.Errorf("esperava rate_limit_exceeded, got %q", resp.Error)
	}
}

func TestDownload_NotFound(t *testing.T) {
	_, r := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/download/naoexiste", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("esperava 404, got %d", w.Code)
	}
}
