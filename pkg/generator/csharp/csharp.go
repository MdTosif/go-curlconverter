package csharp

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

	imports := []string{"System.Net.Http"}

	methods := map[string]string{
		"DELETE": "Delete",
		"GET":    "Get",
		"PATCH":  "Patch",
		"POST":   "Post",
		"PUT":    "Put",
	}

	moreMethods := map[string]string{
		"DELETE":  "Delete",
		"GET":     "Get",
		"HEAD":    "Head",
		"OPTIONS": "Options",
		"PATCH":   "Patch",
		"POST":    "Post",
		"PUT":     "Put",
		"TRACE":   "Trace",
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	methodStr := "new HttpMethod(" + reprStr(method) + ")"
	if m, ok := moreMethods[strings.ToUpper(method)]; ok {
		methodStr = "HttpMethod." + m
	}

	simple := false
	if _, ok := methods[strings.ToUpper(method)]; ok {
		simple = len(r.HeaderKV) == 0 && r.BasicAuth == "" && len(r.FormParts) == 0 && r.Body == "" && r.BodyFile == "" && r.Output == ""
	}

	var s string

	if r.Insecure || r.Proxy != "" || r.Compressed {
		s += "HttpClientHandler handler = new HttpClientHandler();\n"
		if r.Insecure {
			s += "handler.ServerCertificateCustomValidationCallback = HttpClientHandler.DangerousAcceptAnyServerCertificateValidator;\n"
		}
		if r.Proxy != "" {
			s += "handler.Proxy = new WebProxy(" + reprStr(r.Proxy) + ");\n"
		}
		if r.Compressed {
			s += "handler.AutomaticDecompression = DecompressionMethods.All;\n"
		}
		s += "\n"
		s += "HttpClient client = new HttpClient(handler);\n\n"
	} else {
		s += "HttpClient client = new HttpClient();\n\n"
	}

	if simple {
		if strings.ToUpper(method) == "GET" {
			s += "string responseBody = await client.GetStringAsync(" + reprStr(r.URLs[0].URL) + ");\n"
		} else {
			s += "HttpResponseMessage response = await client." + methods[strings.ToUpper(method)] + "Async(" + reprStr(r.URLs[0].URL) + ");\n"
			s += "response.EnsureSuccessStatusCode();\n"
			s += "string responseBody = await response.Content.ReadAsStringAsync();\n"
		}
		if len(imports) > 0 {
			s = "using " + strings.Join(imports, ";\nusing ") + ";\n\n" + s
		}
		return s
	}

	s += "HttpRequestMessage request = new HttpRequestMessage(" + methodStr + ", " + reprStr(r.URLs[0].URL) + ");\n"

	contentHeaders := map[string]string{
		"content-length":   "ContentLength",
		"content-location": "ContentLocation",
		"content-md5":      "ContentMD5",
		"content-range":    "ContentRange",
		"content-type":     "ContentType",
		"expires":          "Expires",
		"last-modified":    "LastModified",
	}

	reqHeaders := make([][2]string, 0)
	reqContentHeaders := make([][2]string, 0)

	for _, h := range r.HeaderKV {
		lowerKey := strings.ToLower(h.Key)
		if _, ok := contentHeaders[lowerKey]; ok {
			reqContentHeaders = append(reqContentHeaders, [2]string{h.Key, h.Value})
		} else {
			reqHeaders = append(reqHeaders, [2]string{h.Key, h.Value})
		}
	}

	if len(reqHeaders) > 0 || r.BasicAuth != "" {
		s += "\n"
		for _, h := range reqHeaders {
			if strings.ToLower(h[0]) == "accept-encoding" {
				s += "// "
			}
			s += "request.Headers.Add(" + reprStr(h[0]) + ", " + reprStr(h[1]) + ");\n"
		}
		if r.BasicAuth != "" {
			s += "request.Headers.Add(\"Authorization\", \"Basic \" + Convert.ToBase64String(System.Text.ASCIIEncoding.ASCII.GetBytes(" + reprStr(r.BasicAuth) + ")));\n"
		}
		s += "\n"
	}

	if r.BodyFile != "" {
		s += "request.Content = new ByteArrayContent(File.ReadAllBytes(" + reprStr(r.BodyFile) + "));\n"
	} else if len(r.FormParts) > 0 {
		s += "\n"
		s += "MultipartFormDataContent content = new MultipartFormDataContent();\n"
		for _, part := range r.FormParts {
			name := reprStr(part.Name)
			s += "content.Add(new "
			if part.IsFile {
				s += "ByteArrayContent(File.ReadAllBytes(" + reprStr(part.FileName) + ")), " + name
			} else {
				s += "StringContent(" + reprStr(part.Value) + "), " + name
			}
			s += ");\n"
		}
		s += "request.Content = content;\n"
	} else if r.Body != "" {
		s += "request.Content = new StringContent(" + reprStr(r.Body) + ");\n"
	} else if hasContentType(r) {
		s += "request.Content = new StringContent(\"\");\n"
	}

	if len(reqContentHeaders) > 0 {
		for _, h := range reqContentHeaders {
			lowerKey := strings.ToLower(h[0])
			if lowerKey == "content-type" {
				imports = append(imports, "System.Net.Http.Headers")
				if strings.Contains(h[1], ";") {
					s += "request.Content.Headers.ContentType = MediaTypeHeaderValue.Parse(" + reprStr(h[1]) + ");\n"
				} else {
					s += "request.Content.Headers.ContentType = new MediaTypeHeaderValue(" + reprStr(h[1]) + ");\n"
				}
			} else if lowerKey == "content-length" {
				s += "// request.Content.Headers.ContentLength = " + h[1] + ";\n"
			} else if headerName, ok := contentHeaders[lowerKey]; ok {
				s += "request.Content.Headers." + headerName + " = " + reprStr(h[1]) + ";\n"
			}
		}
	}

	if r.BodyFile != "" || r.Body != "" || len(r.FormParts) > 0 || len(reqContentHeaders) > 0 {
		s += "\n"
	}

	s += "HttpResponseMessage response = await client.SendAsync(request);\n"
	s += "response.EnsureSuccessStatusCode();\n"
	s += "string responseBody = await response.Content.ReadAsStringAsync();\n"

	if len(imports) > 0 {
		sort.Strings(imports)
		s = "using " + strings.Join(imports, ";\nusing ") + ";\n\n" + s
	}
	return s
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\x00":
			return "\\0"
		case "\x07":
			return "\\a"
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
		case "\v":
			return "\\v"
		case "\\":
			return "\\\\"
		case "\"":
			return "\\\""
		default:
			if len(c) == 2 {
				first := c[0]
				second := c[1]
				return fmt.Sprintf("\\u%04X\\u%04X", first, second)
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "\\u" + hex
		}
	})
	return "\"" + escaped + "\""
}

func hasContentType(r *request.Request) bool {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return true
		}
	}
	return false
}
