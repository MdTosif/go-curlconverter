package nodeky

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "node-ky")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No node-ky fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	t.Skip("node-ky fixtures not yet implemented")
}

func TestGenerateBasicNodeKy(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "import ky from 'ky'") {
		t.Error("Expected ky import")
	}
	if !strings.Contains(generated, "ky('") {
		t.Error("Expected ky() call")
	}
}

func TestGenerateNodeKyPost(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "POST",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "ky.post('") {
		t.Error("Expected ky.post() for POST request")
	}
}

func TestGenerateNodeKyPut(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "PUT",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "ky.put('") {
		t.Error("Expected ky.put() for PUT request")
	}
}
