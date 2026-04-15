package nodegot

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "node-got")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No node-got fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	// TODO: Add fixture tests when node-got fixtures are available
	t.Skip("node-got fixtures not yet implemented")
}

func TestGenerateBasicNodeGot(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "import got from 'got'") {
		t.Error("Expected got import")
	}
	if !strings.Contains(generated, "got('") {
		t.Error("Expected got() call")
	}
}

func TestGenerateNodeGotPost(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "POST",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "got.post('") {
		t.Error("Expected got.post() for POST request")
	}
}

func TestGenerateNodeGotWithAuth(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method:    "GET",
		BasicAuth: "user:pass",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "username: 'user'") {
		t.Error("Expected username in options")
	}
	if !strings.Contains(generated, "password: 'pass'") {
		t.Error("Expected password in options")
	}
}
