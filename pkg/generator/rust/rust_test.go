package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesRustFixtures(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "test", "fixtures", "rust"))
	if err != nil {
		t.Fatalf("read rust fixtures: %v", err)
	}

	for _, entry := range entries {
		entry := entry
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".rs" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			base := entry.Name()[:len(entry.Name())-len(".rs")]
			cmdPath := filepath.Join("..", "..", "..", "test", "fixtures", "curl_commands", base+".sh")
			expectedPath := filepath.Join("..", "..", "..", "test", "fixtures", "rust", entry.Name())

			cmd, err := os.ReadFile(cmdPath)
			if err != nil {
				t.Fatalf("read command fixture: %v", err)
			}
			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read expected fixture: %v", err)
			}

			req, err := parser.Parse(string(cmd))
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			actual := Generate(req)
			if actual != string(expected) {
				t.Fatalf("generated output mismatch\nexpected:\n%s\nactual:\n%s", string(expected), actual)
			}
		})
	}
}
