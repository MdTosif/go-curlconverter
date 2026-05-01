package noderequest

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "node-request")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No node-request fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	t.Skip("node-request fixtures not yet implemented")
}

func TestGenerateBasicNodeRequest(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "import request from 'request'") {
		t.Error("Expected request import")
	}
	if !strings.Contains(generated, "request('") {
		t.Error("Expected request() call")
	}
}

func TestGenerateNodeRequestPost(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method:  "POST",
		HasBody: true,
		Body:    "data=test",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "method: 'POST'") {
		t.Error("Expected method option in POST request")
	}
	if !strings.Contains(generated, "body:") {
		t.Error("Expected body in POST request")
	}
}
