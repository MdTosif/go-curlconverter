package javahttpurlconnection

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
		"java.io.InputStream",
		"java.net.HttpURLConnection",
		"java.net.URL",
		"java.util.Scanner",
	}

	var javaCode string

	javaCode += fmt.Sprintf("        URL url = new URL(%s);\n", reprStr(r.URLs[0].URL))
	javaCode += "        HttpURLConnection httpConn = (HttpURLConnection) url.openConnection();\n"
	javaCode += fmt.Sprintf("        httpConn.setRequestMethod(%s);\n\n", reprStr(r.URLs[0].Method))

	gzip := false
	for _, h := range r.HeaderKV {
		if h.Value == "" {
			continue
		}
		javaCode += fmt.Sprintf("        httpConn.setRequestProperty(%s, %s);\n", reprStr(h.Key), reprStr(h.Value))
		if strings.EqualFold(h.Key, "accept-encoding") && strings.Contains(h.Value, "gzip") {
			gzip = true
		}
	}
	if len(r.HeaderKV) > 0 {
		javaCode += "\n"
	}

	if r.BasicAuth != "" {
		javaCode += fmt.Sprintf("        byte[] message = (%s).getBytes(\"UTF-8\");\n", reprStr(r.BasicAuth))
		javaCode += "        String basicAuth = DatatypeConverter.printBase64Binary(message);\n"
		javaCode += "        httpConn.setRequestProperty(\"Authorization\", \"Basic \" + basicAuth);\n"
		javaCode += "\n"
		imports = append(imports, "javax.xml.bind.DatatypeConverter")
	}

	if r.Body != "" {
		javaCode += "        httpConn.setDoOutput(true);\n"
		javaCode += "        OutputStreamWriter writer = new OutputStreamWriter(httpConn.getOutputStream());\n"
		javaCode += fmt.Sprintf("        writer.write(%s);\n", reprStr(r.Body))
		javaCode += "        writer.flush();\n"
		javaCode += "        writer.close();\n"
		javaCode += "        httpConn.getOutputStream().close();\n"
		javaCode += "\n"
		imports = append(imports, "java.io.OutputStreamWriter")
	}

	javaCode += "        InputStream responseStream = httpConn.getResponseCode() / 100 == 2\n"
	javaCode += "                ? httpConn.getInputStream()\n"
	javaCode += "                : httpConn.getErrorStream();\n"
	if gzip {
		javaCode += "        if (\"gzip\".equals(httpConn.getContentEncoding())) {\n"
		javaCode += "            responseStream = new GZIPInputStream(responseStream);\n"
		javaCode += "        }\n"
		imports = append(imports, "java.util.zip.GZIPInputStream")
	}
	javaCode += "        Scanner s = new Scanner(responseStream).useDelimiter(\"\\\\A\");\n"
	javaCode += "        String response = s.hasNext() ? s.next() : \"\";\n"
	javaCode += "        System.out.println(response);\n"

	javaCode += "    }\n"
	javaCode += "}"

	var preambleCode string
	sort.Strings(imports)
	for _, imp := range imports {
		preambleCode += "import " + imp + ";\n"
	}
	if len(imports) > 0 {
		preambleCode += "\n"
	}

	preambleCode += "class Main {\n"
	preambleCode += "\n"
	preambleCode += "    public static void main(String[] args) throws IOException {\n"

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
