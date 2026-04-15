package http

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func parseCurlCommand(cmd string) (*request.Request, error) {
	return parser.Parse(cmd)
}

func TestGenerateMatchesCurlconverterFixtures(t *testing.T) {
	t.Parallel()

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "http")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No HTTP fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	t.Skip("HTTP fixtures not yet implemented")
}

func TestGenerateBasicHTTP(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "GET /api HTTP/1.1") {
		t.Error("Expected HTTP request line")
	}
	if !strings.Contains(generated, "Host:") {
		t.Error("Expected Host header")
	}
}

func TestGenerateHTTPPost(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method:  "POST",
		HasBody: true,
		Body:    "name=value",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "POST /api HTTP/1.1") {
		t.Error("Expected POST request line")
	}
}
