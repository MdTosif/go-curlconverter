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
