package powershell

import (
	"fmt"
	urlpkg "net/url"
	"regexp"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	return GenerateRestMethod(r)
}

func GenerateRestMethod(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}
	return toPowershell(r, true)
}

func GenerateWebRequest(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}
	return toPowershell(r, false)
}

func toPowershell(r *request.Request, restMethod bool) string {
	var commands []string
	for _, url := range r.URLs {
		commands = append(commands, requestToPowershell(r, url, restMethod))
	}
	return strings.Join(commands, "\n\n")
}

func requestToPowershell(r *request.Request, url request.RequestURL, restMethod bool) string {
	var code string
	command := "Invoke-RestMethod"
	if !restMethod {
		command = "Invoke-WebRequest"
	}

	args := [][]string{}
	method := url.Method
	if method == "" {
		method = "GET"
	}

	// Handle query parameters as body for GET requests
	if len(url.QueryList) > 0 && strings.ToUpper(method) == "GET" && r.Body == "" && len(r.FormParts) == 0 && r.BodyFile == "" {
		code += "$body = @{\n"
		for _, q := range url.QueryList {
			code += fmt.Sprintf("    %s = %s\n", reprStr(q[0]), reprStr(q[1]))
		}
		code += "}\n"
		args = append(args, []string{"-Uri", reprStr(url.URLWithoutQueryList)})
		args = append(args, []string{"-Body", "$body"})
	} else {
		args = append(args, []string{"-Uri", reprStr(url.URL)})
	}

	// Method handling
	methods := map[string]string{
		"DEFAULT": "Default",
		"DELETE":  "Delete",
		"GET":     "Get",
		"HEAD":    "Head",
		"MERGE":   "Merge",
		"OPTIONS": "Options",
		"PATCH":   "Patch",
		"POST":    "Post",
		"PUT":     "Put",
		"TRACE":   "Trace",
	}
	if upperMethod := strings.ToUpper(method); upperMethod == "GET" {
		// default
	} else if m, ok := methods[upperMethod]; ok {
		args = append(args, []string{"-Method", m})
	} else {
		args = append(args, []string{"-CustomMethod", reprStr(method)})
	}

	// Cookies
	if len(r.Cookies) > 0 {
		code += "$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession\n"
		for _, c := range r.Cookies {
			code += "$cookie = New-Object System.Net.Cookie\n"
			code += fmt.Sprintf("$cookie.Name = %s\n", reprStr(c[0]))
			code += fmt.Sprintf("$cookie.Value = %s\n", reprStr(c[1]))
			code += fmt.Sprintf("$cookie.Domain = %s\n", reprStr(url.URLObj.Host))
			code += "$session.Cookies.Add($cookie)\n"
		}
		args = append(args, []string{"-WebSession", "$session"})
	}

	// Headers
	headerArgs := [][]string{}
	headerLines := []string{}
	for _, h := range r.HeaderKV {
		if h.Value == "" {
			continue
		}
		if strings.EqualFold(h.Key, "content-type") {
			headerArgs = append(headerArgs, []string{"-ContentType", reprStr(h.Value)})
		} else if strings.EqualFold(h.Key, "user-agent") {
			headerArgs = append(headerArgs, []string{"-UserAgent", reprStr(h.Value)})
		} else {
			headerLines = append(headerLines, fmt.Sprintf("%s = %s", reprStr(h.Key), reprStr(h.Value)))
		}
	}

	// Authentication
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		code += fmt.Sprintf("$username = %s\n", reprStr(user))
		code += fmt.Sprintf("$password = ConvertTo-SecureString %s -AsPlainText -Force\n", reprStr(pass))
		code += "$credential = New-Object System.Management.Automation.PSCredential($username, $password)\n"
		args = append(args, []string{"-Credential", "$credential"})
		args = append(args, []string{"-Authentication", "Basic"})
		if strings.EqualFold(url.URLObj.Scheme, "http") {
			args = append(args, []string{"-AllowUnencryptedAuthentication"})
		}
	}

	if len(headerLines) > 0 {
		code += "$headers = @{\n"
		code += "    " + strings.Join(headerLines, "\n    ") + "\n"
		code += "}\n"
		args = append(args, []string{"-Headers", "$headers"})
	}
	args = append(args, headerArgs...)

	// Body handling
	if r.BodyFile != "" {
		args = append(args, []string{"-InFile", reprStr(r.BodyFile)})
	} else if len(r.FormParts) > 0 {
		code += "$form = @{\n"
		for _, part := range r.FormParts {
			if part.IsFile {
				code += fmt.Sprintf("    %s = Get-Item %s\n", reprStr(part.Name), reprStr(part.FileName))
			} else {
				code += fmt.Sprintf("    %s = %s\n", reprStr(part.Name), reprStr(part.Value))
			}
		}
		code += "}\n"
		args = append(args, []string{"-Form", "$form"})
	} else if r.Body != "" {
		contentType := getContentType(r)
		if contentType == "application/x-www-form-urlencoded" {
			values, err := urlpkg.ParseQuery(r.Body)
			if err == nil && len(values) > 0 {
				code += "$body = @{\n"
				for k, vs := range values {
					for _, v := range vs {
						code += fmt.Sprintf("    %s = %s\n", reprStr(k), reprStr(v))
					}
				}
				code += "}\n"
				args = append(args, []string{"-Body", "$body"})
			} else {
				args = append(args, []string{"-Body", reprStr(r.Body)})
			}
		} else {
			args = append(args, []string{"-Body", reprStr(r.Body)})
		}
	}

	// Redirects
	if !r.FollowRedirects {
		args = append(args, []string{"-MaximumRedirection", "0"})
	} else if r.MaxRedirects != "" && r.MaxRedirects != "5" {
		args = append(args, []string{"-MaximumRedirection", r.MaxRedirects})
	}

	// Proxy
	if r.Proxy != "" {
		args = append(args, []string{"-Proxy", reprStr(r.Proxy)})
	}

	// SSL
	if r.Insecure {
		args = append(args, []string{"-SkipCertificateCheck"})
	}

	// Timeout
	if r.MaxTime != "" {
		args = append(args, []string{"-TimeoutSec", r.MaxTime})
	}

	// HTTP version
	if r.HTTP2 {
		args = append(args, []string{"-HttpVersion", "2.0"})
	}

	// Multiline formatting
	multiline := len(args) > 3 || totalArgsLength(args) > 75
	joiner := " "
	if multiline {
		joiner = " `\n    "
	}

	joinedArgs := []string{}
	for _, arg := range args {
		joinedArgs = append(joinedArgs, strings.Join(arg, " "))
	}

	return code + "$response = " + command + " " + strings.Join(joinedArgs, joiner) + "\n"
}

func reprStr(s string) string {
	quote := "\""
	if !strings.ContainsAny(s, "$`\"") && strings.Contains(s, "'") && !strings.Contains(s, "\"") {
		quote = "'"
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}

	regexDoubleEscape := regexp.MustCompile(`\$|"|\p{C}|[^ \P{Z}]`)
	escaped := regexDoubleEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\x00":
			return "`0"
		case "\a":
			return "`a"
		case "\b":
			return "`b"
		case "\x1b":
			return "`e"
		case "\f":
			return "`f"
		case "\n":
			return "`n"
		case "\r":
			return "`r"
		case "\t":
			return "`t"
		case "\v":
			return "`v"
		case "$":
			return "`$"
		case "`":
			return "``"
		case "\"":
			return "`\""
		default:
			if len(c) == 1 {
				hex := fmt.Sprintf("%02X", c[0])
				return "`u{" + hex + "}"
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "`u{" + hex + "}"
		}
	})
	return quote + escaped + quote
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func getContentType(r *request.Request) string {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return strings.TrimSpace(strings.Split(h.Value, ";")[0])
		}
	}
	return ""
}

func totalArgsLength(args [][]string) int {
	total := 0
	for _, arg := range args {
		total += len(arg[0]) + len(arg[1])
	}
	return total
}
