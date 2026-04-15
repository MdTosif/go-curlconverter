package nodesuperagent

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

	// Build superagent chain
	code.WriteString("superagent")

	// Method call
	methodLower := strings.ToLower(method)
	code.WriteString(".")
	code.WriteString(methodLower)

	code.WriteString("('")
	code.WriteString(escapeSingle(u))
	code.WriteString("')")

	// Add .set() for headers
	for _, h := range r.HeaderKV {
		code.WriteString("\n  .set('")
		code.WriteString(escapeSingle(h.Key))
		code.WriteString("', '")
		code.WriteString(escapeSingle(h.Value))
		code.WriteString("')")
	}

	// Add .auth() for basic auth
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		code.WriteString("\n  .auth('")
		code.WriteString(escapeSingle(user))
		code.WriteString("', '")
		code.WriteString(escapeSingle(pass))
		code.WriteString("')")
	}

	// Add .send() for body
	if r.HasBody && r.Body != "" {
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			if _, ok := parsed.(map[string]any); ok {
				code.WriteString("\n  .send(")
				code.WriteString(jsJSONValue(parsed, 1))
				code.WriteString(")")
			} else {
				code.WriteString("\n  .send('")
				code.WriteString(escapeSingle(r.Body))
				code.WriteString("')")
			}
		} else {
			code.WriteString("\n  .send('")
			code.WriteString(escapeSingle(r.Body))
			code.WriteString("')")
		}
	}

	// Add .then() callback
	code.WriteString("\n  .then(res => {\n")
	code.WriteString("    console.log(res.body);\n")
	code.WriteString("  })\n")
	code.WriteString("  .catch(err => {\n")
	code.WriteString("    console.error(err);\n")
	code.WriteString("  });\n")

	return "import superagent from 'superagent';\n\n" + code.String()
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
