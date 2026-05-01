package json

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "json")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No JSON fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	t.Skip("JSON fixtures not yet implemented")
}

func TestGenerateBasicJSON(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, `"url"`) {
		t.Error("Expected url field in JSON output")
	}
	if !strings.Contains(generated, `"method"`) {
		t.Error("Expected method field in JSON output")
	}
}

func TestGenerateJSONWithHeaders(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
		HeaderKV: []request.Header{
			{Key: "X-Custom", Value: "test"},
		},
	}

	generated := Generate(req)
	if !strings.Contains(generated, `"headers"`) {
		t.Error("Expected headers field in JSON output")
	}
}
