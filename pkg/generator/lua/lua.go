package lua

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := []string{"http"}

	var code string
	code += "local body, code, headers, status = http.request"

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	simpleGet := strings.ToUpper(method) == "GET" && r.Body == "" && len(r.HeaderKV) == 0 && r.BasicAuth == ""
	simplePost := strings.ToUpper(method) == "POST" && r.Body != "" && r.BasicAuth == "" && len(r.HeaderKV) == 1 && strings.EqualFold(r.HeaderKV[0].Key, "content-type") && strings.EqualFold(r.HeaderKV[0].Value, "application/x-www-form-urlencoded")

	if simpleGet {
		code += fmt.Sprintf("(%s)\n", reprStr(r.URLs[0].OriginalURL))
	} else if simplePost {
		code += "(\n"
		code += fmt.Sprintf("\t%s\n", reprStr(r.URLs[0].OriginalURL))
		code += fmt.Sprintf("\t%s\n", reprStr(r.Body))
		code += ")\n"
	} else {
		code += "{\n"
		if strings.ToUpper(method) != "GET" {
			code += fmt.Sprintf("\tmethod = %s,\n", reprStr(method))
		}
		code += fmt.Sprintf("\turl = %s,\n", reprStr(r.URLs[0].OriginalURL))

		if r.Body != "" {
			code += "\tsource = ltn12.source.string("
			contentType := getContentType(r)
			if contentType == "application/json" {
				var parsed any
				if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
					code += "json.encode(" + jsonToLua(parsed, 1) + ")"
					imports = append(imports, "json")
				} else {
					code += reprStr(r.Body)
				}
			} else {
				code += reprStr(r.Body)
			}
			code += "),\n"
			imports = append(imports, "ltn12")
		}

		if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
			code += "\theaders = {\n"
			for _, h := range r.HeaderKV {
				if h.Value == "" {
					continue
				}
				code += fmt.Sprintf("\t\t%s = %s,\n", reprKey(h.Key), reprStr(h.Value))
			}
			if r.BasicAuth != "" {
				code += fmt.Sprintf("\t\tauthentication = \"Basic \" .. (mime.b64(%s)),\n", reprStr(r.BasicAuth))
				imports = append(imports, "mime")
			}
			if strings.HasSuffix(code, ",\n") {
				code = code[:len(code)-2] + "\n"
			}
			code += "\t},\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += "}\n"
	}

	var importCode string
	sort.Strings(imports)
	for _, imp := range imports {
		if imp == "http" {
			importCode += "local http = require(\"socket.http\")\n"
		} else {
			importCode += fmt.Sprintf("local %s = require(\"%s\")\n", imp, imp)
		}
	}

	return importCode + "\n" + code
}

func reprStr(s string) string {
	if strings.Contains(s, "\"") && !strings.Contains(s, "'") {
		return pyreprStr(s, "'")
	}
	return pyreprStr(s, "\"")
}

func pyreprStr(s string, quote string) string {
	escaped := strings.Builder{}
	for _, c := range s {
		switch c {
		case '\\':
			escaped.WriteString("\\\\")
		case '\n':
			escaped.WriteString("\\n")
		case '\r':
			escaped.WriteString("\\r")
		case '\t':
			escaped.WriteString("\\t")
		default:
			if string(c) == quote {
				escaped.WriteRune('\\')
			}
			escaped.WriteRune(c)
		}
	}
	return quote + escaped.String() + quote
}

func reprKey(s string) string {
	validKey := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if validKey.MatchString(s) {
		return s
	}
	return "[" + reprStr(s) + "]"
}

func jsonToLua(data any, indent int) string {
	switch v := data.(type) {
	case string:
		return reprStr(v)
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "nil"
	case []any:
		if len(v) == 0 {
			return "{}"
		}
		code := "{\n"
		for _, item := range v {
			code += strings.Repeat("\t", indent+1) + jsonToLua(item, indent+1) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat("\t", indent) + "}"
		return code
	case map[string]any:
		if len(v) == 0 {
			return "{}"
		}
		code := "{\n"
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			code += strings.Repeat("\t", indent+1) + reprKey(key) + " = " + jsonToLua(v[key], indent+1) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat("\t", indent) + "}"
		return code
	default:
		return reprStr(fmt.Sprintf("%v", v))
	}
}

func getContentType(r *request.Request) string {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return strings.TrimSpace(strings.Split(h.Value, ";")[0])
		}
	}
	return ""
}
