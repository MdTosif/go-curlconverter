package curlconverter

import "testing"

func TestPublicAPI(t *testing.T) {
	t.Parallel()

	reqs, err := Parse(`curl http://localhost:28139 -H 'X-Test: yes'`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(reqs) != 1 || reqs[0].URLs[0].URL != "http://localhost:28139" {
		t.Fatalf("unexpected requests %#v", reqs)
	}

	req, err := ParseRequest(`curl http://localhost:28139 -H 'X-Test: yes'`)
	if err != nil {
		t.Fatalf("ParseRequest() error = %v", err)
	}
	if req.URLs[0].URL != "http://localhost:28139" {
		t.Fatalf("unexpected request %#v", req)
	}

	jsonOutput, err := ParseJSON(`curl -b cookie.txt http://localhost:28139/`)
	if err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}
	if jsonOutput == "" {
		t.Fatal("expected parser JSON output")
	}

	if got := SupportedLanguages(); len(got) != 7 {
		t.Fatalf("SupportedLanguages() length = %d, want 7", len(got))
	}

	code, warnings, err := ToJavaScriptWarn(`curl http://localhost:28139`)
	if err != nil {
		t.Fatalf("ToJavaScriptWarn() error = %v", err)
	}
	if code != "fetch('http://localhost:28139');\n" {
		t.Fatalf("unexpected code %q", code)
	}
	if warnings != nil {
		t.Fatalf("expected nil warnings, got %#v", warnings)
	}

	if code := GenerateJavaScript(req); code == "" {
		t.Fatal("expected GenerateJavaScript() output")
	}
	if code := GenerateNodeAxios(req); code == "" {
		t.Fatal("expected GenerateNodeAxios() output")
	}
	if code := GenerateGo(req); code == "" {
		t.Fatal("expected GenerateGo() output")
	}
	if code := GeneratePython(req); code == "" {
		t.Fatal("expected GeneratePython() output")
	}
}

func TestPublicAPIArgs(t *testing.T) {
	t.Parallel()

	reqs, err := ParseArgs([]string{"curl", "http://localhost:28139", "-H", "X-Test: yes"})
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}
	if len(reqs) != 1 || reqs[0].URLs[0].URL != "http://localhost:28139" {
		t.Fatalf("unexpected requests %#v", reqs)
	}

	req, err := ParseRequestArgs([]string{"curl", "http://localhost:28139", "-d", "name=codex"})
	if err != nil {
		t.Fatalf("ParseRequestArgs() error = %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("unexpected method %q", req.Method)
	}

	jsonOutput, err := ParseJSONArgs([]string{"curl", "http://localhost:28139"})
	if err != nil {
		t.Fatalf("ParseJSONArgs() error = %v", err)
	}
	if jsonOutput == "" {
		t.Fatal("expected parser JSON output")
	}

	if code, err := ToJavaScriptArgs([]string{"curl", "http://localhost:28139"}); err != nil || code == "" {
		t.Fatalf("ToJavaScriptArgs() = %q, %v", code, err)
	}
	if code, err := ToNodeAxiosArgs([]string{"curl", "http://localhost:28139"}); err != nil || code == "" {
		t.Fatalf("ToNodeAxiosArgs() = %q, %v", code, err)
	}
	if code, err := ToGoArgs([]string{"curl", "http://localhost:28139"}); err != nil || code == "" {
		t.Fatalf("ToGoArgs() = %q, %v", code, err)
	}
	if code, err := ToPythonArgs([]string{"curl", "http://localhost:28139"}); err != nil || code == "" {
		t.Fatalf("ToPythonArgs() = %q, %v", code, err)
	}
}

func TestParseWarnAPIsIncludeParserWarnings(t *testing.T) {
	t.Parallel()

	reqs, warnings, err := ParseWarn("curl 'http://localhost:28139")
	if err != nil {
		t.Fatalf("ParseWarn() error = %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected one request, got %d", len(reqs))
	}
	if len(warnings) != 1 || warnings[0][0] != "unterminated-single-quote" {
		t.Fatalf("unexpected parse warnings %#v", warnings)
	}

	jsonOutput, warnings, err := ParseJSONWarn("curl http://localhost:28139 \\")
	if err != nil {
		t.Fatalf("ParseJSONWarn() error = %v", err)
	}
	if jsonOutput == "" {
		t.Fatal("expected parser JSON output")
	}
	if len(warnings) != 1 || warnings[0][0] != "dangling-backslash" {
		t.Fatalf("unexpected JSON warnings %#v", warnings)
	}
}

func TestWarnAPIsReturnCurrentIgnoredPartWarnings(t *testing.T) {
	t.Parallel()

	_, warnings, err := ToJavaScriptWarn(`curl https://first.example --next https://second.example`)
	if err != nil {
		t.Fatalf("ToJavaScriptWarn() error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "next" {
		t.Fatalf("unexpected next warnings %#v", warnings)
	}

	_, warnings, err = ToJavaScriptWarn("curl 'https://example.com")
	if err != nil {
		t.Fatalf("ToJavaScriptWarn() quote error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "unterminated-single-quote" {
		t.Fatalf("unexpected quote warnings %#v", warnings)
	}

	_, warnings, err = ToJavaScriptWarn(`curl "https://example.com/$USER"`)
	if err != nil {
		t.Fatalf("ToJavaScriptWarn() expansion error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "expansion" {
		t.Fatalf("unexpected expansion warnings %#v", warnings)
	}

	_, warnings, err = ToJavaScriptWarn(`curl https://example.com -b cookie.txt`)
	if err != nil {
		t.Fatalf("ToJavaScriptWarn() cookie error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "cookie-files" {
		t.Fatalf("unexpected cookie warnings %#v", warnings)
	}

	_, warnings, err = ToPythonWarn(`curl https://example.com --data-binary @body.txt`)
	if err != nil {
		t.Fatalf("ToPythonWarn() data error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "unsafe-data" {
		t.Fatalf("unexpected data warnings %#v", warnings)
	}

	_, warnings, err = ToGoWarn(`curl https://one.example https://two.example`)
	if err != nil {
		t.Fatalf("ToGoWarn() multiple url error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "multiple-urls" {
		t.Fatalf("unexpected multiple URL warnings %#v", warnings)
	}

	_, warnings, err = ToJavaScriptWarn(`echo hi ; curl https://example.com`)
	if err != nil {
		t.Fatalf("ToJavaScriptWarn() ignored-command error = %v", err)
	}
	if len(warnings) != 1 || warnings[0][0] != "ignored-command" {
		t.Fatalf("unexpected ignored-command warnings %#v", warnings)
	}
}
