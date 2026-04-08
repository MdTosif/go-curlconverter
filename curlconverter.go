package curlconverter

import (
	gogen "github.com/mdtosif/go-curlconverter/pkg/generator/golang"
	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	axgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeaxios"
	pygen "github.com/mdtosif/go-curlconverter/pkg/generator/python"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type Warning [2]string
type Warnings []Warning

func Parse(curlCommand string) ([]*request.Request, error) {
	return parser.ParseAll(curlCommand)
}

func ParseJSON(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return parser.MarshalJSON(reqs)
}

func ToJavaScript(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return jsgen.Generate(reqs[0]), nil
}

func ToJavaScriptWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToJavaScript(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}

func ToNodeAxios(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return axgen.Generate(reqs[0]), nil
}

func ToNodeAxiosWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToNodeAxios(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}

func ToGo(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return gogen.Generate(reqs[0]), nil
}

func ToGoWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToGo(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}

func ToPython(curlCommand string) (string, error) {
	reqs, err := parser.ParseAll(curlCommand)
	if err != nil {
		return "", err
	}
	return pygen.Generate(reqs[0]), nil
}

func ToPythonWarn(curlCommand string) (string, Warnings, error) {
	code, err := ToPython(curlCommand)
	if err != nil {
		return "", nil, err
	}
	return code, nil, nil
}
