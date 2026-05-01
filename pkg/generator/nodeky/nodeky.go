package nodeky

import (
	"encoding/json"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	u := r.URLs[0].URL
	method := r.Method
	if method == "" {
		method = "GET"
	}

	var code strings.Builder

	// Form data handling
	hasForm := len(r.FormParts) > 0
	if hasForm {
		code.WriteString(renderFormData(r.FormParts))
		code.WriteString("\n")
	}

	// Build ky call - ky uses method shortcuts
	methods := map[string]string{"GET": "", "POST": ".post", "PUT": ".put", "PATCH": ".patch", "HEAD": ".head", "DELETE": ".delete"}

	code.WriteString("const response = await ky")
	if m, ok := methods[method]; ok && m != "" {
		code.WriteString(m)
	}

	code.WriteString("('")
	code.WriteString(escapeSingle(u))
	code.WriteString("'")

	// Build options object
	opts := buildOptions(r, method, hasForm)
	if opts != "" {
		code.WriteString(", ")
		code.WriteString(opts)
	}

	code.WriteString(");\n")

	return "import ky from 'ky';\n\n" + code.String()
}

func buildOptions(r *request.Request, method string, hasForm bool) string {
	var opts []string

	// Method (only if not using shortcut)
	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "HEAD" && method != "DELETE" {
		opts = append(opts, "  method: '"+escapeSingle(method)+"'")
	}

	// Headers
	if len(r.HeaderKV) > 0 {
		headerStr := "  headers: {\n"
		for _, h := range r.HeaderKV {
			headerStr += "    '" + escapeSingle(h.Key) + "': '" + escapeSingle(h.Value) + "',\n"
		}
		headerStr += "  }"
		opts = append(opts, headerStr)
	}

	// Auth (ky uses headers for auth)
	if r.BasicAuth != "" {
		opts = append(opts, "  headers: {\n    'Authorization': 'Basic "+escapeSingle(r.BasicAuth)+"'\n  }")
	}

	// Body
	if hasForm {
		opts = append(opts, "  body: form")
	} else if r.HasBody && r.Body != "" {
		bodyStr := getBodyString(r.Body)
		opts = append(opts, "  "+bodyStr)
	}

	// Timeout
	if r.MaxTime != "" {
		opts = append(opts, "  timeout: "+r.MaxTime+"000")
	}

	// Retry
	if r.Retry != "" {
		opts = append(opts, "  retry: { limit: "+r.Retry+" }")
	}

	if len(opts) == 0 {
		return ""
	}

	return "{\n" + strings.Join(opts, ",\n") + "\n}"
}

func getBodyString(body string) string {
	var parsed any
	if err := json.Unmarshal([]byte(body), &parsed); err == nil {
		if _, ok := parsed.(map[string]any); ok {
			return "json: " + jsJSONValue(parsed, 1)
		}
	}
	return "body: '" + escapeSingle(body) + "'"
}

func escapeSingle(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func renderFormData(parts []request.FormPart) string {
	var lines []string
	lines = append(lines, "const form = new FormData();")
	for _, part := range parts {
		if part.IsFile {
			lines = append(lines, "form.append('"+escapeSingle(part.Name)+"', fs.readFileSync('"+escapeSingle(part.FileName)+"'), '"+escapeSingle(part.FileName)+"');")
		} else {
			lines = append(lines, "form.append('"+escapeSingle(part.Name)+"', '"+escapeSingle(part.Value)+"');")
		}
	}
	return strings.Join(lines, "\n")
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
		return ""
	default:
		return "'" + escapeSingle("") + "'"
	}
}
