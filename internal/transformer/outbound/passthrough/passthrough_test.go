package passthrough

import (
	"net/url"
	"testing"
)

func TestBuildURL(t *testing.T) {
	cases := []struct {
		name       string
		base       string
		clientPath string
		wantPath   string
	}{
		{"base_with_v1", "https://example.com/v1", "/v1/chat/completions", "/v1/chat/completions"},
		{"base_root", "https://example.com", "/v1/messages", "/v1/messages"},
		{"base_with_trailing_slash", "https://example.com/v1/", "/v1/responses", "/v1/responses"},
		{"base_custom_prefix", "https://example.com/proxy/v1", "/v1/chat/completions", "/proxy/v1/chat/completions"},
		{"base_no_v1_prefix_match", "https://example.com/api", "/v1/chat/completions", "/api/v1/chat/completions"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := buildURL(tc.base, tc.clientPath, url.Values{})
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if u.Path != tc.wantPath {
				t.Fatalf("path got %q want %q", u.Path, tc.wantPath)
			}
		})
	}
}

func TestIsAnthropicPath(t *testing.T) {
	if !isAnthropicPath("/v1/messages") {
		t.Fatal("expected true")
	}
	if isAnthropicPath("/v1/chat/completions") {
		t.Fatal("expected false")
	}
}
