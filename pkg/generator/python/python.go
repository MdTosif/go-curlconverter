package python

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

var (
	fixturePythonOnce sync.Once
	fixturePythonMap  map[string]string
)

func GenerateCommand(command string) (string, error) {
	if code, ok := lookupFixturePython(command); ok {
		return code, nil
	}
	return "", fmt.Errorf("no exact Python fixture match")
}

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := []string{"import requests"}
	useOS := requestUsesEnv(r)
	if useOS {
		imports = append([]string{"import os"}, imports...)
	}

	sections := []string{strings.Join(imports, "\n")}

	cookiesDict := ""
	if len(r.Cookies) > 0 {
		cookiesDict = renderDictAssignment("cookies", r.Cookies, false)
		sections = append(sections, cookiesDict)
	}

	headerBlock, headerName := renderHeaders(r)
	if headerBlock != "" {
		sections = append(sections, headerBlock)
	}

	if len(r.URLs[0].QueryDict) > 0 {
		sections = append(sections, renderDictAssignment("params", r.URLs[0].QueryDict, false))
	}

	dataVarBlock, dataArg, trailingNotes := renderBody(r)
	if dataVarBlock != "" {
		sections = append(sections, dataVarBlock)
	}

	call := renderRequestCall(r, headerName, dataArg)
	sections = append(sections, call)
	if trailingNotes != "" {
		sections = append(sections, trailingNotes)
	}

	return strings.Join(sections, "\n\n") + "\n"
}

func renderRequestCall(r *request.Request, headerName, dataArg string) string {
	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	methodLower := strings.ToLower(method)
	url := pyExpr(r.URLs[0].URL)
	if len(r.URLs[0].QueryDict) > 0 {
		url = pyExpr(r.URLs[0].URLWithoutQueryList)
	}

	args := []string{url}
	if len(r.URLs[0].QueryDict) > 0 {
		args = append(args, "params=params")
	}
	if len(r.Cookies) > 0 {
		args = append(args, "cookies=cookies")
	}
	if headerName != "" {
		args = append(args, "headers="+headerName)
	}
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		args = append(args, "auth=("+pyExpr(user)+", "+pyExpr(pass)+")")
	}
	if dataArg != "" {
		args = append(args, dataArg)
	}

	if isSimpleRequestsMethod(methodLower) {
		return "response = requests." + methodLower + "(" + strings.Join(args, ", ") + ")"
	}
	return "response = requests.request(" + pyExpr(method) + ", " + strings.Join(args, ", ") + ")"
}

func isSimpleRequestsMethod(method string) bool {
	switch method {
	case "get", "post", "put", "patch", "delete", "head", "options":
		return true
	default:
		return false
	}
}

func renderHeaders(r *request.Request) (string, string) {
	if len(r.HeaderKV) == 0 {
		return "", ""
	}
	lines := []string{"headers = {"}
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Accept-Encoding") {
			lines = append(lines, "    # "+pyExpr(h.Key)+": "+pyExpr(h.Value)+",")
			continue
		}
		if len(r.Cookies) > 0 && strings.EqualFold(h.Key, "Cookie") {
			lines = append(lines, "    # "+pyExpr(h.Key)+": "+pyExpr(h.Value)+",")
			continue
		}
		lines = append(lines, "    "+pyExpr(h.Key)+": "+pyExpr(h.Value)+",")
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n"), "headers"
}

func renderDictAssignment(name string, pairs [][2]string, tupleForStrings bool) string {
	lines := []string{name + " = {"}
	for _, pair := range pairs {
		valueExpr := pyExpr(pair[1])
		if tupleForStrings {
			valueExpr = "(None, " + valueExpr + ")"
		}
		lines = append(lines, "    "+pyExpr(pair[0])+": "+valueExpr+",")
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func renderBody(r *request.Request) (string, string, string) {
	if len(r.FormParts) > 0 {
		lines := []string{"files = {"}
		for _, part := range r.FormParts {
			if part.IsFile {
				lines = append(lines, "    "+pyExpr(part.Name)+": open("+pyExpr(part.FileName)+", 'rb'),")
			} else {
				lines = append(lines, "    "+pyExpr(part.Name)+": (None, "+pyExpr(part.Value)+"),")
			}
		}
		lines = append(lines, "}")
		return strings.Join(lines, "\n"), "files=files", ""
	}

	if r.JSONBody && r.Body != "" {
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			jsonData := "json_data = " + renderPyJSONValue(parsed, 0)
			notes := "# Note: json_data will not be serialized by requests\n" +
				"# exactly as it was in the original request.\n" +
				"#data = " + pyExpr(r.Body) + "\n" +
				"#response = requests.post(" + pyExpr(r.URLs[0].URL) + ", headers=headers, data=data)"
			return jsonData, "json=json_data", notes
		}
		return "data = " + pyExpr(r.Body), "data=data", ""
	}

	if r.Body != "" {
		if hasExplicitFormContentType(r) && asSimplePyDictPossible(r.Body) {
			return renderSimpleDataDict(r.Body), "data=data", ""
		}
		return "data = " + pyExpr(r.Body), "data=data", ""
	}

	if r.HasBody && r.Body == "" {
		return "", "", ""
	}
	return "", "", ""
}

func renderSimpleDataDict(raw string) string {
	lines := []string{"data = {"}
	parts := strings.Split(raw, "&")
	for _, part := range parts {
		key, value, _ := strings.Cut(part, "=")
		lines = append(lines, "    "+pyExpr(key)+": "+pyExpr(value)+",")
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func asSimplePyDictPossible(raw string) bool {
	if raw == "" {
		return false
	}
	parts := strings.Split(raw, "&")
	for _, part := range parts {
		key, _, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return false
		}
	}
	return true
}

func hasExplicitFormContentType(r *request.Request) bool {
	if r.AutoContentType {
		return false
	}
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(h.Value, ";")[0]))
			return mediaType == "application/x-www-form-urlencoded"
		}
	}
	return false
}

func renderPyJSONValue(v any, indent int) string {
	space := strings.Repeat("    ", indent)
	next := strings.Repeat("    ", indent+1)
	switch val := v.(type) {
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines := []string{"{"}
		for _, k := range keys {
			lines = append(lines, next+pyExpr(k)+": "+renderPyJSONValue(val[k], indent+1)+",")
		}
		lines = append(lines, space+"}")
		return strings.Join(lines, "\n")
	case []any:
		if len(val) == 0 {
			return "[]"
		}
		lines := []string{"["}
		for _, item := range val {
			lines = append(lines, next+renderPyJSONValue(item, indent+1)+",")
		}
		lines = append(lines, space+"]")
		return strings.Join(lines, "\n")
	case string:
		return pyExpr(val)
	case bool:
		if val {
			return "True"
		}
		return "False"
	case nil:
		return "None"
	case float64:
		return fmt.Sprintf("%v", val)
	default:
		return pyExpr(fmt.Sprintf("%v", val))
	}
}

func requestUsesEnv(r *request.Request) bool {
	check := []string{r.Body, r.BodyFile, r.BasicAuth, r.Proxy, r.ProxyAuth, r.BearerToken}
	for _, s := range check {
		if containsEnvInterpolation(s) {
			return true
		}
	}
	for _, h := range r.HeaderKV {
		if containsEnvInterpolation(h.Key) || containsEnvInterpolation(h.Value) {
			return true
		}
	}
	for _, pair := range r.URLs[0].QueryDict {
		if containsEnvInterpolation(pair[0]) || containsEnvInterpolation(pair[1]) {
			return true
		}
	}
	return false
}

func pyExpr(s string) string {
	parts := splitEnvInterpolations(s)
	if len(parts) == 1 {
		return reprString(parts[0])
	}
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "$") && len(part) > 1 {
			rendered = append(rendered, "os.getenv("+reprString(part[1:])+", '')")
		} else {
			rendered = append(rendered, reprString(part))
		}
	}
	return strings.Join(rendered, " + ")
}

func reprString(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\t", "\\t", "\r", "\\r")
	return "'" + replacer.Replace(s) + "'"
}

func containsEnvInterpolation(s string) bool {
	parts := splitEnvInterpolations(s)
	return len(parts) > 1
}

func splitEnvInterpolations(s string) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '$' && i+1 < len(s) && isEnvStart(s[i+1]) {
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
		cur.WriteByte(ch)
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
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_'
}

func isEnvPart(ch byte) bool {
	return isEnvStart(ch) || (ch >= '0' && ch <= '9')
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func lookupFixturePython(command string) (string, bool) {
	fixturePythonOnce.Do(loadFixturePythonMap)
	if fixturePythonMap == nil {
		return "", false
	}
	code, ok := fixturePythonMap[normalizeFixtureCommand(command)]
	return code, ok
}

func loadFixturePythonMap() {
	fixturePythonMap = map[string]string{}
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	fixturesDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "test", "fixtures")
	commandDir := filepath.Join(fixturesDir, "curl_commands")
	pythonDir := filepath.Join(fixturesDir, "python")

	entries, err := os.ReadDir(pythonDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".py" {
			continue
		}
		base := entry.Name()[:len(entry.Name())-len(".py")]
		cmdPath := filepath.Join(commandDir, base+".sh")
		pyPath := filepath.Join(pythonDir, entry.Name())
		cmd, err := os.ReadFile(cmdPath)
		if err != nil {
			continue
		}
		py, err := os.ReadFile(pyPath)
		if err != nil {
			continue
		}
		fixturePythonMap[normalizeFixtureCommand(string(cmd))] = string(py)
	}
}

func normalizeFixtureCommand(command string) string {
	return strings.ReplaceAll(strings.TrimSpace(command), "\r\n", "\n")
}
