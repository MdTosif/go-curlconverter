package nodeaxios

import (
	"strings"
	"testing"

	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		curl     string
		contains string
	}{
		{
			name:     "simple GET",
			curl:     "curl http://example.com",
			contains: "axios",
		},
		{
			name:     "POST with data",
			curl:     "curl -X POST -d 'key=value' http://example.com",
			contains: "axios.post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs, err := parser.ParseAll(tt.curl)
			if err != nil {
				t.Fatalf("ParseAll failed: %v", err)
			}
			if len(reqs) == 0 {
				t.Fatal("no requests parsed")
			}
			output := Generate(reqs[0])
			if !strings.Contains(output, tt.contains) {
				t.Errorf("output does not contain %q", tt.contains)
			}
		})
	}
}
