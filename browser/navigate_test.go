package browser

import "testing"

func TestCheckURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"http", "http://example.com", false},
		{"https", "https://example.com", false},
		{"https with path", "https://example.com/path?q=1", false},
		{"about:blank", "about:blank", false},
		{"HTTP uppercase", "HTTP://example.com", false},
		{"HTTPS mixed case", "HtTpS://example.com", false},
		{"file URL", "file:///etc/passwd", true},
		{"file URL Windows", "file:///C:/Users/secret.txt", true},
		{"chrome URL", "chrome://settings", true},
		{"chrome extension", "chrome-extension://abc123/page.html", true},
		{"devtools", "devtools://devtools/bundled/inspector.html", true},
		{"javascript", "javascript:alert(1)", true},
		{"data URL", "data:text/html,<h1>hi</h1>", true},
		{"ftp", "ftp://example.com/file", true},
		{"no scheme", "example.com", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkURL(%q) error = %v, wantErr = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
