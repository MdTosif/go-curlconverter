package curlconverter

import (
	"errors"

	gogen "github.com/mdtosif/go-curlconverter/pkg/generator/golang"
	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	jsjquerygen "github.com/mdtosif/go-curlconverter/pkg/generator/javascriptjquery"
	jsxhrgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascriptxhr"
	axgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeaxios"
	gotgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodegot"
	httpgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodehttp"
	kygen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeky"
	reqgen "github.com/mdtosif/go-curlconverter/pkg/generator/noderequest"
	superagentgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodesuperagent"
	pygen "github.com/mdtosif/go-curlconverter/pkg/generator/python"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type Warning = parser.Warning
type Warnings = parser.Warnings

var ErrNoRequests = errors.New("curlconverter: no requests parsed from input")

func SupportedLanguages() []string {
	return []string{"javascript", "javascript-jquery", "javascript-xhr", "node-http", "node-got", "node-ky", "node-request", "node-axios", "node-superagent", "go", "python", "parser"}
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
	return httpgen.Generate(req)
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
