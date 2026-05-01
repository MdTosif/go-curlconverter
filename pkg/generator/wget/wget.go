package wget

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var args []string

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	hasData := r.BodyFile != "" || r.Body != "" || len(r.FormParts) > 0
	if hasData {
		if method != "POST" {
			args = append(args, "--method="+reprStr(method))
		}
	} else if method != "GET" {
		args = append(args, "--method="+reprStr(method))
	}

	if len(r.CookieFiles) > 0 {
		for _, cookieFile := range r.CookieFiles {
			args = append(args, "--load-cookies="+reprStr(cookieFile))
		}
	}

	if r.CookieJar != "" {
		args = append(args, "--save-cookies="+reprStr(r.CookieJar))
	}

	for _, h := range r.HeaderKV {
		args = append(args, "--header="+reprStr(h.Key+": "+h.Value))
	}

	if r.Compressed {
		args = append(args, "--compression=auto")
	}

	if r.MaxRedirects != "" && r.MaxRedirects != "20" {
		args = append(args, "--max-redirect="+reprStr(r.MaxRedirects))
	}

	if r.IPv4 {
		args = append(args, "--inet4-only")
	}
	if r.IPv6 {
		args = append(args, "--inet6-only")
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		args = append(args, "--user="+reprStr(user))
		if pass != "" {
			args = append(args, "--password="+reprStr(pass))
		}
		if r.AuthType == "basic" {
			args = append(args, "--auth-no-challenge")
		}
	}

	if r.BodyFile != "" {
		if method == "POST" {
			args = append(args, "--post-file="+reprStr(r.BodyFile))
		} else {
			args = append(args, "--body-file="+reprStr(r.BodyFile))
		}
	} else if len(r.FormParts) > 0 {
		// Wget doesn't support multipart
	} else if r.Body != "" {
		if method == "POST" {
			args = append(args, "--post-data="+reprStr(r.Body))
		} else {
			args = append(args, "--body-data="+reprStr(r.Body))
		}
	}

	if r.Insecure {
		args = append(args, "--no-check-certificate")
	}

	if r.CACert != "" {
		args = append(args, "--ca-certificate="+reprStr(r.CACert))
	}
	if r.Cert != "" {
		args = append(args, "--certificate="+reprStr(r.Cert))
	}
	if r.Key != "" {
		args = append(args, "--private-key="+reprStr(r.Key))
	}

	if r.Netrc == "ignored" {
		args = append(args, "--no-netrc")
	}

	if r.NoProxy == "*" {
		args = append(args, "--no-proxy")
	}

	if r.MaxTime != "" {
		args = append(args, "--timeout="+reprStr(r.MaxTime))
	}
	if r.ConnectTimeout != "" {
		args = append(args, "--connect-timeout="+reprStr(r.ConnectTimeout))
	}

	if r.SpeedLimit != "" {
		args = append(args, "--limit-rate="+reprStr(r.SpeedLimit))
	}

	if r.Output != "" {
		args = append(args, "--output-document="+reprStr(r.Output))
	} else {
		args = append(args, "--output-document -")
	}

	if !r.Clobber {
		args = append(args, "--no-clobber")
	}

	if !r.RemoteTime {
		args = append(args, "--no-use-server-timestamps")
	}

	if r.ContinueAt != "" {
		if r.ContinueAt == "-" {
			args = append(args, "--continue")
		} else {
			args = append(args, "--start-pos="+reprStr(r.ContinueAt))
		}
	}

	// KeepAlive not available in Go request struct, skip for now

	if r.Globoff {
		args = append(args, "--no-glob")
	}

	if r.Silent {
		args = append(args, "--quiet")
	} else if !r.Verbose {
		args = append(args, "--no-verbose")
	}

	for _, url := range r.URLs {
		args = append(args, reprStr(url.URL))
	}

	multiline := len(args) > 3 || totalLength(args) > 75
	joiner := " "
	if multiline {
		joiner = " \\\n  "
	}

	return "wget " + strings.Join(args, joiner) + "\n"
}

func reprStr(s string) string {
	// Check for unprintable characters
	if containsUnprintable(s) {
		return "$'" + escapeAnsiC(s) + "'"
	}

	// Check if we need quotes
	needsQuote := needsQuoting(s)
	if needsQuote && !strings.Contains(s, "'") {
		return "'" + s + "'"
	}

	if needsQuote {
		return "\"" + escapeDoubleQuote(s) + "\""
	}

	return s
}

func containsUnprintable(s string) bool {
	for _, r := range s {
		if r < 32 || r == 127 || (r >= 0x7F && r <= 0x9F) {
			return true
		}
	}
	return false
}

func escapeAnsiC(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\a':
			result.WriteString("\\a")
		case '\b':
			result.WriteString("\\b")
		case '\x1b':
			result.WriteString("\\e")
		case '\f':
			result.WriteString("\\f")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		case '\v':
			result.WriteString("\\v")
		case '\\':
			result.WriteString("\\\\")
		case '\'':
			result.WriteString("\\'")
		default:
			if r < 32 || r == 127 || (r >= 0x7F && r <= 0x9F) {
				hex := fmt.Sprintf("\\x%02X", r)
				result.WriteString(hex)
			} else if r > 0xFFFF {
				hex := fmt.Sprintf("\\U%08X", r)
				result.WriteString(hex)
			} else if r > 0xFF {
				hex := fmt.Sprintf("\\u%04X", r)
				result.WriteString(hex)
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

func escapeDoubleQuote(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '$':
			result.WriteString("\\$")
		case '`':
			result.WriteString("\\`")
		case '"':
			result.WriteString("\\\"")
		case '\\':
			result.WriteString("\\\\")
		case '!':
			result.WriteString("\"'!'\"")
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	shellChars := regexp.MustCompile(`[\x02-\x09\x0b-\x1a\\#?()\x60{}\[\]^*<=>~|; "!$&'\x82-\xff]`)
	return shellChars.MatchString(s) || strings.ContainsAny(s, "\" $`!")
}

func totalLength(args []string) int {
	total := 0
	for _, arg := range args {
		total += len(arg)
	}
	return total
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
