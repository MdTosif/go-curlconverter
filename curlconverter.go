package curlconverter

import (
	"errors"

	ansiblegen "github.com/mdtosif/go-curlconverter/pkg/generator/ansible"
	cgen "github.com/mdtosif/go-curlconverter/pkg/generator/c"
	cfmlgen "github.com/mdtosif/go-curlconverter/pkg/generator/cfml"
	clojuregen "github.com/mdtosif/go-curlconverter/pkg/generator/clojure"
	csharpgen "github.com/mdtosif/go-curlconverter/pkg/generator/csharp"
	dartgen "github.com/mdtosif/go-curlconverter/pkg/generator/dart"
	elixirgen "github.com/mdtosif/go-curlconverter/pkg/generator/elixir"
	gogen "github.com/mdtosif/go-curlconverter/pkg/generator/golang"
	hargen "github.com/mdtosif/go-curlconverter/pkg/generator/har"
	httplibgen "github.com/mdtosif/go-curlconverter/pkg/generator/http"
	httpiegen "github.com/mdtosif/go-curlconverter/pkg/generator/httpie"
	javagen "github.com/mdtosif/go-curlconverter/pkg/generator/java"
	javahttpurlconnectiongen "github.com/mdtosif/go-curlconverter/pkg/generator/javahttpurlconnection"
	javajsoupgen "github.com/mdtosif/go-curlconverter/pkg/generator/javajsoup"
	javaokhttpgen "github.com/mdtosif/go-curlconverter/pkg/generator/javaokhttp"
	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	jsjquerygen "github.com/mdtosif/go-curlconverter/pkg/generator/javascriptjquery"
	jsxhrgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascriptxhr"
	jsongen "github.com/mdtosif/go-curlconverter/pkg/generator/json"
	juliagen "github.com/mdtosif/go-curlconverter/pkg/generator/julia"
	kotlingen "github.com/mdtosif/go-curlconverter/pkg/generator/kotlin"
	lua "github.com/mdtosif/go-curlconverter/pkg/generator/lua"
	matlabgen "github.com/mdtosif/go-curlconverter/pkg/generator/matlab"
	axgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeaxios"
	fetchgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodefetch"
	gotgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodegot"
	httpgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodehttp"
	kygen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeky"
	reqgen "github.com/mdtosif/go-curlconverter/pkg/generator/noderequest"
	superagentgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodesuperagent"
	objectiveccgen "github.com/mdtosif/go-curlconverter/pkg/generator/objectivec"
	ocamlgen "github.com/mdtosif/go-curlconverter/pkg/generator/ocaml"
	perlgen "github.com/mdtosif/go-curlconverter/pkg/generator/perl"
	phpgen "github.com/mdtosif/go-curlconverter/pkg/generator/php"
	phpguzzlegen "github.com/mdtosif/go-curlconverter/pkg/generator/phpguzzle"
	phprequestsgen "github.com/mdtosif/go-curlconverter/pkg/generator/phprequests"
	powershellgen "github.com/mdtosif/go-curlconverter/pkg/generator/powershell"
	pygen "github.com/mdtosif/go-curlconverter/pkg/generator/python"
	pythonhttpgen "github.com/mdtosif/go-curlconverter/pkg/generator/pythonhttp"
	pythonrequestsgen "github.com/mdtosif/go-curlconverter/pkg/generator/pythonrequests"
	rgen "github.com/mdtosif/go-curlconverter/pkg/generator/r"
	rhttr2gen "github.com/mdtosif/go-curlconverter/pkg/generator/rhttr2"
	rubygen "github.com/mdtosif/go-curlconverter/pkg/generator/ruby"
	rubyhttpartygen "github.com/mdtosif/go-curlconverter/pkg/generator/rubyhttparty"
	rustgen "github.com/mdtosif/go-curlconverter/pkg/generator/rust"
	swiftgen "github.com/mdtosif/go-curlconverter/pkg/generator/swift"
	wgetgen "github.com/mdtosif/go-curlconverter/pkg/generator/wget"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type Warning = parser.Warning
type Warnings = parser.Warnings

var ErrNoRequests = errors.New("curlconverter: no requests parsed from input")

func SupportedLanguages() []string {
	return []string{"javascript", "javascript-jquery", "javascript-xhr", "node", "node-fetch", "node-http", "node-got", "node-ky", "node-request", "node-axios", "node-superagent", "go", "python", "json", "har", "http", "ansible", "cfml", "php-requests", "php-guzzle", "php", "rust", "r", "r-httr2", "ruby", "ruby-httparty", "wget", "clojure", "csharp", "kotlin", "httpie", "java-httpurlconnection", "java-jsoup", "java-okhttp", "java", "lua", "objectivec", "powershell", "dart", "perl", "ocaml", "elixir", "swift", "c", "julia", "python-http", "python-requests", "matlab", "parser"}
}

func Parse(curlCommand string) ([]*request.Request, error) {
	return parser.ParseAll(curlCommand)
}

func ParseWarn(curlCommand string) ([]*request.Request, Warnings, error) {
	return parser.ParseAllWarn(curlCommand)
}

func ParseArgs(args []string) ([]*request.Request, error) {
	return parser.ParseAllArgs(args)
}

func ParseArgsWarn(args []string) ([]*request.Request, Warnings, error) {
	return parser.ParseAllArgsWarn(args)
}

func ParseRequest(curlCommand string) (*request.Request, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return nil, err
	}
	if len(reqs) == 0 {
		return nil, ErrNoRequests
	}
	return reqs[0], nil
}

func ParseRequestWarn(curlCommand string) (*request.Request, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return nil, warnings, err
	}
	if len(reqs) == 0 {
		return nil, warnings, ErrNoRequests
	}
	return reqs[0], warnings, nil
}

func ParseRequestArgs(args []string) (*request.Request, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return nil, err
	}
	if len(reqs) == 0 {
		return nil, ErrNoRequests
	}
	return reqs[0], nil
}

func ParseRequestArgsWarn(args []string) (*request.Request, Warnings, error) {
	reqs, warnings, err := parser.ParseAllArgsWarn(args)
	if err != nil {
		return nil, warnings, err
	}
	if len(reqs) == 0 {
		return nil, warnings, ErrNoRequests
	}
	return reqs[0], warnings, nil
}

func ParseJSON(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return parser.MarshalJSON(reqs)
}

func ParseJSONWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	jsonOutput, err := parser.MarshalJSON(reqs)
	return jsonOutput, warnings, err
}

func ParseJSONArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	return parser.MarshalJSON(reqs)
}

func ParseJSONArgsWarn(args []string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllArgsWarn(args)
	if err != nil {
		return "", warnings, err
	}
	jsonOutput, err := parser.MarshalJSON(reqs)
	return jsonOutput, warnings, err
}

func GenerateJavaScript(req *request.Request) string {
	return jsgen.Generate(req)
}

func ToJavaScript(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateJavaScript(req), nil
}

func ToJavaScriptArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateJavaScript(req), nil
}

func ToJavaScriptWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{dataReadsFile: true})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJavaScript(req), warnings, nil
}

func GenerateNodeAxios(req *request.Request) string {
	return axgen.Generate(req)
}

func GenerateNode(req *request.Request) string {
	return fetchgen.Generate(req)
}

func GenerateNodeFetch(req *request.Request) string {
	return fetchgen.Generate(req)
}

func ToNode(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNode(req), nil
}

func ToNodeArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNode(req), nil
}

func ToNodeWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNode(req), warnings, nil
}

func ToNodeFetch(curlCommand string) (string, error) {
	return ToNode(curlCommand)
}

func ToNodeFetchArgs(args []string) (string, error) {
	return ToNodeArgs(args)
}

func ToNodeFetchWarn(curlCommand string) (string, Warnings, error) {
	return ToNodeWarn(curlCommand)
}

func ToNodeAxios(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeAxios(req), nil
}

func ToNodeAxiosArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeAxios(req), nil
}

func ToNodeAxiosWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNodeAxios(req), warnings, nil
}

func GenerateGo(req *request.Request) string {
	return gogen.Generate(req)
}

func ToGo(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateGo(req), nil
}

func ToGoArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateGo(req), nil
}

func ToGoWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{dataReadsFile: true})
	if err != nil {
		return "", warnings, err
	}
	return GenerateGo(req), warnings, nil
}

func GeneratePython(req *request.Request) string {
	return pygen.Generate(req)
}

func ToPython(curlCommand string) (string, error) {
	if code, err := pygen.GenerateCommand(curlCommand); err == nil {
		return code, nil
	}
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePython(req), nil
}

func ToPythonArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePython(req), nil
}

func ToPythonWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	if code, err := pygen.GenerateCommand(curlCommand); err == nil {
		return code, warnings, nil
	}
	return GeneratePython(req), warnings, nil
}

func GenerateJavaScriptXHR(req *request.Request) string {
	return jsxhrgen.Generate(req)
}

func ToJavaScriptXHR(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaScriptXHR(req), nil
}

func ToJavaScriptXHRArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaScriptXHR(req), nil
}

func ToJavaScriptXHRWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJavaScriptXHR(req), warnings, nil
}

func GenerateJavaScriptJQuery(req *request.Request) string {
	return jsjquerygen.Generate(req)
}

func ToJavaScriptJQuery(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaScriptJQuery(req), nil
}

func ToJavaScriptJQueryArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaScriptJQuery(req), nil
}

func ToJavaScriptJQueryWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJavaScriptJQuery(req), warnings, nil
}

func GenerateNodeHTTP(req *request.Request) string {
	return httplibgen.Generate(req)
}

func ToNodeHTTP(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeHTTP(req), nil
}

func ToNodeHTTPArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeHTTP(req), nil
}

func ToNodeHTTPWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNodeHTTP(req), warnings, nil
}

func GenerateNodeGot(req *request.Request) string {
	return gotgen.Generate(req)
}

func ToNodeGot(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeGot(req), nil
}

func ToNodeGotArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeGot(req), nil
}

func ToNodeGotWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNodeGot(req), warnings, nil
}

func GenerateNodeKy(req *request.Request) string {
	return kygen.Generate(req)
}

func ToNodeKy(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeKy(req), nil
}

func ToNodeKyArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeKy(req), nil
}

func ToNodeKyWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNodeKy(req), warnings, nil
}

func GenerateNodeRequest(req *request.Request) string {
	return reqgen.Generate(req)
}

func ToNodeRequest(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeRequest(req), nil
}

func ToNodeRequestArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeRequest(req), nil
}

func ToNodeRequestWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNodeRequest(req), warnings, nil
}

func GenerateNodeSuperagent(req *request.Request) string {
	return superagentgen.Generate(req)
}

func ToNodeSuperagent(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeSuperagent(req), nil
}

func ToNodeSuperagentArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateNodeSuperagent(req), nil
}

func ToNodeSuperagentWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateNodeSuperagent(req), warnings, nil
}

func GenerateJSON(req *request.Request) string {
	return jsongen.Generate(req)
}

func ToJSON(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJSON(req), nil
}

func ToJSONArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJSON(req), nil
}

func ToJSONWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJSON(req), warnings, nil
}

func GenerateHAR(req *request.Request) string {
	return hargen.Generate(req)
}

func ToHAR(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateHAR(req), nil
}

func ToHARArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateHAR(req), nil
}

func ToHARWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateHAR(req), warnings, nil
}

func GenerateHTTP(req *request.Request) string {
	return httpgen.Generate(req)
}

func ToHTTP(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateHTTP(req), nil
}

func ToHTTPArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateHTTP(req), nil
}

func ToHTTPWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateHTTP(req), warnings, nil
}

func GenerateAnsible(req *request.Request) string {
	return ansiblegen.Generate(req)
}

func ToAnsible(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateAnsible(req), nil
}

func ToAnsibleArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateAnsible(req), nil
}

func ToAnsibleWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{dataReadsFile: true})
	if err != nil {
		return "", warnings, err
	}
	return GenerateAnsible(req), warnings, nil
}

func GenerateCFML(req *request.Request) string {
	return cfmlgen.Generate(req)
}

func ToCFML(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateCFML(req), nil
}

func ToCFMLArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{dataReadsFile: true})
	if err != nil {
		return "", err
	}
	return GenerateCFML(req), nil
}

func ToCFMLWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{dataReadsFile: true})
	if err != nil {
		return "", warnings, err
	}
	return GenerateCFML(req), warnings, nil
}

func GeneratePHPRequests(req *request.Request) string {
	return phprequestsgen.Generate(req)
}

func ToPHPRequests(curlCommand string) (string, error) {
	if code, err := phprequestsgen.GenerateCommand(curlCommand); err == nil {
		return code, nil
	}
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePHPRequests(req), nil
}

func ToPHPRequestsArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePHPRequests(req), nil
}

func ToPHPRequestsWarn(curlCommand string) (string, Warnings, error) {
	if code, err := phprequestsgen.GenerateCommand(curlCommand); err == nil {
		return code, nil, nil
	}
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	req, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePHPRequests(req), warnings, nil
}

func GenerateRust(req *request.Request) string {
	return rustgen.Generate(req)
}

func ToRust(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateRust(req), nil
}

func ToRustArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateRust(req), nil
}

func ToRustWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateRust(r), warnings, nil
}

func GenerateR(req *request.Request) string {
	return rgen.Generate(req)
}

func ToR(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateR(req), nil
}

func ToRArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateR(req), nil
}

func ToRWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateR(r), warnings, nil
}

func GenerateRHttr2(req *request.Request) string {
	return rhttr2gen.Generate(req)
}

func ToRHttr2(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateRHttr2(req), nil
}

func ToRHttr2Args(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateRHttr2(req), nil
}

func ToRHttr2Warn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateRHttr2(r), warnings, nil
}

func GenerateRubyHTTParty(req *request.Request) string {
	return rubyhttpartygen.Generate(req)
}

func ToRubyHTTParty(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return "require 'httparty'\n\n" + GenerateRubyHTTParty(req), nil
}

func ToRubyHTTPartyArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return "require 'httparty'\n\n" + GenerateRubyHTTParty(req), nil
}

func ToRubyHTTPartyWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return "require 'httparty'\n\n" + GenerateRubyHTTParty(r), warnings, nil
}

func GenerateWget(req *request.Request) string {
	return wgetgen.Generate(req)
}

func ToWget(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateWget(req), nil
}

func ToWgetArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateWget(req), nil
}

func ToWgetWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateWget(r), warnings, nil
}

func GeneratePhpGuzzle(req *request.Request) string {
	return phpguzzlegen.Generate(req)
}

func ToPhpGuzzle(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePhpGuzzle(req), nil
}

func ToPhpGuzzleArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePhpGuzzle(req), nil
}

func ToPhpGuzzleWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePhpGuzzle(r), warnings, nil
}

func GenerateClojure(req *request.Request) string {
	return clojuregen.Generate(req)
}

func ToClojure(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateClojure(req), nil
}

func ToClojureArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateClojure(req), nil
}

func ToClojureWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateClojure(r), warnings, nil
}

func GenerateCSharp(req *request.Request) string {
	return csharpgen.Generate(req)
}

func ToCSharp(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateCSharp(req), nil
}

func ToCSharpArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateCSharp(req), nil
}

func ToCSharpWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateCSharp(r), warnings, nil
}

func GenerateKotlin(req *request.Request) string {
	return kotlingen.Generate(req)
}

func ToKotlin(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateKotlin(req), nil
}

func ToKotlinArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateKotlin(req), nil
}

func ToKotlinWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateKotlin(r), warnings, nil
}

func GenerateHttpie(req *request.Request) string {
	return httpiegen.Generate(req)
}

func ToHttpie(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateHttpie(req), nil
}

func ToHttpieArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateHttpie(req), nil
}

func ToHttpieWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateHttpie(r), warnings, nil
}

func GenerateJavaHttpUrlConnection(req *request.Request) string {
	return javahttpurlconnectiongen.Generate(req)
}

func ToJavaHttpUrlConnection(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaHttpUrlConnection(req), nil
}

func ToJavaHttpUrlConnectionArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaHttpUrlConnection(req), nil
}

func ToJavaHttpUrlConnectionWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJavaHttpUrlConnection(r), warnings, nil
}

func GenerateJavaJsoup(req *request.Request) string {
	return javajsoupgen.Generate(req)
}

func ToJavaJsoup(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaJsoup(req), nil
}

func ToJavaJsoupArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaJsoup(req), nil
}

func ToJavaJsoupWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJavaJsoup(r), warnings, nil
}

func GenerateLua(req *request.Request) string {
	return lua.Generate(req)
}

func ToLua(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateLua(req), nil
}

func ToLuaArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateLua(req), nil
}

func ToLuaWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateLua(r), warnings, nil
}

func GenerateObjectiveC(req *request.Request) string {
	return objectiveccgen.Generate(req)
}

func ToObjectiveC(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateObjectiveC(req), nil
}

func ToObjectiveCArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateObjectiveC(req), nil
}

func ToObjectiveCWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateObjectiveC(r), warnings, nil
}

func GeneratePowerShell(req *request.Request) string {
	return powershellgen.Generate(req)
}

func ToPowerShell(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePowerShell(req), nil
}

func ToPowerShellArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePowerShell(req), nil
}

func ToPowerShellWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePowerShell(r), warnings, nil
}

func GenerateDart(req *request.Request) string {
	return dartgen.Generate(req)
}

func ToDart(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateDart(req), nil
}

func ToDartArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateDart(req), nil
}

func ToDartWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateDart(r), warnings, nil
}

func GeneratePerl(req *request.Request) string {
	return perlgen.Generate(req)
}

func ToPerl(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePerl(req), nil
}

func ToPerlArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePerl(req), nil
}

func ToPerlWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePerl(r), warnings, nil
}

func GenerateOCaml(req *request.Request) string {
	return ocamlgen.Generate(req)
}

func ToOCaml(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateOCaml(req), nil
}

func ToOCamlArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateOCaml(req), nil
}

func ToOCamlWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateOCaml(r), warnings, nil
}

func GenerateElixir(req *request.Request) string {
	return elixirgen.Generate(req)
}

func ToElixir(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateElixir(req), nil
}

func ToElixirArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateElixir(req), nil
}

func ToElixirWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateElixir(r), warnings, nil
}

func GenerateJavaOkHttp(req *request.Request) string {
	return javaokhttpgen.Generate(req)
}

func ToJavaOkHttp(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaOkHttp(req), nil
}

func ToJavaOkHttpArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJavaOkHttp(req), nil
}

func ToJavaOkHttpWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJavaOkHttp(r), warnings, nil
}

func GeneratePHP(req *request.Request) string {
	return phpgen.Generate(req)
}

func ToPHP(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePHP(req), nil
}

func ToPHPArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePHP(req), nil
}

func ToPHPWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePHP(r), warnings, nil
}

func GenerateSwift(req *request.Request) string {
	return swiftgen.Generate(req)
}

func ToSwift(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateSwift(req), nil
}

func ToSwiftArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateSwift(req), nil
}

func ToSwiftWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateSwift(r), warnings, nil
}

func GenerateC(req *request.Request) string {
	return cgen.Generate(req)
}

func ToC(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateC(req), nil
}

func ToCArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateC(req), nil
}

func ToCWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateC(r), warnings, nil
}

func GenerateJulia(req *request.Request) string {
	return juliagen.Generate(req)
}

func ToJulia(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJulia(req), nil
}

func ToJuliaArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJulia(req), nil
}

func ToJuliaWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJulia(r), warnings, nil
}

func GeneratePythonHttp(req *request.Request) string {
	return pythonhttpgen.Generate(req)
}

func ToPythonHttp(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePythonHttp(req), nil
}

func ToPythonHttpArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePythonHttp(req), nil
}

func ToPythonHttpWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePythonHttp(r), warnings, nil
}

func GenerateJava(req *request.Request) string {
	return javagen.Generate(req)
}

func ToJava(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJava(req), nil
}

func ToJavaArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateJava(req), nil
}

func ToJavaWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateJava(r), warnings, nil
}

func GeneratePythonRequests(req *request.Request) string {
	return pythonrequestsgen.Generate(req)
}

func ToPythonRequests(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePythonRequests(req), nil
}

func ToPythonRequestsArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GeneratePythonRequests(req), nil
}

func ToPythonRequestsWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GeneratePythonRequests(r), warnings, nil
}

func GenerateMatlab(req *request.Request) string {
	return matlabgen.Generate(req)
}

func ToMatlab(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateMatlab(req), nil
}

func ToMatlabArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateMatlab(req), nil
}

func ToMatlabWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateMatlab(r), warnings, nil
}

func GenerateRuby(req *request.Request) string {
	return rubygen.Generate(req)
}

func ToRuby(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateRuby(req), nil
}

func ToRubyArgs(args []string) (string, error) {
	reqs, err := parser.ParseAllArgs(args)
	if err != nil {
		return "", err
	}
	req, _, err := getFirstRequest(reqs, nil, support{})
	if err != nil {
		return "", err
	}
	return GenerateRuby(req), nil
}

func ToRubyWarn(curlCommand string) (string, Warnings, error) {
	reqs, warnings, err := parser.ParseAllWarn(curlCommand)
	if err != nil {
		return "", warnings, err
	}
	r, warnings, err := getFirstRequest(reqs, warnings, support{})
	if err != nil {
		return "", warnings, err
	}
	return GenerateRuby(r), warnings, nil
}
