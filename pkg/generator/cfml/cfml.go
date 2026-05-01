package cfml

import (
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	lines := []string{
		"httpService = new http();",
		"httpService.setUrl(" + repr(r.URLs[0].URL) + ");",
		"httpService.setMethod(" + repr(methodFor(r)) + ");",
	}

	cookieHeaderPresent := false
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Cookie") {
			cookieHeaderPresent = true
			break
		}
	}

	if cookieHeaderPresent && len(r.Cookies) > 0 {
		for _, cookie := range r.Cookies {
			lines = append(lines,
				`httpService.addParam(type="cookie", name=`+repr(cookie[0])+`, value=`+repr(cookie[1])+`);`,
			)
		}
	}

	for _, h := range r.HeaderKV {
		if cookieHeaderPresent && strings.EqualFold(h.Key, "Cookie") {
			continue
		}
		lines = append(lines,
			`httpService.addParam(type="header", name=`+repr(h.Key)+`, value=`+repr(h.Value)+`);`,
		)
	}

	if r.MaxTime != "" {
		if timeout, err := strconv.Atoi(r.MaxTime); err == nil {
			lines = append(lines, "httpService.setTimeout("+strconv.Itoa(timeout)+");")
		} else {
			lines = append(lines, "httpService.setTimeout(0);")
		}
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		lines = append(lines, "httpService.setUsername("+repr(user)+");")
		lines = append(lines, "httpService.setPassword("+repr(pass)+");")
	}

	if r.Proxy != "" {
		proxyHost, proxyPort := splitProxy(r.Proxy)
		lines = append(lines, "httpService.setProxyServer("+repr(proxyHost)+");")
		lines = append(lines, "httpService.setProxyPort("+proxyPort+");")
		if r.ProxyAuth != "" {
			user, pass := splitBasicAuth(r.ProxyAuth)
			lines = append(lines, "httpService.setProxyUser("+repr(user)+");")
			lines = append(lines, "httpService.setProxyPassword("+repr(pass)+");")
		}
	}

	if len(r.FormParts) > 0 {
		for _, part := range r.FormParts {
			if part.IsFile {
				lines = append(lines,
					`httpService.addParam(type="file", name=`+repr(part.Name)+`, file="#expandPath(`+repr(part.FileName)+`)#");`,
				)
				continue
			}
			lines = append(lines,
				`httpService.addParam(type="formfield", name=`+repr(part.Name)+`, value=`+repr(part.Value)+`);`,
			)
		}
	} else if r.BodyFile != "" {
		reader := "fileRead"
		if r.IsDataBinary != nil && *r.IsDataBinary {
			reader = "fileReadBinary"
		}
		lines = append(lines,
			`httpService.addParam(type="body", value="#`+reader+`(expandPath(`+repr(r.BodyFile)+`))#");`,
		)
	} else if r.HasBody {
		lines = append(lines,
			`httpService.addParam(type="body", value=`+repr(r.Body)+`);`,
		)
	}

	lines = append(lines, "", "result = httpService.send().getPrefix();", "writeDump(result);")
	return strings.Join(lines, "\n") + "\n"
}

func methodFor(r *request.Request) string {
	if r.Method != "" {
		return r.Method
	}
	return r.URLs[0].Method
}

func repr(s string) string {
	s = strings.ReplaceAll(s, "#", "##")
	if strings.Contains(s, `"`) && !strings.Contains(s, `'`) {
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func splitProxy(proxy string) (string, string) {
	lastColon := strings.LastIndex(proxy, ":")
	if lastColon == -1 {
		return proxy, "1080"
	}
	port := proxy[lastColon+1:]
	if port == "" {
		return proxy, "1080"
	}
	for _, r := range port {
		if r < '0' || r > '9' {
			return proxy, "1080"
		}
	}
	return proxy[:lastColon], port
}
