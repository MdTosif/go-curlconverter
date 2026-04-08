package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMatchesCheckedInParserFixtures(t *testing.T) {
	t.Parallel()

	fixturesDir := filepath.Join("..", "..", "test", "fixtures")
	parserDir := filepath.Join(fixturesDir, "parser")
	commandDir := filepath.Join(fixturesDir, "curl_commands")

	entries, err := os.ReadDir(parserDir)
	if err != nil {
		t.Fatalf("read parser fixtures: %v", err)
	}

	for _, entry := range entries {
		entry := entry
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			expectedPath := filepath.Join(parserDir, entry.Name())
			commandPath := filepath.Join(commandDir, entry.Name()[:len(entry.Name())-len(".json")]+".sh")

			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read expected fixture: %v", err)
			}
			command, err := os.ReadFile(commandPath)
			if err != nil {
				t.Fatalf("read command fixture: %v", err)
			}

			reqs, err := ParseAll(string(command))
			if err != nil {
				t.Fatalf("ParseAll() error = %v", err)
			}

			actual, err := MarshalJSON(reqs)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			if actual != string(expected) {
				t.Fatalf("fixture mismatch\nexpected:\n%s\nactual:\n%s", string(expected), actual)
			}
		})
	}
}
