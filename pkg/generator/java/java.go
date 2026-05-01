package java

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

	imports := map[string]bool{
		"java.net.URI":               true,
		"java.net.http.HttpClient":   true,
		"java.net.http.HttpRequest":  true,
		"java.net.http.HttpResponse": true,
		"java.io.IOException":        true,
	}

	var javaCode string
	url := r.URLs[0]

	javaCode += "HttpClient client = "

	clientLines := []string{}
	if r.FollowRedirects {
		clientLines = append(clientLines, "    .followRedirects(HttpClient.Redirect.NORMAL)\n")
	} else if !r.FollowRedirects {
		clientLines = append(clientLines, "    .followRedirects(HttpClient.Redirect.NEVER)\n")
	}
	if r.ConnectTimeout != "" {
		clientLines = append(clientLines, fmt.Sprintf("    .connectTimeout(Duration.ofSeconds(%s))\n", r.ConnectTimeout))
		imports["java.time.Duration"] = true
	}

	if len(clientLines) > 0 {
		javaCode += "HttpClient.newBuilder()\n"
		for _, line := range clientLines {
			javaCode += line
		}
		javaCode += "    .build()"
	} else {
		javaCode += "HttpClient.newHttpClient()"
	}
	javaCode += ";\n"
	javaCode += "\n"

	methodCallArgs := []string{}
	if r.Body != "" {
		methodCallArgs = append(methodCallArgs, "BodyPublishers.ofString("+reprStr(r.Body)+")")
		imports["java.net.http.HttpRequest.BodyPublishers"] = true
	}

	if r.BasicAuth != "" {
		javaCode += "String credentials = " + reprStr(r.BasicAuth) + ";\n"
		javaCode += "String auth = \"Basic \" + Base64.getEncoder().encodeToString(credentials.getBytes());\n\n"
		imports["java.util.Base64"] = true
	}

	javaCode += "HttpRequest request = HttpRequest.newBuilder()\n"
	javaCode += "    .uri(URI.create(" + reprStr(url.URL) + "))\n"

	methods := []string{"DELETE", "GET", "POST", "PUT"}
	dataMethods := []string{"POST", "PUT"}
	method := strings.ToUpper(url.Method)
	methodCall := "method"

	isStandardMethod := false
	for _, m := range methods {
		if m == method {
			isStandardMethod = true
			break
		}
	}

	if !isStandardMethod || (len(methodCallArgs) > 0 && !contains(dataMethods, method)) {
		if len(methodCallArgs) == 0 {
			methodCallArgs = append(methodCallArgs, "HttpRequest.BodyPublishers.noBody()")
		}
		methodCallArgs = append([]string{reprStr(method)}, methodCallArgs...)
	} else {
		if len(methodCallArgs) == 0 && contains(dataMethods, method) {
			methodCallArgs = append(methodCallArgs, "HttpRequest.BodyPublishers.noBody()")
		}
		methodCall = strings.ToLower(method)
	}

	javaCode += "    ." + methodCall + "(" + strings.Join(methodCallArgs, ", ") + ")\n"

	for _, h := range r.HeaderKV {
		if h.Value != "" {
			javaCode += "    .setHeader(" + reprStr(h.Key) + ", " + reprStr(h.Value) + ")\n"
		}
	}
	if r.BasicAuth != "" {
		javaCode += "    .setHeader(\"Authorization\", auth)\n"
	}

	if r.MaxTime != "" {
		javaCode += fmt.Sprintf("    .timeout(Duration.ofSeconds(%s))\n", r.MaxTime)
		imports["java.time.Duration"] = true
	}

	javaCode += "    .build();\n"
	javaCode += "\n"
	javaCode += "HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());\n"

	var preambleCode string
	importList := make([]string, 0, len(imports))
	for imp := range imports {
		importList = append(importList, "import "+imp+";")
	}
	sort.Strings(importList)
	for _, imp := range importList {
		preambleCode += imp + "\n"
	}
	if len(imports) > 0 {
		preambleCode += "\n"
	}

	return preambleCode + javaCode
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
		}
		if len(c) == 0 {
			return ""
		}
		hex := fmt.Sprintf("%X", c[0])
		return "\\u" + hex
	})
	return "\"" + escaped + "\""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
