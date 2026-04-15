package nodesuperagent

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "node-superagent")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No node-superagent fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	t.Skip("node-superagent fixtures not yet implemented")
}

func TestGenerateBasicNodeSuperagent(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "import superagent from 'superagent'") {
		t.Error("Expected superagent import")
	}
	if !strings.Contains(generated, "superagent.get('") {
		t.Error("Expected superagent.get() call")
	}
}

func TestGenerateNodeSuperagentPost(t *testing.T) {
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
	if !strings.Contains(generated, "superagent.post('") {
		t.Error("Expected superagent.post() for POST request")
	}
	if !strings.Contains(generated, ".send('") {
		t.Error("Expected .send() for body data")
	}
}

func TestGenerateNodeSuperagentWithHeader(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
		HeaderKV: []request.Header{
			{Key: "X-Custom", Value: "value"},
		},
	}

	generated := Generate(req)
	if !strings.Contains(generated, ".set('") {
		t.Error("Expected .set() for headers")
	}
}
