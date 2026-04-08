package javascript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
	"github.com/mdtosif/go-curlconverter/pkg/request"
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
		{
			name:         "get with data",
			commandFile:  "get_with_data.sh",
			expectedFile: "get_with_data.js",
		},
		{
			name:         "multiple d post",
			commandFile:  "multiple_d_post.sh",
			expectedFile: "multiple_d_post.js",
		},
		{
			name:         "get basic auth",
			commandFile:  "get_basic_auth.sh",
			expectedFile: "get_basic_auth.js",
		},
		{
			name:         "head with I option",
			commandFile:  "head_with_I_option.sh",
			expectedFile: "head_with_I_option.js",
		},
		{
			name:         "get proxy",
			commandFile:  "get_proxy.sh",
			expectedFile: "get_proxy.js",
		},
		{
			name:         "get digest auth",
			commandFile:  "get_digest_auth.sh",
			expectedFile: "get_digest_auth.js",
		},
		{
			name:         "get referer",
			commandFile:  "get_referer.sh",
			expectedFile: "get_referer.js",
		},
		{
			name:         "get with form",
			commandFile:  "get_with_form.sh",
			expectedFile: "get_with_form.js",
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

func TestGenerateJsonBodyMatchesFixtureStyle(t *testing.T) {
	req, err := parser.Parse(`curl --json '{ "drink":' --json ' "coffe" }' http://localhost:28139`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('http://localhost:28139', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Content-Type': 'application/json',\n" +
		"    'Accept': 'application/json'\n" +
		"  },\n" +
		"  // body: '{ \"drink\": \"coffe\" }',\n" +
		"  body: JSON.stringify({\n" +
		"    'drink': 'coffe'\n" +
		"  })\n" +
		"});\n"

	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateSimpleGetWithoutHeadersOrBody(t *testing.T) {
	req, err := parser.Parse(`curl 'https://example.com'`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('https://example.com');\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateAddsDefaultFormContentType(t *testing.T) {
	req, err := parser.Parse(`curl 'https://example.com/form' --data 'a=1&b=2'`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('https://example.com/form', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Content-Type': 'application/x-www-form-urlencoded'\n" +
		"  },\n" +
		"  body: 'a=1&b=2'\n" +
		"});\n"

	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateInvalidJSONFallsBackToRawStringBody(t *testing.T) {
	req, err := parser.Parse(`curl --json '{ "drink":' http://localhost:28139`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('http://localhost:28139', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Content-Type': 'application/json',\n" +
		"    'Accept': 'application/json'\n" +
		"  },\n" +
		"  body: '{ \"drink\":'\n" +
		"});\n"

	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateReturnsEmptyStringForNilOrMissingURL(t *testing.T) {
	if got := Generate(nil); got != "" {
		t.Fatalf("expected empty string for nil request, got %q", got)
	}

	got := Generate(&request.Request{})
	if got != "" {
		t.Fatalf("expected empty string for request without URL, got %q", got)
	}
}

func TestGenerateUsesURLSearchParamsForSimpleFormBodies(t *testing.T) {
	req, err := parser.Parse(`curl 'http://localhost:28139/echo/html/' -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' --data 'msg1=wow&msg2=such'`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('http://localhost:28139/echo/html/', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'\n" +
		"  },\n" +
		"  body: new URLSearchParams({\n" +
		"    'msg1': 'wow',\n" +
		"    'msg2': 'such'\n" +
		"  })\n" +
		"});\n"

	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateIgnoresProxyFlagsForBrowserFetchParity(t *testing.T) {
	req, err := parser.Parse(`curl 'http://localhost:28139' --proxy 'http://localhost:8080' -U 'anonymous:anonymous'`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('http://localhost:28139');\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateUsesFileValueForBinaryFileBody(t *testing.T) {
	req, err := parser.Parse(`curl -X POST --data-binary @./sample.sparql -H 'Content-type: application/sparql-query' http://localhost:28139/american-art/query`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "import { readFile } from 'node:fs/promises';\n\n" +
		"fetch('http://localhost:28139/american-art/query', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Content-type': 'application/sparql-query'\n" +
		"  },\n" +
		"  body: await readFile('./sample.sparql')\n" +
		"});\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateUploadFileReadsFromDisk(t *testing.T) {
	req, err := parser.Parse(`curl http://localhost:28139 --upload-file file.txt`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "import { readFile } from 'node:fs/promises';\n\n" +
		"fetch('http://localhost:28139/file.txt', {\n" +
		"  method: 'PUT',\n" +
		"  body: await readFile('file.txt')\n" +
		"});\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateMultipartFileReadsFromDisk(t *testing.T) {
	req, err := parser.Parse(`curl http://localhost:28139/api/2.0/files/content -H "Authorization: Bearer ACCESS_TOKEN" -X POST -F attributes='{"name":"tigers.jpeg", "parent":{"id":"11446498"}}' -F file=@myfile.jpg`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "import { readFile } from 'node:fs/promises';\n\n" +
		"const form = new FormData();\n" +
		"form.append('attributes', '{\"name\":\"tigers.jpeg\", \"parent\":{\"id\":\"11446498\"}}');\n" +
		"form.append('file', new File([await readFile('myfile.jpg')], 'myfile.jpg'));\n\n" +
		"fetch('http://localhost:28139/api/2.0/files/content', {\n" +
		"  method: 'POST',\n" +
		"  headers: {\n" +
		"    'Authorization': 'Bearer ACCESS_TOKEN'\n" +
		"  },\n" +
		"  body: form\n" +
		"});\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateBasicAuthWithoutUserMatchesFixtureShape(t *testing.T) {
	req, err := parser.Parse(`curl "http://localhost:28139/" -u ":some_password"`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('http://localhost:28139/', {\n" +
		"  headers: {\n" +
		"    'Authorization': 'Basic ' + btoa(':some_password')\n" +
		"  }\n" +
		"});\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}

func TestGenerateBearerTokenAsAuthorizationHeader(t *testing.T) {
	req, err := parser.Parse(`curl http://localhost:28139 --oauth2-bearer AAAAAAAAAAAA`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	code := Generate(req)
	expected := "fetch('http://localhost:28139', {\n" +
		"  headers: {\n" +
		"    'Authorization': 'Bearer AAAAAAAAAAAA'\n" +
		"  }\n" +
		"});\n"
	if code != expected {
		t.Fatalf("unexpected generated code\nexpected:\n%s\nactual:\n%s", expected, code)
	}
}
