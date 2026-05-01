package nodefetch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerateMatchesSelectedNodeFixtures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		commandFile  string
		expectedFile string
	}{
		{"get_with_single_header.sh", "get_with_single_header.js"},
		{"post_json.sh", "post_json.js"},
		{"get_with_env_var.sh", "get_with_env_var.js"},
		{"get_proxy.sh", "get_proxy.js"},
		{"head.sh", "head.js"},
		{"strange_http_method.sh", "strange_http_method.js"},
		{"multipart_post.sh", "multipart_post.js"},
		{"j_patch_file_only.sh", "j_patch_file_only.js"},
		{"put_with_T_option.sh", "put_with_T_option.js"},
		{"post_with_urlencoded_data.sh", "post_with_urlencoded_data.js"},
		{"get_with_form.sh", "get_with_form.js"},
		{"get_with_header_without_value.sh", "get_with_header_without_value.js"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.expectedFile, func(t *testing.T) {
			t.Parallel()

			cmdPath := filepath.Join("..", "..", "..", "test", "fixtures", "curl_commands", tc.commandFile)
			expectedPath := filepath.Join("..", "..", "..", "test", "fixtures", "node", tc.expectedFile)

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
