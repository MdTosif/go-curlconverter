package nodehttp

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "node-http")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No node-http fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	// TODO: Add fixture tests when node-http fixtures are available
	t.Skip("node-http fixtures not yet implemented")
}

func TestGenerateBasicNodeHttp(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "http.get('") {
		t.Error("Expected http.get() for simple GET request")
	}
	if !strings.Contains(generated, "import http from 'http'") {
		t.Error("Expected http import")
	}
}

func TestGenerateNodeHttpPost(t *testing.T) {
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
	if !strings.Contains(generated, "http.request(") {
		t.Error("Expected http.request() for POST request")
	}
	if !strings.Contains(generated, "req.write('") {
		t.Error("Expected req.write() for body data")
	}
}

func TestGenerateNodeHttpHttps(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "https://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "import https from 'https'") {
		t.Error("Expected https import for HTTPS URL")
	}
}
