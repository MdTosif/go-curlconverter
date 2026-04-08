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
	if len(req.HeaderKV) != 1 || req.HeaderKV[0].Key != "X-Test" || req.HeaderKV[0].Value != "yes" {
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
		{name: "missing header arg", cmd: "curl -H", want: "missing argument for -H/--header"},
		{name: "missing cookie arg", cmd: "curl -b", want: "missing argument for -b/--cookie"},
		{name: "missing data arg", cmd: "curl --data", want: "missing argument for -d/--data"},
		{name: "missing json arg", cmd: "curl --json", want: "missing argument for --json"},
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
