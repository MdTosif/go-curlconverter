package javascriptxhr

import (
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
	nonDataMethods := map[string]bool{"GET": true, "HEAD": true}
	hasData := r.HasBody || len(r.FormParts) > 0

	if len(r.FormParts) > 0 {
		code.WriteString(renderFormData(r.FormParts))
		code.WriteString("\n")
	} else if r.HasBody && r.Body != "" {
		code.WriteString("const data = '")
		code.WriteString(escapeSingle(r.Body))
		code.WriteString("';\n\n")
	}

	// Warn about data with GET/HEAD
	if nonDataMethods[method] && hasData {
		// XHR doesn't send data with GET or HEAD requests
		// This would be a warning in the full implementation
	}

	// Create XHR
	code.WriteString("let xhr = new XMLHttpRequest();\n")
	code.WriteString("xhr.withCredentials = true;\n")

	// Open connection
	code.WriteString("xhr.open('")
	code.WriteString(escapeSingle(method))
	code.WriteString("', '")
	code.WriteString(escapeSingle(u))
	code.WriteString("')")

	// Add auth if present
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		code.WriteString(", true, '")
		code.WriteString(escapeSingle(user))
		code.WriteString("', '")
		code.WriteString(escapeSingle(pass))
		code.WriteString("'")
	}
	code.WriteString(";\n")

	// Set headers
	for _, header := range r.HeaderKV {
		code.WriteString("xhr.setRequestHeader('")
		code.WriteString(escapeSingle(header.Key))
		code.WriteString("', '")
		code.WriteString(escapeSingle(header.Value))
		code.WriteString("');\n")
	}

	// Set timeout
	if r.MaxTime != "" {
		// Convert seconds to milliseconds
		code.WriteString("xhr.timeout = ")
		code.WriteString(r.MaxTime)
		code.WriteString(" * 1000;\n")
	}

	// Add onload handler
	code.WriteString("\nxhr.onload = function() {\n")
	code.WriteString("  console.log(xhr.response);\n")
	code.WriteString("};\n\n")

	// Send data
	if len(r.FormParts) > 0 {
		code.WriteString("xhr.send(form);\n")
	} else if r.HasBody && r.Body != "" {
		code.WriteString("xhr.send(data);\n")
	} else {
		code.WriteString("xhr.send();\n")
	}

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
	return strings.Join(lines, "\n")
}
