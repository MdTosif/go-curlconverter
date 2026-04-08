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

	jsonOutput, err := ParseJSON(`curl -b cookie.txt http://localhost:28139/`)
	if err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}
	if jsonOutput == "" {
		t.Fatal("expected parser JSON output")
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
}
