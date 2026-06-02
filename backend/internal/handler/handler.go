package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"conversorj/internal/converter"
	"conversorj/internal/model"
	"conversorj/internal/validator"
)

const (
	maxConcurrent   = 5
	rateLimit       = 10
	rateLimitWindow = time.Hour
	downloadTTL     = 10 * time.Minute
	convertTimeout  = 5 * time.Minute
)

type convertRequest struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Format   string `json:"format"`
}

type convertResponse struct {
	DownloadURL string `json:"downloadUrl"`
	Filename    string `json:"filename"`
	ExpiresAt   string `json:"expiresAt"`
}

type errResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Handler struct {
	db      *gorm.DB
	conv    *converter.Converter
	sem     chan struct{}
	mu      sync.Mutex
	ipTimes map[string][]time.Time
}

func New(db *gorm.DB, conv *converter.Converter) *Handler {
	h := &Handler{
		db:      db,
		conv:    conv,
		sem:     make(chan struct{}, maxConcurrent),
		ipTimes: make(map[string][]time.Time),
	}
	go h.cleanupLoop()
	return h
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/api/health", h.health)
	r.Post("/api/convert", h.convert)
	r.Get("/api/download/{id}", h.download)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) convert(w http.ResponseWriter, r *http.Request) {
	var req convertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid_request", "JSON inválido.")
		return
	}
	req.Platform = strings.TrimSpace(req.Platform)
	req.URL = strings.TrimSpace(req.URL)
	req.Format = strings.TrimSpace(req.Format)

	if req.Platform != "youtube" && req.Platform != "x" {
		jsonErr(w, http.StatusBadRequest, "invalid_platform", "Plataforma inválida. Use 'youtube' ou 'x'.")
		return
	}
	if req.Format != "mp3" && req.Format != "mp4" {
		jsonErr(w, http.StatusBadRequest, "invalid_format", "Formato inválido. Use 'mp3' ou 'mp4'.")
		return
	}
	if !validator.Platform(req.Platform, req.URL) {
		jsonErr(w, http.StatusBadRequest, "invalid_url",
			fmt.Sprintf("O link informado não é válido para a plataforma %s.", platformLabel(req.Platform)))
		return
	}

	if !h.checkRate(clientIP(r)) {
		jsonErr(w, http.StatusTooManyRequests, "rate_limit_exceeded",
			"Você atingiu o limite de conversões por hora. Tente novamente mais tarde.")
		return
	}

	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	default:
		jsonErr(w, http.StatusTooManyRequests, "server_busy",
			"Muitos vídeos estão sendo processados no momento. Aguarde um instante e tente novamente.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), convertTimeout)
	defer cancel()

	duration, err := h.conv.CheckDuration(ctx, req.URL)
	if err != nil {
		jsonErr(w, http.StatusBadGateway, "extraction_failed",
			"Não foi possível obter informações do vídeo. Verifique o link e tente novamente.")
		return
	}
	if duration > converter.MaxDuration {
		jsonErr(w, http.StatusBadRequest, "duration_exceeded",
			"O vídeo tem mais de 20 minutos e não pode ser convertido.")
		return
	}

	id := randomID()
	result, err := h.conv.Convert(ctx, id, req.URL, req.Format)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "conversion_failed",
			"Erro ao converter o vídeo. Tente novamente.")
		return
	}

	expiresAt := time.Now().Add(downloadTTL)
	h.db.Create(&model.Conversion{
		ID:        id,
		Platform:  req.Platform,
		Format:    req.Format,
		Filename:  result.Filename,
		FilePath:  result.FilePath,
		ExpiresAt: expiresAt,
	})

	jsonOK(w, convertResponse{
		DownloadURL: "/api/download/" + id,
		Filename:    result.Filename,
		ExpiresAt:   expiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var conv model.Conversion
	if err := h.db.First(&conv, "id = ?", id).Error; err != nil {
		jsonErr(w, http.StatusNotFound, "not_found", "Arquivo não encontrado ou expirado.")
		return
	}
	if time.Now().After(conv.ExpiresAt) {
		go h.remove(&conv)
		jsonErr(w, http.StatusGone, "expired", "O link de download expirou.")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, conv.Filename))
	w.Header().Set("Content-Type", mimeType(conv.Format))
	http.ServeFile(w, r, conv.FilePath)

	go h.remove(&conv)
}

func (h *Handler) remove(conv *model.Conversion) {
	os.RemoveAll(filepath.Dir(conv.FilePath))
	h.db.Delete(conv)
}

// checkRate retorna true se o IP ainda está dentro do limite (e registra a requisição).
func (h *Handler) checkRate(ip string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	cutoff := time.Now().Add(-rateLimitWindow)
	prev := h.ipTimes[ip]
	valid := prev[:0]
	for _, t := range prev {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rateLimit {
		h.ipTimes[ip] = valid
		return false
	}
	h.ipTimes[ip] = append(valid, time.Now())
	return true
}

func (h *Handler) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		var expired []model.Conversion
		h.db.Where("expires_at < ?", time.Now()).Find(&expired)
		for i := range expired {
			h.remove(&expired[i])
		}
	}
}

// ---- helpers ----

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}

func platformLabel(p string) string {
	if p == "youtube" {
		return "YouTube"
	}
	return "X (Twitter)"
}

func mimeType(format string) string {
	if format == "mp3" {
		return "audio/mpeg"
	}
	return "video/mp4"
}

func randomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResponse{Error: code, Message: msg})
}
