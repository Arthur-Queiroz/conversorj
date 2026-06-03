package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const MaxDuration = 20 * 60 // 20 minutos em segundos

type Result struct {
	Filename string
	FilePath string
}

type metadata struct {
	Duration float64 `json:"duration"`
}

type Converter struct {
	TmpDir string
}

func New(tmpDir string) (*Converter, error) {
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("criar diretório tmp: %w", err)
	}
	return &Converter{TmpDir: tmpDir}, nil
}

// CheckDuration retorna a duração em segundos sem baixar o vídeo.
func (c *Converter) CheckDuration(ctx context.Context, url string) (float64, error) {
	args := append(c.baseArgs(), "--dump-json", "--no-playlist", "--no-warnings", "--skip-download", url)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("yt-dlp metadata: %w", err)
	}
	var m metadata
	if err := json.Unmarshal(out, &m); err != nil {
		return 0, fmt.Errorf("parse metadata: %w", err)
	}
	return m.Duration, nil
}

// Convert baixa e converte o vídeo, salvando em TmpDir/id/.
func (c *Converter) Convert(ctx context.Context, id, url, format string) (Result, error) {
	dir := filepath.Join(c.TmpDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Result{}, fmt.Errorf("criar dir de saída: %w", err)
	}

	args := append(c.baseArgs(), buildArgs(url, format, dir)...)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		return Result{}, fmt.Errorf("yt-dlp: %w — %s", err, out)
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		os.RemoveAll(dir)
		return Result{}, fmt.Errorf("arquivo de saída não encontrado após conversão")
	}

	filename := entries[0].Name()
	return Result{Filename: filename, FilePath: filepath.Join(dir, filename)}, nil
}

func (c *Converter) baseArgs() []string {
	args := []string{"--js-runtimes", "node"}
	if path := os.Getenv("YTDLP_COOKIES_FILE"); path != "" {
		if info, err := os.Stat(path); err == nil && info.Size() > 0 {
			args = append(args, "--cookies", path)
		}
	}
	return args
}

func buildArgs(url, format, dir string) []string {
	base := []string{"--no-playlist", "--no-warnings", "--no-part", "-P", dir, "-o", "%(title).80B.%(ext)s"}
	if format == "mp3" {
		return append([]string{"-x", "--audio-format", "mp3", "--audio-quality", "0"}, append(base, url)...)
	}
	// mp4
	return append([]string{
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format", "mp4",
	}, append(base, url)...)
}
