package dart

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := []string{}

	if r.BasicAuth != "" || (r.IsDataBinary != nil && *r.IsDataBinary) {
		imports = append(imports, "dart:convert")
	}

	var s string
	s += "void main() async {\n"

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		s += fmt.Sprintf("  final uname = %s;\n", reprStr(user))
		s += fmt.Sprintf("  final pword = %s;\n", reprStr(pass))
		s += "  final authn = 'Basic ${base64Encode(utf8.encode('$uname:$pword'))}';\n"
		s += "\n"
	}

	methods := []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}
	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	rawRequestObj := len(r.FormParts) > 0 || !contains(methods, strings.ToUpper(method))
	hasHeaders := len(r.HeaderKV) > 0 || r.Compressed || (r.IsDataBinary != nil && *r.IsDataBinary) || strings.ToUpper(method) == "PUT"

	if hasHeaders && !rawRequestObj {
		s += "  final headers = {\n"
		for _, h := range r.HeaderKV {
			s += fmt.Sprintf("    %s: %s,\n", reprStr(h.Key), reprStr(h.Value))
		}
		if r.BasicAuth != "" {
			s += "    'Authorization': authn,\n"
		}
		if r.Compressed {
			s += "    'Accept-Encoding': 'gzip',\n"
		}
		s += "  };\n"
		s += "\n"
	}

	if len(r.URLs[0].QueryDict) > 0 {
		s += "  final params = {\n"
		for _, q := range r.URLs[0].QueryDict {
			s += fmt.Sprintf("    %s: %s,\n", reprStr(q[0]), reprStr(q[1]))
		}
		s += "  };\n"
		s += "\n"
	}

	hasData := r.Body != "" && len(r.FormParts) == 0
	if hasData {
		values, err := url.ParseQuery(r.Body)
		if err == nil && len(values) > 0 {
			s += "  final data = {\n"
			for k, vs := range values {
				for _, v := range vs {
					s += fmt.Sprintf("    %s: %s,\n", reprStr(k), reprStr(v))
				}
			}
			s += "  };\n"
			s += "\n"
		} else {
			s += fmt.Sprintf("  final data = %s;\n\n", reprStr(r.Body))
		}
	}

	if len(r.URLs[0].QueryDict) > 0 {
		urlString := reprStr(r.URLs[0].URLWithoutQueryList)
		s += fmt.Sprintf("  final url = Uri.parse(%s)\n", urlString)
		s += "      .replace(queryParameters: params);\n"
	} else {
		s += fmt.Sprintf("  final url = Uri.parse(%s);\n", reprStr(r.URLs[0].URL))
	}
	s += "\n"

	if rawRequestObj {
		multipart := "http."
		if len(r.FormParts) > 0 {
			multipart += "MultipartRequest"
		} else {
			multipart += "Request"
		}
		multipart += fmt.Sprintf("(%s, url)", reprStr(method))

		for _, part := range r.FormParts {
			name := reprStr(part.Name)
			if part.IsFile {
				multipart += "\n    ..files.add(await http.MultipartFile."
				multipart += fmt.Sprintf("fromPath(\n      %s, %s))", name, reprStr(part.FileName))
			} else {
				multipart += fmt.Sprintf("\n    ..fields[%s] = %s", name, reprStr(part.Value))
			}
		}
		multipart += ";\n\n"

		if hasHeaders || r.BasicAuth != "" {
			s += "  final req = " + multipart
			for _, h := range r.HeaderKV {
				s += fmt.Sprintf("  req.headers[%s] = %s;\n", reprStr(h.Key), reprStr(h.Value))
			}
			if len(r.HeaderKV) > 0 {
				s += "\n"
			}
			if r.BasicAuth != "" {
				s += "  req.headers['Authorization'] = authn;\n"
				s += "\n"
			}
			s += "  final stream = await req.send();\n"
			s += "  final res = await http.Response.fromStream(stream);\n"
		} else {
			s += "  final req = " + multipart
			s += "  final stream = await req.send();\n"
			s += "  final res = await http.Response.fromStream(stream);\n"
		}

		s += "  final status = res.statusCode;\n"
		s += "  if (status != 200) throw Exception('http.send error: statusCode= $status');\n"
		s += "\n"
		s += "  print(res.body);\n"
		s += "}"
	} else {
		s += fmt.Sprintf("  final res = await http.%s(url", strings.ToLower(method))
		if hasHeaders {
			s += ", headers: headers"
		} else if r.BasicAuth != "" {
			s += ", headers: {'Authorization': authn}"
		}
		if hasData {
			s += ", body: data"
		}
		s += ");\n"

		s += "  final status = res.statusCode;\n"
		s += fmt.Sprintf("  if (status != 200) throw Exception('http.%s error: statusCode= $status');\n", strings.ToLower(method))
		s += "\n"
		s += "  print(res.body);\n"
		s += "}"
	}

	var importString string
	sort.Strings(imports)
	for _, imp := range imports {
		importString += fmt.Sprintf("import '%s';\n", imp)
	}
	importString += "import 'package:http/http.dart' as http;\n"
	return importString + "\n" + s + "\n"
}

func reprStr(s string) string {
	quote := "'"
	if strings.Contains(s, "'") && !strings.Contains(s, "\"") {
		quote = "\""
	}
	return quote + escape(s, quote) + quote
}

func escape(s string, quote string) string {
	// JavaScript-style escaping with $ escaping for Dart
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
		case '\b':
			escaped.WriteString("\\b")
		case '\f':
			escaped.WriteString("\\f")
		case '$':
			escaped.WriteString("\\$")
		default:
			if string(c) == quote {
				escaped.WriteRune('\\')
			}
			escaped.WriteRune(c)
		}
	}
	return escaped.String()
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
