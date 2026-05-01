package javaokhttp

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	requestURL := r.URLs[0]

	imports := map[string]bool{
		"java.io.IOException":  true,
		"okhttp3.OkHttpClient": true,
		"okhttp3.Request":      true,
		"okhttp3.Response":     true,
	}

	var javaCode string

	javaCode += "OkHttpClient client = new OkHttpClient()"
	clientLines := []string{}
	if r.MaxTime != "" {
		clientLines = append(clientLines, fmt.Sprintf("    .callTimeout(%s, TimeUnit.SECONDS)\n", r.MaxTime))
		imports["java.util.concurrent.TimeUnit"] = true
	}
	if r.ConnectTimeout != "" {
		clientLines = append(clientLines, fmt.Sprintf("    .connectTimeout(%s, TimeUnit.SECONDS)\n", r.ConnectTimeout))
		imports["java.util.concurrent.TimeUnit"] = true
	}
	if !r.FollowRedirects {
		clientLines = append(clientLines, "    .followRedirects(false)\n")
		clientLines = append(clientLines, "    .followSslRedirects(false)\n")
	}
	if len(clientLines) > 0 {
		javaCode += ".newBuilder()\n"
		for _, line := range clientLines {
			javaCode += line
		}
		javaCode += "    .build()"
	}
	javaCode += ";\n"
	javaCode += "\n"

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		javaCode += fmt.Sprintf("String credential = Credentials.basic(%s, %s);\n\n", reprStr(user), reprStr(pass))
		imports["okhttp3.Credentials"] = true
	}

	methodCallArgs := []string{}
	if r.BodyFile != "" {
		methodCallArgs = append(methodCallArgs, "file")
		javaCode += fmt.Sprintf("File file = new File(%s);\n\n", reprStr(r.BodyFile))
		imports["java.io.File"] = true
	} else if len(r.FormParts) > 0 {
		methodCallArgs = append(methodCallArgs, "requestBody")
		javaCode += "RequestBody requestBody = new MultipartBody.Builder()\n"
		javaCode += "    .setType(MultipartBody.FORM)\n"
		for _, part := range r.FormParts {
			args := []string{reprStr(part.Name)}
			if !part.IsFile {
				args = append(args, reprStr(part.Value))
			} else {
				filename := part.FileName
				if filename != "" && filename != part.FileName {
					args = append(args, reprStr(filename))
				}
				args = append(args, fmt.Sprintf("RequestBody.create(\"\", new File(%s))", reprStr(part.FileName)))
				imports["java.io.File"] = true
			}
			javaCode += "    .addFormDataPart(" + strings.Join(args, ", ") + ")\n"
		}
		javaCode += "    .build();\n\n"
		imports["okhttp3.RequestBody"] = true
		imports["okhttp3.MultipartBody"] = true
	} else if r.Body != "" {
		contentType := getContentType(r.HeaderKV)
		if contentType == "application/x-www-form-urlencoded" {
			values, err := url.ParseQuery(r.Body)
			if err == nil && len(values) > 0 {
				methodCallArgs = append(methodCallArgs, "formBody")
				javaCode += "RequestBody formBody = new FormBody.Builder()\n"
				for name, vs := range values {
					for _, v := range vs {
						javaCode += fmt.Sprintf("    .add(%s, %s)\n", reprStr(name), reprStr(v))
					}
				}
				javaCode += "    .build();\n\n"
				imports["okhttp3.FormBody"] = true
				imports["okhttp3.RequestBody"] = true
			} else {
				methodCallArgs = append(methodCallArgs, "requestBody")
				javaCode += fmt.Sprintf("String requestBody = %s;\n\n", reprStr(r.Body))
			}
		} else {
			methodCallArgs = append(methodCallArgs, "requestBody")
			javaCode += fmt.Sprintf("String requestBody = %s;\n\n", reprStr(r.Body))
		}
	}

	javaCode += "Request request = new Request.Builder()\n"
	javaCode += fmt.Sprintf("    .url(%s)\n", reprStr(requestURL.URL))
	methods := []string{"DELETE", "GET", "HEAD", "PATCH", "POST", "PUT"}
	dataMethods := []string{"DELETE", "PATCH", "POST", "PUT"}
	method := requestURL.Method
	if method == "" {
		method = "GET"
	}
	methodCall := "method"
	if !contains(methods, strings.ToUpper(method)) || (len(methodCallArgs) > 0 && !contains(dataMethods, strings.ToUpper(method))) {
		methodCallArgs = append([]string{reprStr(method)}, methodCallArgs...)
	} else {
		methodCall = strings.ToLower(method)
	}
	if methodCall != "get" {
		javaCode += fmt.Sprintf("    .%s(%s)\n", methodCall, strings.Join(methodCallArgs, ", "))
	}

	for _, h := range r.HeaderKV {
		if h.Value != "" {
			javaCode += fmt.Sprintf("    .header(%s, %s)\n", reprStr(h.Key), reprStr(h.Value))
		}
	}
	if r.BasicAuth != "" {
		javaCode += "    .header(\"Authorization\", credential)\n"
	}

	javaCode += "    .build();\n"

	javaCode += "\n"
	javaCode += "try (Response response = client.newCall(request).execute()) {\n"
	javaCode += "    if (!response.isSuccessful()) throw new IOException(\"Unexpected code \" + response);\n"
	javaCode += "    response.body().string();\n"
	javaCode += "}\n"

	var preambleCode string
	importList := make([]string, 0, len(imports))
	for imp := range imports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)
	for _, imp := range importList {
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
		}

		if len(c) == 2 {
			first := c[0]
			second := c[1]
			return fmt.Sprintf("\\u%04X\\u%04X", first, second)
		}

		if c == "\x00" {
			return "\\0"
		}
		return fmt.Sprintf("\\u%04X", c[0])
	})
	return "\"" + escaped + "\""
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getContentType(headers []request.Header) string {
	for _, h := range headers {
		if strings.ToLower(h.Key) == "content-type" {
			return h.Value
		}
	}
	return ""
}
