package javascriptjquery

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

	// Handle data
	dataString := ""

	if len(r.FormParts) > 0 {
		code.WriteString(renderFormData(r.FormParts))
		dataString = "form"
	} else if r.HasBody && r.Body != "" {
		// Check if it's JSON
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			dataString = "JSON.stringify(" + jsJSONValue(parsed, 1) + ")"
		} else {
			dataString = "'" + escapeSingle(r.Body) + "'"
		}
	}

	// Check if we can use simple $.get() or $.post()
	// jQuery supports $.get() and $.post() for simple requests without extra config
	canUseSimple := method == "GET" || method == "POST"
	needsConfig := len(r.HeaderKV) > 0 || r.BasicAuth != "" || r.MaxTime != "" || !canUseSimple

	// Use simple form for GET/POST without extra config
	if !needsConfig && (method == "GET" || method == "POST") {
		fn := "get"
		if method == "POST" {
			fn = "post"
		}
		code.WriteString("$.")
		code.WriteString(fn)
		code.WriteString("('")
		code.WriteString(escapeSingle(u))
		code.WriteString("'")
		if dataString != "" {
			code.WriteString(", ")
			code.WriteString(dataString)
		}
		code.WriteString(")")
	} else {
		// Use $.ajax() for complex requests
		code.WriteString("$.ajax({\n")
		code.WriteString("  url: '")
		code.WriteString(escapeSingle(u))
		code.WriteString("',\n")
		code.WriteString("  crossDomain: true,\n")

		if method != "GET" {
			code.WriteString("  method: '")
			code.WriteString(escapeSingle(method))
			code.WriteString("',\n")
		}

		// Headers
		if len(r.HeaderKV) > 0 {
			code.WriteString("  headers: {\n")
			for _, header := range r.HeaderKV {
				code.WriteString("    '")
				code.WriteString(escapeSingle(header.Key))
				code.WriteString("': '")
				code.WriteString(escapeSingle(header.Value))
				code.WriteString("',\n")
			}
			code.WriteString("  },\n")
		}

		// Basic auth
		if r.BasicAuth != "" {
			user, pass := splitBasicAuth(r.BasicAuth)
			code.WriteString("  username: '")
			code.WriteString(escapeSingle(user))
			code.WriteString("',\n")
			code.WriteString("  password: '")
			code.WriteString(escapeSingle(pass))
			code.WriteString("',\n")
		}

		// Data
		if dataString != "" {
			code.WriteString("  data: ")
			code.WriteString(dataString)
			code.WriteString(",\n")
		}

		// Timeout
		if r.MaxTime != "" {
			code.WriteString("  timeout: ")
			code.WriteString(r.MaxTime)
			code.WriteString(" * 1000,\n")
		}

		code.WriteString("})\n")
	}

	// Add done callback
	code.WriteString("  .done(function(response) {\n")
	code.WriteString("    console.log(response);\n")
	code.WriteString("  });\n")

	return code.String()
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
			lines = append(lines, "form.append('"+escapeSingle(part.Name)+"', new File(['file content'], '"+escapeSingle(part.FileName)+"'));")
		} else {
			lines = append(lines, "form.append('"+escapeSingle(part.Name)+"', '"+escapeSingle(part.Value)+"');")
		}
	}
	return strings.Join(lines, "\n") + "\n\n"
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
