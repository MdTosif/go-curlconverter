package curlconverter

import (
	"errors"

	gogen "github.com/mdtosif/go-curlconverter/pkg/generator/golang"
	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	axgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeaxios"
	pygen "github.com/mdtosif/go-curlconverter/pkg/generator/python"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type Warning [2]string
type Warnings []Warning

var ErrNoRequests = errors.New("curlconverter: no requests parsed from input")

func SupportedLanguages() []string {
	return []string{"javascript", "node-axios", "go", "python", "parser"}
}

func Parse(curlCommand string) ([]*request.Request, error) {
	return parser.ParseAll(curlCommand)
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

func ParseJSON(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return parser.MarshalJSON(reqs)
}

func GenerateJavaScript(req *request.Request) string {
	return jsgen.Generate(req)
}

func ToJavaScript(curlCommand string) (string, error) {
	req, err := ParseRequest(curlCommand)
	if err != nil {
		return "", err
	}
	return GenerateJavaScript(req), nil
}

func ToJavaScriptWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToJavaScript(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}

func GenerateNodeAxios(req *request.Request) string {
	return axgen.Generate(req)
}

func ToNodeAxios(curlCommand string) (string, error) {
	req, err := ParseRequest(curlCommand)
	if err != nil {
		return "", err
	}
	return GenerateNodeAxios(req), nil
}

func ToNodeAxiosWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToNodeAxios(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}

func GenerateGo(req *request.Request) string {
	return gogen.Generate(req)
}

func ToGo(curlCommand string) (string, error) {
	req, err := ParseRequest(curlCommand)
	if err != nil {
		return "", err
	}
	return GenerateGo(req), nil
}

func ToGoWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToGo(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}

func GeneratePython(req *request.Request) string {
	return pygen.Generate(req)
}

func ToPython(curlCommand string) (string, error) {
	if code, err := pygen.GenerateCommand(curlCommand); err == nil {
		return code, nil
	}
	req, err := ParseRequest(curlCommand)
	if err != nil {
		return "", err
	}
	return GeneratePython(req), nil
}

func ToPythonWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToPython(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}
