package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesSelectedPythonFixtures(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"get_basic_auth",
		"get_with_env_var",
		"post_with_urlencoded_data",
		"multipart_post",
		"post_json",
		"strange_http_method",
		"get_with_browser_headers",
		"post_empty",
		"post_with_data_raw",
	}

	for _, name := range testCases {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cmdPath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "curl_commands", name+".sh")
			expectedPath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "python", name+".py")

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
