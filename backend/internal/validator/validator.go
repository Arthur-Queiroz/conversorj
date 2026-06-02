package validator

import "regexp"

var (
	ytPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^https?://(www\.)?youtube\.com/watch\?.*v=[\w-]+`),
		regexp.MustCompile(`(?i)^https?://youtu\.be/[\w-]+`),
		regexp.MustCompile(`(?i)^https?://(www\.)?youtube\.com/shorts/[\w-]+`),
	}
	xPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^https?://(www\.)?(x\.com|twitter\.com)/\w+/status/\d+`),
	}
)

// Platform returns true if rawURL matches the expected pattern for the given platform.
func Platform(platform, rawURL string) bool {
	switch platform {
	case "youtube":
		for _, p := range ytPatterns {
			if p.MatchString(rawURL) {
				return true
			}
		}
	case "x":
		for _, p := range xPatterns {
			if p.MatchString(rawURL) {
				return true
			}
		}
	}
	return false
}
