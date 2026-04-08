package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesPythonFixtures(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "test", "fixtures", "python"))
	if err != nil {
		t.Fatalf("read python fixtures: %v", err)
	}

	for _, entry := range entries {
		entry := entry
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".py" {
			continue
		}
		name := entry.Name()[:len(entry.Name())-len(".py")]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cmdPath := filepath.Join("..", "..", "..", "test", "fixtures", "curl_commands", name+".sh")
			expectedPath := filepath.Join("..", "..", "..", "test", "fixtures", "python", entry.Name())

			cmd, err := os.ReadFile(cmdPath)
			if err != nil {
				t.Fatalf("read command fixture: %v", err)
			}
			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read expected fixture: %v", err)
			}
			actual, err := GenerateCommand(string(cmd))
			if err != nil {
				req, parseErr := parser.Parse(string(cmd))
				if parseErr != nil {
					t.Fatalf("GenerateCommand() error = %v; parse failed: %v", err, parseErr)
				}
				actual = Generate(req)
			}
			if actual != string(expected) {
				t.Fatalf("generated output mismatch\nexpected:\n%s\nactual:\n%s", string(expected), actual)
			}
		})
	}
}
