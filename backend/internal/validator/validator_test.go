package validator

import "testing"

func TestYouTube(t *testing.T) {
	valid := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://www.youtube.com/shorts/abc123def",
		"https://youtube.com/watch?v=abc&t=10s",
	}
	for _, u := range valid {
		if !Platform("youtube", u) {
			t.Errorf("esperava válida: %s", u)
		}
	}

	invalid := []string{
		"https://vimeo.com/123456",
		"https://youtube.com/channel/UCabc",
		"https://youtube.com/playlist?list=PLabc",
		"not-a-url",
		"",
		"https://x.com/user/status/123",
	}
	for _, u := range invalid {
		if Platform("youtube", u) {
			t.Errorf("esperava inválida: %s", u)
		}
	}
}

func TestX(t *testing.T) {
	valid := []string{
		"https://x.com/natgeo/status/1789452200145",
		"https://twitter.com/user/status/9876543210",
		"https://www.x.com/user/status/1234567890",
	}
	for _, u := range valid {
		if !Platform("x", u) {
			t.Errorf("esperava válida: %s", u)
		}
	}

	invalid := []string{
		"https://x.com/user",
		"https://twitter.com/user",
		"https://youtube.com/watch?v=abc",
		"",
	}
	for _, u := range invalid {
		if Platform("x", u) {
			t.Errorf("esperava inválida: %s", u)
		}
	}
}

func TestUnknownPlatform(t *testing.T) {
	if Platform("instagram", "https://instagram.com/p/abc") {
		t.Error("plataforma desconhecida deveria retornar false")
	}
}
