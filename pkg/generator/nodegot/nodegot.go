package nodegot

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

	// Build got call
	code.WriteString("const response = await got")

	// Use shortcut methods for common HTTP methods
	methods := map[string]string{"POST": "post", "PUT": "put", "PATCH": "patch", "HEAD": "head", "DELETE": "delete"}
	if m, ok := methods[method]; ok {
		code.WriteString(".")
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

	return "import got from 'got';\n\n" + code.String()
}

func buildOptions(r *request.Request, method string, hasForm bool) string {
	var opts []string

	// Method
	if method != "GET" && method != "POST" {
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

	// Auth
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		opts = append(opts, "  username: '"+escapeSingle(user)+"'")
		if pass != "" {
			opts = append(opts, "  password: '"+escapeSingle(pass)+"'")
		}
	}

	// Body/form
	if hasForm {
		opts = append(opts, "  body: form")
	} else if r.HasBody && r.Body != "" {
		bodyStr := getBodyString(r.Body)
		opts = append(opts, "  "+bodyStr)
	}

	// Timeout
	if r.MaxTime != "" {
		opts = append(opts, "  timeout: { request: "+r.MaxTime+" * 1000 }")
	}
	if r.ConnectTimeout != "" {
		opts = append(opts, "  timeout: { connect: "+r.ConnectTimeout+" * 1000 }")
	}

	// Redirects (got follows by default, curl doesn't)
	if !r.FollowRedirects {
		opts = append(opts, "  followRedirect: false")
	}
	if r.MaxRedirects != "" {
		opts = append(opts, "  maxRedirects: "+r.MaxRedirects)
	}

	// Insecure
	if r.Insecure {
		opts = append(opts, "  https: { rejectUnauthorized: false }")
	}

	// HTTP2
	if r.HTTP2 {
		opts = append(opts, "  http2: true")
	}

	if len(opts) == 0 {
		return ""
	}

	return "{\n" + strings.Join(opts, ",\n") + "\n}"
}

func getBodyString(body string) string {
	// Check if it's JSON
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

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
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
