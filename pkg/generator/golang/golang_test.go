package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesGoFixtures(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "test", "fixtures", "go"))
	if err != nil {
		t.Fatalf("read go fixtures: %v", err)
	}

	for _, entry := range entries {
		entry := entry
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			base := entry.Name()[:len(entry.Name())-len(".go")]
			cmdPath := filepath.Join("..", "..", "..", "test", "fixtures", "curl_commands", base+".sh")
			expectedPath := filepath.Join("..", "..", "..", "test", "fixtures", "go", entry.Name())

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
