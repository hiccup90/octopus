package relay

import (
	"testing"
)

func TestBuildImagesUpstreamURL(t *testing.T) {
	cases := []struct {
		name        string
		base        string
		endpoint    string
		clientPath  string
		passthrough bool
		wantPath    string
	}{
		{
			name:        "chat_style",
			base:        "https://example.com/v1",
			endpoint:    "/images/generations",
			clientPath:  "/v1/images/generations",
			passthrough: false,
			wantPath:    "/v1/images/generations",
		},
		{
			name:        "passthrough_follow_client",
			base:        "https://example.com/v1",
			endpoint:    "/images/generations",
			clientPath:  "/v1/images/generations",
			passthrough: true,
			wantPath:    "/v1/images/generations",
		},
		{
			name:        "passthrough_base_root",
			base:        "https://example.com",
			endpoint:    "/images/edits",
			clientPath:  "/v1/images/edits",
			passthrough: true,
			wantPath:    "/v1/images/edits",
		},
		{
			name:        "passthrough_custom_prefix",
			base:        "https://example.com/proxy/v1",
			endpoint:    "/images/generations",
			clientPath:  "/v1/images/generations",
			passthrough: true,
			wantPath:    "/proxy/v1/images/generations",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := buildImagesUpstreamURL(tc.base, tc.endpoint, tc.clientPath, tc.passthrough)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if u.Path != tc.wantPath {
				t.Fatalf("path got %q want %q", u.Path, tc.wantPath)
			}
		})
	}
}
