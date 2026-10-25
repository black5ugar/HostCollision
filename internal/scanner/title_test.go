package scanner

import "testing"

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"plain", "<html><head><title>Example</title></head></html>", "Example"},
		{"case and attributes", `<TITLE data-note="a > b">Example</TITLE >`, "Example"},
		{"entities and unicode", "<title>用户中心 &amp; &#34;设置&#34; &#x1F680;</title>", "用户中心 & \"设置\" 🚀"},
		{"whitespace", "<title> \n Welcome\t to\r\n Example&nbsp; </title>", "Welcome to Example"},
		{"controls", "<title>Hello\x1b[31m\x07 World</title>", "Hello [31m World"},
		{"missing", "<html><body>No page title</body></html>", ""},
		{"empty body", "", ""},
		{"empty title", "<title> \n </title><title>Second</title>", ""},
		{"first title", "<title>First</title><title>Second</title>", "First"},
		{"comments", "<!-- <title>Fake</title> --><title>Real</title>", "Real"},
		{"script", `<script>const title = "<title>Fake</title>";</script><title>Real</title>`, "Real"},
		{"literal markup", "<title>A <b>B</b></title>", "A <b>B</b>"},
		{"unclosed title", "<title>Truncated", "Truncated"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTitle([]byte(tt.body)); got != tt.want {
				t.Fatalf("extractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}
