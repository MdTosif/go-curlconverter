package golang

import (
	"path"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

const ifErr = "\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n"

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := map[string]bool{
		"fmt":      true,
		"io":       true,
		"log":      true,
		"net/http": true,
	}

	vars := map[string]string{}
	var body strings.Builder

	if len(r.FormParts) > 0 {
		imports["bytes"] = true
		imports["mime/multipart"] = true
		body.WriteString("\tform := new(bytes.Buffer)\n")
		body.WriteString("\twriter := multipart.NewWriter(form)\n")
		firstFile := true
		firstField := true
		for _, part := range r.FormParts {
			if part.IsFile {
				op := "="
				if firstFile {
					op = ":="
					firstFile = false
				}
				imports["os"] = true
				imports["path/filepath"] = true
				body.WriteString("\tfw, err " + op + " writer.CreateFormFile(" + reprString(part.Name) + ", filepath.Base(" + reprWord(part.FileName, vars, imports) + "))\n")
				body.WriteString(ifErr)
				body.WriteString("\tfd, err " + op + " os.Open(" + reprWord(part.FileName, vars, imports) + ")\n")
				body.WriteString(ifErr)
				body.WriteString("\tdefer fd.Close()\n")
				body.WriteString("\t_, err = io.Copy(fw, fd)\n")
				body.WriteString(ifErr)
			} else {
				op := "="
				if firstField {
					op = ":="
					firstField = false
				}
				body.WriteString("\tformField, err " + op + " writer.CreateFormField(" + reprString(part.Name) + ")\n")
				body.WriteString(ifErr)
				body.WriteString("\t_, err = formField.Write([]byte(" + reprMaybeBacktick(part.Value, vars, imports) + "))\n")
			}
			body.WriteString("\n")
		}
		body.WriteString("\twriter.Close()\n\n")
	}

	body.WriteString("\tclient := &http.Client{}\n")

	hasData := r.Body != "" || r.HasBody && r.Body == ""
	if r.BodyFile == "" && len(r.FormParts) == 0 && hasData {
		imports["strings"] = true
		body.WriteString("\tvar data = strings.NewReader(" + reprBacktick(r.Body, vars, imports) + ")\n")
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	targetURL := r.URLs[0].URL
	if targetURL == "" {
		targetURL = r.URLs[0].OriginalURL
	}

	bodySource := "nil"
	if len(r.FormParts) > 0 {
		bodySource = "form"
	} else if r.Body != "" || (r.HasBody && r.Body == "" && r.BodyFile == "") {
		bodySource = "data"
	}
	body.WriteString("\treq, err := http.NewRequest(" + reprString(method) + ", " + reprWord(targetURL, vars, imports) + ", " + bodySource + ")\n")
	body.WriteString(ifErr)

	for _, header := range r.HeaderKV {
		if len(r.FormParts) > 0 && strings.EqualFold(header.Key, "Content-Type") {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(header.Value, ";")[0]))
			if mediaType == "multipart/form-data" {
				continue
			}
		}
		prefix := "\t"
		if strings.EqualFold(header.Key, "Accept-Encoding") {
			prefix = "\t// "
		}
		body.WriteString(prefix + "req.Header.Set(" + reprString(header.Key) + ", " + reprMaybeBacktick(header.Value, vars, imports) + ")\n")
	}
	if len(r.FormParts) > 0 {
		body.WriteString("\treq.Header.Set(\"Content-Type\", writer.FormDataContentType())\n")
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		body.WriteString("\treq.SetBasicAuth(" + reprWord(user, vars, imports) + ", " + reprWord(pass, vars, imports) + ")\n")
	}

	body.WriteString("\tresp, err := client.Do(req)\n")
	body.WriteString(ifErr)
	body.WriteString("\tdefer resp.Body.Close()\n")
	body.WriteString("\tbodyText, err := io.ReadAll(resp.Body)\n")
	body.WriteString(ifErr)
	body.WriteString("\tfmt.Printf(\"%s\\n\", bodyText)\n")

	var out strings.Builder
	out.WriteString("package main\n\n")
	out.WriteString("import (\n")
	importList := make([]string, 0, len(imports))
	for imp := range imports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)
	for _, imp := range importList {
		out.WriteString("\t\"" + imp + "\"\n")
	}
	out.WriteString(")\n\n")
	out.WriteString("func main() {\n")
	varNames := make([]string, 0, len(vars))
	for name := range vars {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames)
	for _, name := range varNames {
		out.WriteString("\t" + name + ", err := " + vars[name] + "\n")
		out.WriteString(ifErr)
	}
	if len(varNames) > 0 {
		out.WriteString("\n")
	}
	out.WriteString(body.String())
	out.WriteString("}\n")
	return out.String()
}

func reprMaybeBacktick(s string, vars map[string]string, imports map[string]bool) string {
	if strings.Contains(s, "\"") && !strings.Contains(s, "`") && !strings.Contains(s, "\r") && !containsEnvInterpolation(s) {
		return "`" + s + "`"
	}
	return reprWord(s, vars, imports)
}

func reprBacktick(s string, vars map[string]string, imports map[string]bool) string {
	if !strings.Contains(s, "`") && !strings.Contains(s, "\r") && !containsEnvInterpolation(s) {
		return "`" + s + "`"
	}
	return reprWord(s, vars, imports)
}

func reprWord(s string, vars map[string]string, imports map[string]bool) string {
	parts := splitEnvInterpolations(s)
	if len(parts) == 1 {
		return reprString(parts[0])
	}
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "$") && len(part) > 1 {
			imports["os"] = true
			rendered = append(rendered, "os.Getenv("+reprString(part[1:])+")")
		} else {
			rendered = append(rendered, reprString(part))
		}
	}
	return strings.Join(rendered, " + ")
}

func reprString(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", "\\n", "\t", "\\t", "\r", "\\r")
	return "\"" + replacer.Replace(s) + "\""
}

func containsEnvInterpolation(s string) bool {
	parts := splitEnvInterpolations(s)
	return len(parts) > 1
}

func splitEnvInterpolations(s string) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '$' && i+1 < len(s) && isEnvStart(s[i+1]) {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
			j := i + 2
			for j < len(s) && isEnvPart(s[j]) {
				j++
			}
			parts = append(parts, s[i:j])
			i = j - 1
			continue
		}
		cur.WriteByte(ch)
	}
	if len(parts) == 0 {
		return []string{s}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

func isEnvStart(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_'
}

func isEnvPart(ch byte) bool {
	return isEnvStart(ch) || (ch >= '0' && ch <= '9')
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func baseName(file string) string {
	return path.Base(file)
}
