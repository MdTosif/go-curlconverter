package parser

import "testing"

func TestParseEmptyQuotedDataBinary(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl --data-binary "" http://localhost:28139`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %q", req.Method)
	}
	if !req.HasBody {
		t.Fatal("expected HasBody to be true")
	}
	if req.Body != "" {
		t.Fatalf("expected empty body, got %q", req.Body)
	}
	if len(req.URLs) != 1 || req.URLs[0].URL != "http://localhost:28139" {
		t.Fatalf("unexpected URLs: %#v", req.URLs)
	}
}

func TestParseDataAsciiAndHeaders(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl 'http://localhost:28139/post' -H 'X-Test: yes' --data-ascii 'msg1=wow&msg2=such'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %q", req.Method)
	}
	if !req.HasBody {
		t.Fatal("expected HasBody to be true")
	}
	if req.Body != "msg1=wow&msg2=such" {
		t.Fatalf("unexpected body %q", req.Body)
	}
	if got := req.Headers["X-Test"]; got != "yes" {
		t.Fatalf("unexpected header value %q", got)
	}
	if got := req.Headers["Content-Type"]; got != "application/x-www-form-urlencoded" {
		t.Fatalf("unexpected Content-Type %q", got)
	}
	if len(req.HeaderKV) != 2 ||
		req.HeaderKV[0].Key != "X-Test" ||
		req.HeaderKV[0].Value != "yes" ||
		req.HeaderKV[1].Key != "Content-Type" ||
		req.HeaderKV[1].Value != "application/x-www-form-urlencoded" {
		t.Fatalf("unexpected ordered headers: %#v", req.HeaderKV)
	}
}

func TestParseCookieFlagAddsCookieHeader(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl 'https://example.com' -b 'foo=bar; session=abc'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := req.Headers["Cookie"]; got != "foo=bar; session=abc" {
		t.Fatalf("unexpected cookie header %q", got)
	}
	if len(req.HeaderKV) != 1 || req.HeaderKV[0].Key != "Cookie" || req.HeaderKV[0].Value != "foo=bar; session=abc" {
		t.Fatalf("unexpected ordered headers: %#v", req.HeaderKV)
	}
}

func TestParseCookieOverridesExistingCookieHeader(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl 'https://example.com' -H 'Cookie: old=value' -b 'new=value'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := req.Headers["Cookie"]; got != "new=value" {
		t.Fatalf("unexpected cookie header %q", got)
	}
	if len(req.HeaderKV) != 1 || req.HeaderKV[0].Key != "Cookie" || req.HeaderKV[0].Value != "new=value" {
		t.Fatalf("unexpected ordered headers: %#v", req.HeaderKV)
	}
}

func TestParseGetMergesRepeatedDataIntoURL(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl 'http://localhost:28139/path?test=2' -G -d 'limit=100' -d 'w=4'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "GET" {
		t.Fatalf("expected GET, got %q", req.Method)
	}
	if req.HasBody {
		t.Fatal("expected HasBody to be false")
	}
	if req.Body != "" {
		t.Fatalf("expected empty body, got %q", req.Body)
	}
	if len(req.URLs) != 1 || req.URLs[0].URL != "http://localhost:28139/path?test=2&limit=100&w=4" {
		t.Fatalf("unexpected URLs: %#v", req.URLs)
	}
}

func TestParseJsonBuildsCombinedBodyAndHeaders(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl --json '{ "drink":' --json ' "coffe" }' http://localhost:28139`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %q", req.Method)
	}
	if !req.HasBody || !req.JSONBody {
		t.Fatalf("expected JSON body flags, got HasBody=%v JSONBody=%v", req.HasBody, req.JSONBody)
	}
	if req.Body != `{ "drink": "coffe" }` {
		t.Fatalf("unexpected body %q", req.Body)
	}
	if got := req.Headers["Content-Type"]; got != "application/json" {
		t.Fatalf("unexpected Content-Type %q", got)
	}
	if got := req.Headers["Accept"]; got != "application/json" {
		t.Fatalf("unexpected Accept %q", got)
	}
}

func TestParseGetPreservesExplicitMethod(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl -X POST 'https://example.com/items?existing=1' -G -d 'page=2'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected explicit method to be preserved, got %q", req.Method)
	}
	if req.HasBody {
		t.Fatal("expected HasBody to be false")
	}
	if req.URLs[0].URL != "https://example.com/items?existing=1&page=2" {
		t.Fatalf("unexpected URL %q", req.URLs[0].URL)
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cmd  string
		want string
	}{
		{name: "empty command", cmd: "", want: "empty command"},
		{name: "non curl command", cmd: "echo hi", want: "command must start with 'curl'"},
		{name: "missing request arg", cmd: "curl -X", want: "missing argument for -X/--request"},
		{name: "missing upload file arg", cmd: "curl -T", want: "missing argument for -T/--upload-file"},
		{name: "missing proxy arg", cmd: "curl -x", want: "missing argument for -x/--proxy"},
		{name: "missing proxy user arg", cmd: "curl -U", want: "missing argument for -U/--proxy-user"},
		{name: "missing user agent arg", cmd: "curl -A", want: "missing argument for -A/--user-agent"},
		{name: "missing referer arg", cmd: "curl -e", want: "missing argument for -e/--referer"},
		{name: "missing bearer arg", cmd: "curl --oauth2-bearer", want: "missing argument for --oauth2-bearer"},
		{name: "missing header arg", cmd: "curl -H", want: "missing argument for -H/--header"},
		{name: "missing cookie arg", cmd: "curl -b", want: "missing argument for -b/--cookie"},
		{name: "missing data arg", cmd: "curl --data", want: "missing argument for -d/--data"},
		{name: "missing json arg", cmd: "curl --json", want: "missing argument for --json"},
		{name: "missing user arg", cmd: "curl -u", want: "missing argument for -u/--user"},
		{name: "missing url arg", cmd: "curl --url", want: "missing argument for --url"},
		{name: "missing form arg", cmd: "curl -F", want: "missing argument for -F/--form"},
		{name: "invalid form arg", cmd: "curl -F nope http://localhost", want: "invalid argument for -F/--form"},
		{name: "missing url", cmd: "curl -H 'foo: bar'", want: "no URL found in command"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse(tc.cmd)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.want {
				t.Fatalf("expected error %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestParseBasicAuthAndHeadAndURLFlag(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl -I --url 'https://example.com/items' -u 'alice:secret'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "HEAD" {
		t.Fatalf("expected HEAD, got %q", req.Method)
	}
	if req.BasicAuth != "alice:secret" {
		t.Fatalf("unexpected basic auth %q", req.BasicAuth)
	}
	if len(req.URLs) != 1 || req.URLs[0].URL != "https://example.com/items" {
		t.Fatalf("unexpected URLs: %#v", req.URLs)
	}
}

func TestParseProxyAndProxyUser(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl 'http://localhost:28139' --proxy 'http://localhost:8080' -U 'anonymous:anonymous'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Proxy != "http://localhost:8080" {
		t.Fatalf("unexpected proxy %q", req.Proxy)
	}
	if req.ProxyAuth != "anonymous:anonymous" {
		t.Fatalf("unexpected proxy auth %q", req.ProxyAuth)
	}
}

func TestParseUploadFileSetsPutAndAppendsURLPath(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl http://localhost:28139 --upload-file file.txt`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "PUT" {
		t.Fatalf("expected PUT, got %q", req.Method)
	}
	if req.BodyFile != "file.txt" {
		t.Fatalf("unexpected body file %q", req.BodyFile)
	}
	if len(req.URLs) != 1 || req.URLs[0].URL != "http://localhost:28139/file.txt" {
		t.Fatalf("unexpected URLs: %#v", req.URLs)
	}
}

func TestParseDataBinaryFileUsesFileBody(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl -X POST --data-binary @./sample.sparql http://localhost:28139/query`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %q", req.Method)
	}
	if req.BodyFile != "./sample.sparql" {
		t.Fatalf("unexpected body file %q", req.BodyFile)
	}
	if req.Body != "" {
		t.Fatalf("expected empty inline body, got %q", req.Body)
	}
}

func TestParseDigestAuthFlag(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl "http://localhost:28139/" -u "some_username:some_password" --digest`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !req.DigestAuth {
		t.Fatal("expected DigestAuth to be true")
	}
	if req.BasicAuth != "some_username:some_password" {
		t.Fatalf("unexpected basic auth %q", req.BasicAuth)
	}
}

func TestParseUserAgentRefererAndBearerFlags(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl http://localhost:28139 -A "SimCity" -e "https://website.com" --oauth2-bearer AAAAAAAAAAAA`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := req.Headers["User-Agent"]; got != "SimCity" {
		t.Fatalf("unexpected User-Agent %q", got)
	}
	if got := req.Headers["Referer"]; got != "https://website.com" {
		t.Fatalf("unexpected Referer %q", got)
	}
	if req.BearerToken != "AAAAAAAAAAAA" {
		t.Fatalf("unexpected bearer token %q", req.BearerToken)
	}
}

func TestParseHandlesLineContinuationsAndInlineComments(t *testing.T) {
	t.Parallel()

	req, err := Parse("curl \\\n  'https://example.com/api' \\\n  -H 'X-Test: yes' # ignore me\n  --data-raw 'a=1'")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(req.URLs) != 1 || req.URLs[0].URL != "https://example.com/api" {
		t.Fatalf("unexpected URLs: %#v", req.URLs)
	}
	if got := req.Headers["X-Test"]; got != "yes" {
		t.Fatalf("unexpected header value %q", got)
	}
	if req.Body != "a=1" {
		t.Fatalf("unexpected body %q", req.Body)
	}
}

func TestParseMultipartForms(t *testing.T) {
	t.Parallel()

	req, err := Parse(`curl http://localhost:28139/api -F attributes='{"name":"tigers.jpeg"}' -F file=@myfile.jpg --form-string note='hello'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %q", req.Method)
	}
	if len(req.FormParts) != 3 {
		t.Fatalf("expected 3 form parts, got %#v", req.FormParts)
	}
	if req.FormParts[0].Name != "attributes" || req.FormParts[0].Value != `{"name":"tigers.jpeg"}` || req.FormParts[0].IsFile {
		t.Fatalf("unexpected first form part %#v", req.FormParts[0])
	}
	if !req.FormParts[1].IsFile || req.FormParts[1].FileName != "myfile.jpg" || req.FormParts[1].Name != "file" {
		t.Fatalf("unexpected file form part %#v", req.FormParts[1])
	}
	if req.FormParts[2].Name != "note" || req.FormParts[2].Value != "hello" || req.FormParts[2].IsFile {
		t.Fatalf("unexpected string form part %#v", req.FormParts[2])
	}
}
