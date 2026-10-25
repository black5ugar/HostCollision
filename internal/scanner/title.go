package scanner

import (
	"bytes"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// extractTitle reads the first HTML title from the existing response body.
// Tokenization handles case, attributes, entities, and title-like text inside
// comments or scripts without changing the body used for similarity checks.
func extractTitle(body []byte) string {
	z := html.NewTokenizer(bytes.NewReader(body))
	var title strings.Builder
	inTitle := false
	for {
		switch z.Next() {
		case html.ErrorToken:
			return normalizeTitle(title.String())
		case html.StartTagToken:
			name, _ := z.TagName()
			if bytes.Equal(name, []byte("title")) {
				inTitle = true
			}
		case html.TextToken:
			if inTitle {
				title.Write(z.Text())
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			if inTitle && bytes.Equal(name, []byte("title")) {
				return normalizeTitle(title.String())
			}
		}
	}
}

func normalizeTitle(title string) string {
	// Keep terminal and CSV output on one line and remove control characters.
	title = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, title)
	return strings.Join(strings.Fields(title), " ")
}
