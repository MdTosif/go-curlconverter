package phprequests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesPHPRequestsFixtures(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "test", "fixtures", "php-requests"))
	if err != nil {
		t.Fatalf("read php-requests fixtures: %v", err)
	}

	for _, entry := range entries {
		entry := entry
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".php" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			base := entry.Name()[:len(entry.Name())-len(".php")]
			cmdPath := filepath.Join("..", "..", "..", "test", "fixtures", "curl_commands", base+".sh")
			expectedPath := filepath.Join("..", "..", "..", "test", "fixtures", "php-requests", entry.Name())

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
