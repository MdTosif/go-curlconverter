package javascript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesSelectedCurlconverterFixtures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		commandFile  string
		expectedFile string
	}{
		{
			name:         "get with single header",
			commandFile:  "get_with_single_header.sh",
			expectedFile: "get_with_single_header.js",
		},
		{
			name:         "post with data raw",
			commandFile:  "post_with_data_raw.sh",
			expectedFile: "post_with_data_raw.js",
		},
		{
			name:         "post empty",
			commandFile:  "post_empty.sh",
			expectedFile: "post_empty.js",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmdPath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "curl_commands", tc.commandFile)
			expectedPath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "javascript", tc.expectedFile)

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

func TestGeneratePreservesExplicitContentType(t *testing.T) {
	cmd := `curl -X POST "https://example.com/hello" -H 'Content-Type: application/json' -d '{"name":"world"}'`
	req, err := parser.Parse(cmd)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('https://example.com/hello', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Content-Type': 'application/json'\n" +
		"  },\n" +
		"  body: '{\"name\":\"world\"}'\n" +
		"});\n"

	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateIncludesCookieHeaderFromFlag(t *testing.T) {
	req, err := parser.Parse(`curl 'https://example.com' -b 'foo=bar; session=abc'`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('https://example.com', {\n" +
		"  headers: {\n" +
		"    'Cookie': 'foo=bar; session=abc'\n" +
		"  }\n" +
		"});\n"

	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}
