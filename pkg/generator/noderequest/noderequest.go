package noderequest

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

	// Build request() call
	code.WriteString("request('")
	code.WriteString(escapeSingle(u))
	code.WriteString("'")

	// Build options object
	opts := buildOptions(r, method)
	if opts != "" {
		code.WriteString(", ")
		code.WriteString(opts)
	}

	code.WriteString(", function (error, response, body) {\n")
	code.WriteString("  if (error) {\n")
	code.WriteString("    console.error(error);\n")
	code.WriteString("  } else {\n")
	code.WriteString("    console.log(body);\n")
	code.WriteString("  }\n")
	code.WriteString("});\n")

	return "import request from 'request';\n\n" + code.String()
}

func buildOptions(r *request.Request, method string) string {
	var opts []string

	// Method
	if method != "GET" {
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
		authStr := "  auth: {\n"
		authStr += "    'user': '" + escapeSingle(user) + "',\n"
		authStr += "    'pass': '" + escapeSingle(pass) + "'\n"
		authStr += "  }"
		opts = append(opts, authStr)
	}

	// Body
	if r.HasBody && r.Body != "" {
		bodyStr := getBodyString(r.Body)
		opts = append(opts, "  "+bodyStr)
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
			return "body: " + jsJSONValue(parsed, 1)
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
