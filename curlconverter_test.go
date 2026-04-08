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

	if got := SupportedLanguages(); len(got) != 5 {
		t.Fatalf("SupportedLanguages() length = %d, want 5", len(got))
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
