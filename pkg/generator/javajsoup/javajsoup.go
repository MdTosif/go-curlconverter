package javajsoup

import (
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

	imports := []string{
		"java.io.IOException",
		"java.io.File",
		"java.io.FileInputStream",
		"org.jsoup.Jsoup",
		"org.jsoup.Connection",
	}

	var javaCode string

	javaCode += "\nclass Main {\n\n"
	javaCode += "\tpublic static void main(String[] args) throws IOException {\n"

	if r.BasicAuth != "" {
		javaCode += fmt.Sprintf("\t\tbyte[] message = (%s).getBytes(\"UTF-8\");\n", reprStr(r.BasicAuth))
		javaCode += "\t\tString basicAuth = DatatypeConverter.printBase64Binary(message);\n"
		javaCode += "\n"
		imports = append(imports, "javax.xml.bind.DatatypeConverter")
	}

	javaCode += fmt.Sprintf("\t\tConnection.Response response = Jsoup.connect(%s)\n", reprStr(r.URLs[0].URL))

	for _, h := range r.HeaderKV {
		if h.Value == "" {
			continue
		}

		if strings.EqualFold(h.Key, "user-agent") {
			javaCode += fmt.Sprintf("\t\t\t.userAgent(%s)\n", reprStr(h.Value))
		} else if strings.EqualFold(h.Key, "cookie") && len(r.Cookies) > 0 {
			for _, c := range r.Cookies {
				javaCode += fmt.Sprintf("\t\t\t.cookie(%s, %s)\n", reprStr(c[0]), reprStr(c[1]))
			}
		} else {
			javaCode += fmt.Sprintf("\t\t\t.header(%s, %s)\n", reprStr(h.Key), reprStr(h.Value))
		}
	}

	if r.BasicAuth != "" {
		javaCode += "\t\t\t.header(\"Authorization\", \"Basic \" + basicAuth)\n"
	}

	if len(r.FormParts) > 0 {
		javaCode += "\t\t\t.data("
		for _, part := range r.FormParts {
			if part.IsFile {
				javaCode += fmt.Sprintf("%s, \"filename\", new FileInputStream(new File(%s)), ", reprStr(part.Name), reprStr(part.FileName))
			} else {
				javaCode += fmt.Sprintf("%s, \"filename\", new FileInputStream(new File(%s)), ", reprStr(part.Name), reprStr(part.Value))
			}
		}
		javaCode += ")))\n"
	} else if r.Body != "" {
		javaCode += fmt.Sprintf("\t\t\t.requestBody(%s)\n", reprStr(r.Body))
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	javaCode += fmt.Sprintf("\t\t\t.method(org.jsoup.Connection.Method.%s)\n", strings.ToUpper(method))
	javaCode += "\t\t\t.ignoreContentType(true)\n"
	if r.MaxTime != "" {
		javaCode += fmt.Sprintf("\t\t\t.timeout(%s * 1000)\n", r.MaxTime)
	}
	javaCode += "\t\t\t.execute();\n\n"
	javaCode += "\t\tSystem.out.println(response.parse());\n"

	javaCode += "\t}\n"
	javaCode += "}"

	var preambleCode string
	sort.Strings(imports)
	for _, imp := range imports {
		preambleCode += "import " + imp + ";\n"
	}
	if len(imports) > 0 {
		preambleCode += "\n"
	}

	return preambleCode + javaCode + "\n"
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\\":
			return "\\\\"
		case "\b":
			return "\\b"
		case "\f":
			return "\\f"
		case "\n":
			return "\\n"
		case "\r":
			return "\\r"
		case "\t":
			return "\\t"
		case "\"":
			return "\\\""
		default:
			if len(c) == 2 {
				first := c[0]
				second := c[1]
				return fmt.Sprintf("\\u%04X\\u%04X", first, second)
			}
			if c == "\x00" {
				return "\\0"
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "\\u" + hex
		}
	})
	return "\"" + escaped + "\""
}
