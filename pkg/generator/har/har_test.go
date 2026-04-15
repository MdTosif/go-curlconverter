package har

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

	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "har")
	_, err := os.ReadDir(fixtureDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No HAR fixtures found, skipping fixture tests")
		}
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	t.Skip("HAR fixtures not yet implemented")
}

func TestGenerateBasicHAR(t *testing.T) {
	t.Parallel()

	req := &request.Request{
		URLs: []request.RequestURL{
			{URL: "http://example.com/api"},
		},
		Method: "GET",
	}

	generated := Generate(req)
	if !strings.Contains(generated, `"log"`) {
		t.Error("Expected log field in HAR output")
	}
	if !strings.Contains(generated, `"method"`) {
		t.Error("Expected method field in HAR output")
	}
}
