package rust

import (
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

const indentation = "    "

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := map[string]bool{}
	lines := []string{"fn main() -> Result<(), Box<dyn std::error::Error>> {"}

	if len(r.HeaderKV) > 0 {
		imports["reqwest::header"] = true
		lines = append(lines, indent("let mut headers = header::HeaderMap::new();"))
		for _, h := range r.HeaderKV {
			name := `"` + h.Key + `"`
			if strings.EqualFold(h.Key, "Cookie") {
				name = "header::COOKIE"
			}
			lines = append(lines, indent("headers.insert("+name+", "+renderRustExpr(h.Value, imports)+".parse()?);"))
		}
		lines = append(lines, "")
	}

	if len(r.FormParts) > 0 {
		imports["reqwest::blocking::multipart"] = true
		lines = append(lines, indent("let form = multipart::Form::new()"))
		for i, part := range r.FormParts {
			line := ""
			if part.IsFile {
				line = indent(".file("+reprStr(part.Name)+", "+reprStr(part.FileName)+")?", 2)
			} else {
				line = indent(".text("+reprStr(part.Name)+", "+reprStr(part.Value)+")", 2)
			}
			if i == len(r.FormParts)-1 {
				line += ";"
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	if !r.FollowRedirects {
		lines = append(lines, indent("let client = reqwest::blocking::Client::builder()"))
		lines = append(lines, indent(".redirect(reqwest::redirect::Policy::none())", 2))
		lines = append(lines, indent(".build()?;", 2))
	} else if r.MaxRedirects == "" {
		lines = append(lines, indent("let client = reqwest::blocking::Client::new();"))
	} else {
		lines = append(lines, indent("let client = reqwest::blocking::Client::builder()"))
		if strings.TrimSpace(r.MaxRedirects) == "-1" {
			lines = append(lines, indent(".redirect(reqwest::redirect::Policy::custom(|attempt| { attempt.follow() }))", 2))
		} else {
			lines = append(lines, indent(".redirect(reqwest::redirect::Policy::limited("+strings.TrimSpace(r.MaxRedirects)+"))", 2))
		}
		lines = append(lines, indent(".build()?;", 2))
	}

	method := methodFor(r)
	if isSimpleReqwestMethod(method) {
		lines = append(lines, indent("let res = client."+strings.ToLower(method)+"("+reprStr(r.URLs[0].URL)+")"))
	} else {
		lines = append(lines, indent("let res = client.request("+reprStr(method)+", "+reprStr(r.URLs[0].URL)+")"))
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		lines = append(lines, indent(".basic_auth("+renderRustExpr(user, imports)+", Some("+renderRustExpr(pass, imports)+"))", 2))
	}
	if len(r.HeaderKV) > 0 {
		lines = append(lines, indent(".headers(headers)", 2))
	}
	if len(r.FormParts) > 0 {
		lines = append(lines, indent(".multipart(form)", 2))
	} else if r.Body != "" {
		lines = append(lines, indent(".body("+renderRustExpr(r.Body, imports)+")", 2))
	}

	lines = append(lines,
		indent(".send()?", 2),
		indent(".text()?;", 2),
		indent(`println!("{}", res);`),
		"",
		indent("Ok(())"),
		"}",
	)

	preamble := renderImports(imports)
	if preamble != "" {
		return preamble + "\n" + strings.Join(lines, "\n") + "\n"
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderImports(imports map[string]bool) string {
	if len(imports) == 0 {
		return ""
	}
	lines := []string{}
	if imports["reqwest::header"] && imports["reqwest::blocking::multipart"] {
		lines = append(lines, "use reqwest::{header, blocking::multipart};")
		delete(imports, "reqwest::header")
		delete(imports, "reqwest::blocking::multipart")
	} else {
		if imports["reqwest::blocking::multipart"] {
			lines = append(lines, "use reqwest::blocking::multipart;")
			delete(imports, "reqwest::blocking::multipart")
		}
		if imports["reqwest::header"] {
			lines = append(lines, "use reqwest::header;")
			delete(imports, "reqwest::header")
		}
	}
	items := make([]string, 0, len(imports))
	for imp := range imports {
		items = append(items, imp)
	}
	sort.Strings(items)
	for _, imp := range items {
		lines = append(lines, "use "+imp+";")
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func reprStr(s string) string {
	s = normalizeShellEscapes(s)
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\000':
			b.WriteString(`\0`)
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func renderRustExpr(s string, imports map[string]bool) string {
	parts := splitEnvInterpolations(normalizeShellEscapes(s))
	if len(parts) == 1 {
		return reprStr(parts[0])
	}
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "$") && len(part) > 1 {
			imports["std::env"] = true
			rendered = append(rendered, "env::var("+reprStr(part[1:])+").unwrap_or(\"\")")
		} else {
			rendered = append(rendered, reprStr(part))
		}
	}
	return "[" + strings.Join(rendered, ", ") + "].concat()"
}

func splitEnvInterpolations(s string) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '$' && i+1 < len(s) && isEnvStart(s[i+1]) {
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
		cur.WriteByte(s[i])
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
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isEnvPart(ch byte) bool {
	return isEnvStart(ch) || (ch >= '0' && ch <= '9')
}

func normalizeShellEscapes(s string) string {
	return strings.ReplaceAll(s, `\"`, `"`)
}

func indent(s string, level ...int) string {
	n := 1
	if len(level) > 0 {
		n = level[0]
	}
	return strings.Repeat(indentation, n) + s
}

func methodFor(r *request.Request) string {
	if r.Method != "" {
		return r.Method
	}
	return r.URLs[0].Method
}

func isSimpleReqwestMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
		return true
	default:
		return false
	}
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
