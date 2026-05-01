package java

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		curl     string
		contains string
	}{
		{
			name:     "simple GET",
			curl:     "curl http://example.com",
			contains: "HttpClient",
		},
		{
			name:     "POST with data",
			curl:     "curl -X POST -d 'key=value' http://example.com",
			contains: "HttpRequest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs, err := parser.ParseAll(tt.curl)
			if err != nil {
				t.Fatalf("ParseAll failed: %v", err)
			}
			if len(reqs) == 0 {
				t.Fatal("no requests parsed")
			}
			output := Generate(reqs[0])
			if !strings.Contains(output, tt.contains) {
				t.Errorf("output does not contain %q", tt.contains)
			}
		})
	}
}

func TestFixtures(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("cannot get current file path")
	}
	fixturesDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "test", "fixtures")
	commandDir := filepath.Join(fixturesDir, "curl_commands")
	javaDir := filepath.Join(fixturesDir, "java")

	entries, err := os.ReadDir(javaDir)
	if err != nil {
		t.Skipf("cannot read fixtures directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".java" {
			continue
		}
		base := entry.Name()[:len(entry.Name())-len(".java")]
		t.Run(base, func(t *testing.T) {
			cmdPath := filepath.Join(commandDir, base+".sh")
			javaPath := filepath.Join(javaDir, entry.Name())

			cmd, err := os.ReadFile(cmdPath)
			if err != nil {
				t.Fatalf("cannot read command file: %v", err)
			}
			expected, err := os.ReadFile(javaPath)
			if err != nil {
				t.Fatalf("cannot read expected file: %v", err)
			}

			reqs, err := parser.ParseAll(string(cmd))
			if err != nil {
				t.Fatalf("ParseAll failed: %v", err)
			}
			if len(reqs) == 0 {
				t.Fatal("no requests parsed")
			}

			output := Generate(reqs[0])
			if output != string(expected) {
				t.Errorf("output mismatch\nGot:\n%s\n\nExpected:\n%s", output, expected)
			}
		})
	}
}
