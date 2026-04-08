package javascript

import (
	"encoding/json"
	"fmt"
	"sort"
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
		if r.JSONBody {
			lines = append(lines, renderJSONBody(r.Body)...)
		} else {
			lines = append(lines, "  body: '"+escapeSingle(r.Body)+"'")
		}
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

func renderJSONBody(raw string) []string {
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return []string{"  body: '" + escapeSingle(raw) + "'"}
	}

	return []string{
		"  // body: '" + escapeSingle(raw) + "'",
		"  body: JSON.stringify(" + jsJSONValue(parsed, 1) + ")",
	}
}

func jsJSONValue(v any, indentLevel int) string {
	indent := strings.Repeat("  ", indentLevel)
	nextIndent := strings.Repeat("  ", indentLevel+1)

	switch value := v.(type) {
	case map[string]any:
		if len(value) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines := make([]string, 0, len(keys))
		for _, key := range keys {
			lines = append(lines, nextIndent+"'"+escapeSingle(key)+"': "+jsJSONValue(value[key], indentLevel+1))
		}
		return "{\n" + strings.Join(lines, ",\n") + "\n" + indent + "}"
	case []any:
		if len(value) == 0 {
			return "[]"
		}
		lines := make([]string, 0, len(value))
		for _, item := range value {
			lines = append(lines, nextIndent+jsJSONValue(item, indentLevel+1))
		}
		return "[\n" + strings.Join(lines, ",\n") + "\n" + indent + "]"
	case string:
		return "'" + escapeSingle(value) + "'"
	case bool:
		if value {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	case float64:
		return fmt.Sprintf("%v", value)
	default:
		return "'" + escapeSingle(fmt.Sprintf("%v", value)) + "'"
	}
}
