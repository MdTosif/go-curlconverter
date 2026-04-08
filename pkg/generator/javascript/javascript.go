package javascript

import (
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

// Generate returns a JavaScript (fetch) snippet for the given Request.
// Minimal implementation: handles first URL, method, headers and body.
func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	u := r.URLs[0].URL
	method := r.Method
	if method == "" {
		method = "GET"
	}

	lines := make([]string, 0, 4)
	if method != "GET" {
		lines = append(lines, "  method: '"+escapeSingle(method)+"'")
	}

	headers := append([]request.Header{}, r.HeaderKV...)
	if r.HasBody && !hasHeader(headers, "Content-Type") {
		headers = append(headers, request.Header{
			Key:   "Content-Type",
			Value: "application/x-www-form-urlencoded",
		})
	}
	if len(headers) > 0 {
		headerLines := make([]string, 0, len(headers))
		for _, header := range headers {
			headerLines = append(
				headerLines,
				"    '"+escapeSingle(header.Key)+"': '"+escapeSingle(header.Value)+"'",
			)
		}
		lines = append(lines, "  headers: {\n"+strings.Join(headerLines, ",\n")+"\n  }")
	}
	if r.HasBody {
		lines = append(lines, "  body: '"+escapeSingle(r.Body)+"'")
	}

	if len(lines) == 0 {
		return "fetch('" + escapeSingle(u) + "');\n"
	}
	return "fetch('" + escapeSingle(u) + "', {\n" + strings.Join(lines, ",\n") + "\n});\n"
}

func escapeSingle(s string) string {
	// very small escape for single quotes and backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func hasHeader(headers []request.Header, name string) bool {
	for _, header := range headers {
		if strings.EqualFold(header.Key, name) {
			return true
		}
	}
	return false
}
