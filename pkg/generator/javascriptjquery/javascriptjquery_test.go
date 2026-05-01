package javascriptjquery

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "javascript-jquery")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No jQuery fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	// TODO: Add fixture tests when jQuery fixtures are available
	t.Skip("jQuery fixtures not yet implemented")
}

func TestGenerateBasicJQuery(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, "$.get('") {
		t.Error("Expected $.get() for simple GET request")
	}
	if !strings.Contains(generated, ".done(function(response)") {
		t.Error("Expected .done() callback")
	}
}

func TestGenerateJQueryPost(t *testing.T) {
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
	if !strings.Contains(generated, "$.post('") {
		t.Error("Expected $.post() for simple POST request")
	}
}

func TestGenerateJQueryAjax(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "PUT",
		HeaderKV: []request.Header{
			{Key: "X-Custom", Value: "header"},
		},
	}

	generated := Generate(req)
	if !strings.Contains(generated, "$.ajax({") {
		t.Error("Expected $.ajax() for PUT request with headers")
	}
	if !strings.Contains(generated, "method: 'PUT'") {
		t.Error("Expected method option in ajax call")
	}
}
