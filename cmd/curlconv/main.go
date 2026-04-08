package main

import (
	"fmt"
	"os"
	"strings"

	gogen "github.com/mdtosif/go-curlconverter/pkg/generator/golang"
	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	axgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeaxios"
	pygen "github.com/mdtosif/go-curlconverter/pkg/generator/python"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: curlconv [--language javascript|node-axios|go|python|parser] \"curl ...\"")
		os.Exit(2)
	}

	language := "javascript"
	args := os.Args[1:]
	if len(args) >= 2 && (args[0] == "--language" || args[0] == "-l") {
		language = args[1]
		args = args[2:]
	}
	if len(args) == 0 {
		fmt.Println("Usage: curlconv [--language javascript|node-axios|go|python|parser] \"curl ...\"")
		os.Exit(2)
	}

	cmd := strings.Join(args, " ")
	reqs, err := parser.ParseAll(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	switch language {
	case "javascript":
		code := jsgen.Generate(reqs[0])
		fmt.Print(code)
	case "node-axios":
		code := axgen.Generate(reqs[0])
		fmt.Print(code)
	case "go":
		code := gogen.Generate(reqs[0])
		fmt.Print(code)
	case "python":
		code := pygen.Generate(reqs[0])
		fmt.Print(code)
	case "parser":
		jsonOutput, err := parser.MarshalJSON(reqs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "marshal error:", err)
			os.Exit(1)
		}
		fmt.Print(jsonOutput)
	default:
		fmt.Fprintln(os.Stderr, "unsupported language:", language)
		os.Exit(2)
	}
}
