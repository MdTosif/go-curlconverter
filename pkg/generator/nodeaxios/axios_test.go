package nodeaxios

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesSelectedAxiosFixtures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		commandFile  string
		expectedFile string
	}{
		{"strange_http_method.sh", "strange_http_method.js"},
		{"get_with_env_var.sh", "get_with_env_var.js"},
		{"post_with_urlencoded_data.sh", "post_with_urlencoded_data.js"},
		{"multipart_post.sh", "multipart_post.js"},
		{"put_with_file.sh", "put_with_file.js"},
		{"get_proxy_with_auth.sh", "get_proxy_with_auth.js"},
		{"get_basic_auth.sh", "get_basic_auth.js"},
		{"post_json.sh", "post_json.js"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.expectedFile, func(t *testing.T) {
			t.Parallel()
			cmdPath := filepath.Join("..", "..", "..", "test", "fixtures", "curl_commands", tc.commandFile)
			expectedPath := filepath.Join("..", "..", "..", "test", "fixtures", "node-axios", tc.expectedFile)

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
