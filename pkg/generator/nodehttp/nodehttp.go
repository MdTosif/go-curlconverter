package nodehttp

import (
	"encoding/json"
	"net/url"
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

	parsedURL, err := url.Parse(u)
	if err != nil {
		return ""
	}

	module := "http"
	if parsedURL.Scheme == "https" {
		module = "https"
	}

	var code strings.Builder
	var imports []string
	imports = append(imports, "import "+module+" from '"+module+"';")

	hasForm := len(r.FormParts) > 0
	hasData := r.HasBody && r.Body != ""
	dataString := ""
	formString := ""

	if hasForm {
		formString = renderFormData(r.FormParts)
		imports = append(imports, "import FormData from 'form-data';")
		code.WriteString(formString)
		code.WriteString("\n")
	} else if hasData {
		// Check if it's JSON
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			dataString = "JSON.stringify(" + jsJSONValue(parsed, 1) + ")"
		} else {
			dataString = "'" + escapeSingle(r.Body) + "'"
		}
	}

	// Build options
	needsOptions := method != "GET" || len(r.HeaderKV) > 0 || r.BasicAuth != "" || r.MaxTime != ""

	fn := "get"
	if method != "GET" || hasData || hasForm {
		fn = "request"
	}

	if needsOptions {
		code.WriteString("const options = {\n")
		code.WriteString("  hostname: '")
		code.WriteString(escapeSingle(parsedURL.Hostname()))
		code.WriteString("',\n")

		if parsedURL.Port() != "" {
			code.WriteString("  port: ")
			code.WriteString(parsedURL.Port())
			code.WriteString(",\n")
		}

		if parsedURL.Path != "" || parsedURL.RawQuery != "" {
			path := parsedURL.Path
			if parsedURL.RawQuery != "" {
				path = path + "?" + parsedURL.RawQuery
			}
			code.WriteString("  path: '")
			code.WriteString(escapeSingle(path))
			code.WriteString("',\n")
		}

		if method != "GET" {
			code.WriteString("  method: '")
			code.WriteString(escapeSingle(method))
			code.WriteString("',\n")
		}

		// Headers
		if len(r.HeaderKV) > 0 || hasForm {
			code.WriteString("  headers: {\n")
			for _, header := range r.HeaderKV {
				code.WriteString("    '")
				code.WriteString(escapeSingle(header.Key))
				code.WriteString("': '")
				code.WriteString(escapeSingle(header.Value))
				code.WriteString("',\n")
			}
			if hasForm {
				code.WriteString("    ...form.getHeaders(),\n")
			}
			code.WriteString("  },\n")
		} else if hasForm {
			code.WriteString("  headers: form.getHeaders(),\n")
		}

		// Auth
		if r.BasicAuth != "" {
			code.WriteString("  auth: '")
			code.WriteString(escapeSingle(r.BasicAuth))
			code.WriteString("',\n")
		}

		// Timeout
		if r.MaxTime != "" {
			code.WriteString("  timeout: ")
			code.WriteString(r.MaxTime)
			code.WriteString(" * 1000,\n")
		}

		code.WriteString("};\n\n")

		code.WriteString("const req = ")
		code.WriteString(module)
		code.WriteString(".")
		code.WriteString(fn)
		code.WriteString("(options, function (res) {\n")
	} else {
		code.WriteString("const req = ")
		code.WriteString(module)
		code.WriteString(".")
		code.WriteString(fn)
		code.WriteString("('")
		code.WriteString(escapeSingle(u))
		code.WriteString("', function (res) {\n")
	}

	code.WriteString("  const chunks = [];\n\n")
	code.WriteString("  res.on('data', function (chunk) {\n")
	code.WriteString("    chunks.push(chunk);\n")
	code.WriteString("  });\n\n")
	code.WriteString("  res.on('end', function () {\n")
	code.WriteString("    const body = Buffer.concat(chunks);\n")
	code.WriteString("    console.log(body.toString());\n")
	code.WriteString("  });\n")
	code.WriteString("});\n")

	if hasData {
		code.WriteString("\nreq.write(")
		code.WriteString(dataString)
		code.WriteString(");\n")
	} else if hasForm {
		code.WriteString("\nform.pipe(req);\n")
	}

	if fn != "get" && !hasForm {
		code.WriteString("req.end();\n")
	}

	result := strings.Join(imports, "\n") + "\n\n" + code.String()
	return result
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
