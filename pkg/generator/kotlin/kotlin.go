package kotlin

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

	imports := []string{
		"java.io.IOException",
		"okhttp3.OkHttpClient",
		"okhttp3.Request",
	}

	var kotlinCode string

	kotlinCode += "val client = OkHttpClient()"
	clientLines := []string{}

	if r.MaxTime != "" {
		clientLines = append(clientLines, fmt.Sprintf("  .callTimeout(%s, TimeUnit.SECONDS)\n", r.MaxTime))
		imports = append(imports, "java.util.concurrent.TimeUnit")
	}
	if r.ConnectTimeout != "" {
		clientLines = append(clientLines, fmt.Sprintf("  .connectTimeout(%s, TimeUnit.SECONDS)\n", r.ConnectTimeout))
		imports = append(imports, "java.util.concurrent.TimeUnit")
	}
	if !r.FollowRedirects {
		clientLines = append(clientLines, "  .followRedirects(false)\n")
		clientLines = append(clientLines, "  .followSslRedirects(false)\n")
	}

	if len(clientLines) > 0 {
		kotlinCode += ".newBuilder()\n"
		for _, line := range clientLines {
			kotlinCode += line
		}
		kotlinCode += "  .build()"
	}
	kotlinCode += "\n\n"

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		kotlinCode += fmt.Sprintf("val credential = Credentials.basic(%s, %s);\n\n", reprStr(user), reprStr(pass))
		imports = append(imports, "okhttp3.Credentials")
	}

	methodCallArgs := []string{}
	contentType := getContentType(r)
	exactContentType := getExactContentType(r)

	if r.BodyFile != "" {
		if exactContentType != "" {
			kotlinCode += fmt.Sprintf("val MEDIA_TYPE = %s.toMediaType()\n\n", reprStr(exactContentType))
			imports = append(imports, "okhttp3.MediaType.Companion.toMediaType")
		}
		methodCallArgs = append(methodCallArgs, "file.asRequestBody(MEDIA_TYPE)")
		kotlinCode += fmt.Sprintf("val file = File(%s)\n\n", reprStr(r.BodyFile))
		imports = append(imports, "java.io.File")
		imports = append(imports, "okhttp3.RequestBody.Companion.asRequestBody")
	} else if len(r.FormParts) > 0 {
		methodCallArgs = append(methodCallArgs, "requestBody")
		kotlinCode += "val requestBody = MultipartBody.Builder()\n"
		kotlinCode += "  .setType(MultipartBody.FORM)\n"
		for _, part := range r.FormParts {
			args := []string{reprStr(part.Name)}
			if part.IsFile {
				args = append(args, reprStr(part.FileName))
				args = append(args, fmt.Sprintf("File(%s).asRequestBody()", reprStr(part.FileName)))
				imports = append(imports, "java.io.File")
				imports = append(imports, "okhttp3.RequestBody.Companion.asRequestBody")
			} else {
				args = append(args, reprStr(part.Value))
			}
			kotlinCode += "  .addFormDataPart(" + strings.Join(args, ", ") + ")\n"
		}
		kotlinCode += "  .build()\n\n"
		imports = append(imports, "okhttp3.MultipartBody")
	} else if r.Body != "" {
		if contentType == "application/x-www-form-urlencoded" {
			values, err := url.ParseQuery(r.Body)
			if err == nil && len(values) > 0 {
				methodCallArgs = append(methodCallArgs, "formBody")
				kotlinCode += "val formBody = FormBody.Builder()\n"
				for k, vs := range values {
					for _, v := range vs {
						kotlinCode += fmt.Sprintf("  .add(%s, %s)\n", reprStr(k), reprStr(v))
					}
				}
				kotlinCode += "  .build()\n\n"
				imports = append(imports, "okhttp3.FormBody")
			} else {
				if exactContentType != "" {
					kotlinCode += fmt.Sprintf("val MEDIA_TYPE = %s.toMediaType()\n\n", reprStr(exactContentType))
					imports = append(imports, "okhttp3.MediaType.Companion.toMediaType")
				}
				methodCallArgs = append(methodCallArgs, "requestBody.toRequestBody(MEDIA_TYPE)")
				imports = append(imports, "okhttp3.RequestBody.Companion.toRequestBody")
				kotlinCode += fmt.Sprintf("val requestBody = %s\n\n", reprStr(r.Body))
			}
		} else {
			if exactContentType != "" {
				kotlinCode += fmt.Sprintf("val MEDIA_TYPE = %s.toMediaType()\n\n", reprStr(exactContentType))
				imports = append(imports, "okhttp3.MediaType.Companion.toMediaType")
			}
			methodCallArgs = append(methodCallArgs, "requestBody.toRequestBody(MEDIA_TYPE)")
			kotlinCode += fmt.Sprintf("val requestBody = %s\n\n", reprStr(r.Body))
			imports = append(imports, "okhttp3.RequestBody.Companion.toRequestBody")
		}
	}

	kotlinCode += "val request = Request.Builder()\n"
	kotlinCode += fmt.Sprintf("  .url(%s)\n", reprStr(r.URLs[0].URL))

	methods := map[string]bool{"DELETE": true, "GET": true, "HEAD": true, "PATCH": true, "POST": true, "PUT": true}
	dataMethods := map[string]bool{"DELETE": true, "PATCH": true, "POST": true, "PUT": true}
	requiredDataMethods := map[string]bool{"PATCH": true, "POST": true, "PUT": true}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	methodCall := "method"
	hasData := len(methodCallArgs) > 0
	if _, ok := methods[strings.ToUpper(method)]; !ok || (hasData && !dataMethods[strings.ToUpper(method)]) || (!hasData && requiredDataMethods[strings.ToUpper(method)]) {
		if !hasData {
			methodCallArgs = append([]string{`"".toRequestBody()`}, methodCallArgs...)
			imports = append(imports, "okhttp3.RequestBody.Companion.toRequestBody")
		}
		methodCallArgs = append([]string{reprStr(method)}, methodCallArgs...)
	} else {
		if strings.ToUpper(method) == "DELETE" && !hasData {
			methodCallArgs = append(methodCallArgs, `"".toRequestBody()`)
			imports = append(imports, "okhttp3.RequestBody.Companion.toRequestBody")
		}
		methodCall = strings.ToLower(method)
	}

	if methodCall != "get" {
		kotlinCode += "  ." + methodCall + "(" + strings.Join(methodCallArgs, ", ") + ")\n"
	}

	for _, h := range r.HeaderKV {
		kotlinCode += fmt.Sprintf("  .header(%s, %s)\n", reprStr(h.Key), reprStr(h.Value))
	}

	if r.BasicAuth != "" {
		kotlinCode += "  .header(\"Authorization\", credential)\n"
	}

	kotlinCode += "  .build()\n"

	kotlinCode += "\n"
	kotlinCode += "client.newCall(request).execute().use { response ->\n"
	kotlinCode += "  if (!response.isSuccessful) throw IOException(\"Unexpected code $response\")\n"
	kotlinCode += "  response.body!!.string()\n"
	kotlinCode += "}\n"

	var preambleCode string
	sort.Strings(imports)
	for _, imp := range imports {
		preambleCode += "import " + imp + "\n"
	}
	if len(imports) > 0 {
		preambleCode += "\n"
	}

	return preambleCode + kotlinCode
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`\$|"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "$":
			return "\\$"
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

func getContentType(r *request.Request) string {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return strings.TrimSpace(strings.Split(h.Value, ";")[0])
		}
	}
	return ""
}

func getExactContentType(r *request.Request) string {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return h.Value
		}
	}
	return ""
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
