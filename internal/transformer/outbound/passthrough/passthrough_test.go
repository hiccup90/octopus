package passthrough

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/bestruirui/octopus/internal/transformer/model"
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

func TestEnsureOpenAIStreamIncludeUsage(t *testing.T) {
	streamTrue := true
	req := &model.InternalLLMRequest{
		Model:       "gpt-x",
		Stream:      &streamTrue,
		RequestPath: "/v1/chat/completions",
		RawRequest:  []byte(`{"model":"old","stream":true,"messages":[{"role":"user","content":"hi"}]}`),
	}
	body, err := buildRequestBody(req)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["model"] != "gpt-x" {
		t.Fatalf("model=%v", raw["model"])
	}
	opts, ok := raw["stream_options"].(map[string]any)
	if !ok {
		t.Fatalf("missing stream_options: %s", string(body))
	}
	if opts["include_usage"] != true {
		t.Fatalf("include_usage=%v", opts["include_usage"])
	}
}

func TestEnsureOpenAIStreamIncludeUsage_SkipAnthropic(t *testing.T) {
	streamTrue := true
	req := &model.InternalLLMRequest{
		Model:       "claude-x",
		Stream:      &streamTrue,
		RequestPath: "/v1/messages",
		RawRequest:  []byte(`{"model":"old","stream":true,"messages":[{"role":"user","content":"hi"}]}`),
	}
	body, err := buildRequestBody(req)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["stream_options"]; ok {
		t.Fatalf("should not inject stream_options for anthropic path: %s", string(body))
	}
}

func TestEnsureOpenAIStreamIncludeUsage_SkipEmptyPath(t *testing.T) {
	streamTrue := true
	req := &model.InternalLLMRequest{
		Model:       "gpt-x",
		Stream:      &streamTrue,
		RequestPath: "",
		RawRequest:  []byte(`{"model":"old","stream":true,"messages":[{"role":"user","content":"hi"}]}`),
	}
	body, err := buildRequestBody(req)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["stream_options"]; ok {
		t.Fatalf("should not inject stream_options for empty path: %s", string(body))
	}
}

func TestEnsureOpenAIStreamIncludeUsage_SkipResponses(t *testing.T) {
	streamTrue := true
	req := &model.InternalLLMRequest{
		Model:       "gpt-x",
		Stream:      &streamTrue,
		RequestPath: "/v1/responses",
		RawRequest:  []byte(`{"model":"old","stream":true}`),
	}
	body, err := buildRequestBody(req)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["stream_options"]; ok {
		t.Fatalf("should not inject stream_options for responses path: %s", string(body))
	}
}
