package passthrough

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bestruirui/octopus/internal/transformer/model"
)

// Outbound forwards client requests with minimal rewriting:
// - path follows the client request path
// - model is rewritten to the group-mapped upstream model
// - body format is otherwise left intact
type Outbound struct{}

func (o *Outbound) TransformRequest(ctx context.Context, request *model.InternalLLMRequest, baseUrl, key string) (*http.Request, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	body, err := buildRequestBody(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if request.Stream != nil && *request.Stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	// Support both OpenAI-style and Anthropic-style auth headers.
	// Channel custom headers (applied later) can still override these.
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("X-API-Key", key)
	if isAnthropicPath(request.RequestPath) {
		req.Header.Set("Anthropic-Version", "2023-06-01")
	}

	parsedURL, err := buildURL(baseUrl, request.RequestPath, request.Query)
	if err != nil {
		return nil, err
	}
	req.URL = parsedURL
	req.Method = http.MethodPost
	return req, nil
}

func (o *Outbound) TransformResponse(ctx context.Context, response *http.Response) (*model.InternalLLMResponse, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		// Keep pipeline alive for non-OpenAI-shaped JSON: store raw-ish minimal response.
		return &model.InternalLLMResponse{
			Object: "passthrough",
			Model:  "",
		}, nil
	}
	return &resp, nil
}

func (o *Outbound) TransformStream(ctx context.Context, eventData []byte) (*model.InternalLLMResponse, error) {
	if bytes.HasPrefix(eventData, []byte("[DONE]")) {
		return &model.InternalLLMResponse{Object: "[DONE]"}, nil
	}

	var errCheck struct {
		Error *model.ErrorDetail `json:"error"`
	}
	if err := json.Unmarshal(eventData, &errCheck); err == nil && errCheck.Error != nil {
		return nil, &model.ResponseError{Detail: *errCheck.Error}
	}

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(eventData, &resp); err != nil {
		// Unknown stream chunk shape: skip rather than fail the whole stream.
		return nil, nil
	}
	return &resp, nil
}

func buildRequestBody(request *model.InternalLLMRequest) ([]byte, error) {
	if len(request.RawRequest) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(request.RawRequest, &raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw request: %w", err)
		}
		raw["model"] = request.Model
		// OpenAI streaming often omits final usage unless include_usage is set.
		// Mirror Plus ChatOutbound: force it so logs can capture prompt/completion tokens.
		// Skip Anthropic /messages bodies (different protocol).
		ensureOpenAIStreamIncludeUsage(raw, request)
		body, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal passthrough request: %w", err)
		}
		return body, nil
	}

	// Fallback when raw body is unavailable.
	request.ClearHelpFields()
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	return body, nil
}

// ensureOpenAIStreamIncludeUsage sets stream_options.include_usage=true only for
// explicit OpenAI Chat Completions streaming paths. Unknown/empty paths are left
// untouched to avoid surprising non-chat gateways.
func ensureOpenAIStreamIncludeUsage(raw map[string]any, request *model.InternalLLMRequest) {
	if raw == nil || request == nil {
		return
	}

	path := strings.ToLower(request.RequestPath)
	if path == "" || !strings.Contains(path, "chat/completions") {
		return
	}

	streaming := false
	if request.Stream != nil && *request.Stream {
		streaming = true
	}
	if v, ok := raw["stream"]; ok {
		switch t := v.(type) {
		case bool:
			streaming = t
		case string:
			streaming = strings.EqualFold(t, "true") || t == "1"
		}
	}
	if !streaming {
		return
	}

	opts, _ := raw["stream_options"].(map[string]any)
	if opts == nil {
		opts = map[string]any{}
	}
	opts["include_usage"] = true
	raw["stream_options"] = opts
}

func isAnthropicPath(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "/messages")
}

func buildURL(baseURL, requestPath string, query url.Values) (*url.URL, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return nil, fmt.Errorf("base url is empty")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}

	clientPath := strings.TrimSpace(requestPath)
	if clientPath == "" {
		clientPath = "/"
	}
	if !strings.HasPrefix(clientPath, "/") {
		clientPath = "/" + clientPath
	}

	// Prefer appending the client path relative to base path.
	// Examples:
	//   base=https://x.com/v1 + /v1/chat/completions -> https://x.com/v1/chat/completions
	//   base=https://x.com + /v1/messages            -> https://x.com/v1/messages
	basePath := strings.TrimRight(parsed.Path, "/")
	suffix := clientPath
	if basePath != "" {
		if clientPath == basePath {
			suffix = ""
		} else if strings.HasPrefix(clientPath, basePath+"/") {
			suffix = strings.TrimPrefix(clientPath, basePath)
		} else if strings.HasPrefix(clientPath, "/v1/") && (basePath == "/v1" || strings.HasSuffix(basePath, "/v1")) {
			suffix = strings.TrimPrefix(clientPath, "/v1")
		}
	}

	parsed.Path = basePath + suffix
	if len(query) > 0 {
		parsed.RawQuery = query.Encode()
	}
	return parsed, nil
}
