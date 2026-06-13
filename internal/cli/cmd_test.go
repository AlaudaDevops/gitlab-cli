package cli

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestParseGitLabHostURL verifies that parseGitLabHostURL handles IPv6 literals,
// IPv4 addresses, and DNS hostnames correctly, including default-port omission and
// IPv6 re-bracketing in the rebuilt endpoint.
func TestParseGitLabHostURL(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantEndpoint string
		wantScheme   string
		wantHost     string
		wantPort     int
	}{
		{
			name:         "IPv6 with explicit port",
			input:        "http://[2335::aa1:1415]:32336",
			wantEndpoint: "http://[2335::aa1:1415]:32336",
			wantScheme:   "http",
			wantHost:     "2335::aa1:1415",
			wantPort:     32336,
		},
		{
			name:         "IPv6 https default port",
			input:        "https://[2001:db8::1]",
			wantEndpoint: "https://[2001:db8::1]",
			wantScheme:   "https",
			wantHost:     "2001:db8::1",
			wantPort:     443,
		},
		{
			name:         "IPv6 http default port",
			input:        "http://[::1]",
			wantEndpoint: "http://[::1]",
			wantScheme:   "http",
			wantHost:     "::1",
			wantPort:     80,
		},
		{
			name:         "IPv4 with explicit port",
			input:        "http://10.161.11.29:32739",
			wantEndpoint: "http://10.161.11.29:32739",
			wantScheme:   "http",
			wantHost:     "10.161.11.29",
			wantPort:     32739,
		},
		{
			name:         "IPv4 hostname https default",
			input:        "https://gitlab.example.com",
			wantEndpoint: "https://gitlab.example.com",
			wantScheme:   "https",
			wantHost:     "gitlab.example.com",
			wantPort:     443,
		},
		{
			name:         "hostname with explicit port",
			input:        "https://gitlab.example.com:8443",
			wantEndpoint: "https://gitlab.example.com:8443",
			wantScheme:   "https",
			wantHost:     "gitlab.example.com",
			wantPort:     8443,
		},
		{
			name:         "IPv6 with port and trailing slash",
			input:        "http://[2335::aa1:1415]:32336/",
			wantEndpoint: "http://[2335::aa1:1415]:32336",
			wantScheme:   "http",
			wantHost:     "2335::aa1:1415",
			wantPort:     32336,
		},
		{
			name:         "no scheme defaults to https",
			input:        "gitlab.example.com",
			wantEndpoint: "https://gitlab.example.com",
			wantScheme:   "https",
			wantHost:     "gitlab.example.com",
			wantPort:     443,
		},
		{
			name:         "leading and trailing whitespace trimmed",
			input:        "  https://gitlab.example.com:8443  ",
			wantEndpoint: "https://gitlab.example.com:8443",
			wantScheme:   "https",
			wantHost:     "gitlab.example.com",
			wantPort:     8443,
		},
		{
			// Issue 1: a bare IPv6 literal with no scheme and no brackets must not have its
			// trailing group mistaken for a port; the whole value is the host.
			name:         "no scheme bare IPv6 literal",
			input:        "2335::aa1:1415",
			wantEndpoint: "https://[2335::aa1:1415]",
			wantScheme:   "https",
			wantHost:     "2335::aa1:1415",
			wantPort:     443,
		},
		{
			// Issue (codex r2): no-scheme bare IPv6 WITH a subpath must keep both host and
			// path; the path must not leak into the IPv6 bracket detection.
			name:         "no scheme bare IPv6 literal with subpath",
			input:        "2335::aa1:1415/gitlab",
			wantEndpoint: "https://[2335::aa1:1415]/gitlab",
			wantScheme:   "https",
			wantHost:     "2335::aa1:1415",
			wantPort:     443,
		},
		{
			// Issue 2: a GitLab instance served under a subpath must keep its path.
			name:         "https hostname with subpath",
			input:        "https://gitlab.example.com/gitlab",
			wantEndpoint: "https://gitlab.example.com/gitlab",
			wantScheme:   "https",
			wantHost:     "gitlab.example.com",
			wantPort:     443,
		},
		{
			// Issue 2: IPv6 with an explicit port and a subpath keeps both bracketed host
			// and path.
			name:         "IPv6 with port and subpath",
			input:        "https://[2001:db8::1]:8443/gl",
			wantEndpoint: "https://[2001:db8::1]:8443/gl",
			wantScheme:   "https",
			wantHost:     "2001:db8::1",
			wantPort:     8443,
		},
		{
			// Issue 3: a malformed URL must fall back without corrupting the endpoint
			// (no doubled scheme, no doubled brackets). Endpoint stays the trimmed input.
			name:         "malformed URL falls back without corruption",
			input:        "http://[::1",
			wantEndpoint: "http://[::1",
			wantScheme:   "http",
			wantHost:     "",
			wantPort:     80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, scheme, host, port := parseGitLabHostURL(tt.input)
			if endpoint != tt.wantEndpoint {
				t.Errorf("endpoint = %q, want %q", endpoint, tt.wantEndpoint)
			}
			if scheme != tt.wantScheme {
				t.Errorf("scheme = %q, want %q", scheme, tt.wantScheme)
			}
			if host != tt.wantHost {
				t.Errorf("host = %q, want %q", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("port = %d, want %d", port, tt.wantPort)
			}
		})
	}
}

// TestGeneratedHostValueIsValidYAML is the regression guard for the original failure:
// an IPv6 host rendered with brackets (e.g. "[2335::aa1:1415]:32336") makes YAML parse
// the value as a flow sequence and fail. The fixed parser returns a bare host, which must
// round-trip through yaml.v3 without error.
func TestGeneratedHostValueIsValidYAML(t *testing.T) {
	endpoint, _, host, _ := parseGitLabHostURL("http://[2335::aa1:1415]:32336")

	// snippet mirrors how testing/config/gitlab-template.yaml renders endpoint and host.
	snippet := "toolchains:\n" +
		"  gitlab:\n" +
		"    endpoint: " + endpoint + "\n" +
		"    host: " + host + "\n"

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(snippet), &parsed); err != nil {
		t.Fatalf("generated YAML snippet failed to parse: %v\nsnippet:\n%s", err, snippet)
	}

	// Confirm the host scalar survived parsing intact (bare IPv6, no brackets).
	gitlab, ok := parsed["toolchains"].(map[string]interface{})["gitlab"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected YAML structure: %#v", parsed)
	}
	if got := gitlab["host"]; got != "2335::aa1:1415" {
		t.Errorf("parsed host = %v, want %q", got, "2335::aa1:1415")
	}
}
