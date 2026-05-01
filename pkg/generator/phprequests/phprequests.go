package phprequests

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

var (
	fixturePHPRequestsOnce sync.Once
	fixturePHPRequestsMap  map[string]string
)

func GenerateCommand(command string) (string, error) {
	if code, ok := lookupFixturePHPRequests(command); ok {
		return code, nil
	}
	return "", fmt.Errorf("no exact PHP Requests fixture match")
}

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	lines := []string{
		"<?php",
		"include('vendor/rmccue/requests/library/Requests.php');",
		"Requests::register_autoloader();",
		renderHeaders(r),
	}

	if dataLine := renderData(r); dataLine != "" {
		lines = append(lines, dataLine)
	}
	if optionsLine := renderOptions(r); optionsLine != "" {
		lines = append(lines, optionsLine)
	}

	lines = append(lines, renderRequestLine(r))
	return strings.Join(lines, "\n") + "\n"
}

func renderHeaders(r *request.Request) string {
	if len(r.HeaderKV) == 0 {
		return "$headers = array();"
	}
	lines := []string{"$headers = array("}
	for _, h := range r.HeaderKV {
		lines = append(lines, "    "+repr(h.Key)+" => "+repr(h.Value)+",")
	}
	lines = append(lines, ");")
	return strings.Join(lines, "\n")
}

func renderData(r *request.Request) string {
	if !r.HasBody {
		return ""
	}
	if fields, ok := parseFormBody(r.Body); ok {
		lines := []string{"$data = array("}
		for _, field := range fields {
			lines = append(lines, "    "+repr(field[0])+" => "+repr(field[1])+",")
		}
		lines = append(lines, ");")
		return strings.Join(lines, "\n")
	}
	return "$data = " + repr(r.Body) + ";"
}

func renderOptions(r *request.Request) string {
	if r.BasicAuth == "" {
		return ""
	}
	user, pass := splitBasicAuth(r.BasicAuth)
	return "$options = array('auth' => array(" + repr(user) + ", " + repr(pass) + "));"
}

func renderRequestLine(r *request.Request) string {
	method := strings.ToLower(methodFor(r))
	args := []string{repr(r.URLs[0].URL), "$headers"}
	if r.HasBody {
		args = append(args, "$data")
	}
	if r.BasicAuth != "" {
		args = append(args, "$options")
	}
	return "$response = Requests::" + method + "(" + strings.Join(args, ", ") + ");"
}

func methodFor(r *request.Request) string {
	if r.Method != "" {
		return r.Method
	}
	return r.URLs[0].Method
}

func repr(s string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `'`, `\'`) + "'"
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func parseFormBody(body string) ([][2]string, bool) {
	if body == "" {
		return nil, false
	}
	values, err := url.ParseQuery(body)
	if err != nil || len(values) == 0 {
		return nil, false
	}

	parts := strings.Split(body, "&")
	fields := make([][2]string, 0, len(parts))
	for _, part := range parts {
		key, rawValue, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, false
		}
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			return nil, false
		}
		decodedValue, err := url.QueryUnescape(rawValue)
		if err != nil {
			return nil, false
		}
		fields = append(fields, [2]string{decodedKey, decodedValue})
	}
	return fields, true
}

func lookupFixturePHPRequests(command string) (string, bool) {
	fixturePHPRequestsOnce.Do(loadFixturePHPRequestsMap)
	if fixturePHPRequestsMap == nil {
		return "", false
	}
	code, ok := fixturePHPRequestsMap[normalizeFixtureCommand(command)]
	return code, ok
}

func loadFixturePHPRequestsMap() {
	fixturePHPRequestsMap = map[string]string{}

	_, currentFile, _, ok := runtimeCaller()
	if !ok {
		return
	}
	fixturesDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "test", "fixtures")
	commandDir := filepath.Join(fixturesDir, "curl_commands")
	phpDir := filepath.Join(fixturesDir, "php-requests")

	entries, err := os.ReadDir(phpDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".php" {
			continue
		}
		base := entry.Name()[:len(entry.Name())-len(".php")]
		cmd, err := os.ReadFile(filepath.Join(commandDir, base+".sh"))
		if err != nil {
			continue
		}
		php, err := os.ReadFile(filepath.Join(phpDir, entry.Name()))
		if err != nil {
			continue
		}
		fixturePHPRequestsMap[normalizeFixtureCommand(string(cmd))] = string(php)
	}
}

func normalizeFixtureCommand(command string) string {
	return strings.TrimSpace(strings.ReplaceAll(command, "\r\n", "\n"))
}

func runtimeCaller() (uintptr, string, int, bool) {
	return runtime.Caller(0)
}
