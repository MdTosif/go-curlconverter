package main

import (
	"fmt"
	"os"
	"strings"

	ansiblegen "github.com/mdtosif/go-curlconverter/pkg/generator/ansible"
	cfmlgen "github.com/mdtosif/go-curlconverter/pkg/generator/cfml"
	gogen "github.com/mdtosif/go-curlconverter/pkg/generator/golang"
	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	axgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodeaxios"
	fetchgen "github.com/mdtosif/go-curlconverter/pkg/generator/nodefetch"
	phprequestsgen "github.com/mdtosif/go-curlconverter/pkg/generator/phprequests"
	pygen "github.com/mdtosif/go-curlconverter/pkg/generator/python"
	rgen "github.com/mdtosif/go-curlconverter/pkg/generator/r"
	rustgen "github.com/mdtosif/go-curlconverter/pkg/generator/rust"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: curlconv [--language javascript|node|node-fetch|node-axios|go|python|ansible|cfml|php-requests|rust|r|parser] \"curl ...\"")
		os.Exit(2)
	}

	language := "javascript"
	args := os.Args[1:]
	if len(args) >= 2 && (args[0] == "--language" || args[0] == "-l") {
		language = args[1]
		args = args[2:]
	}
	if len(args) == 0 {
		fmt.Println("Usage: curlconv [--language javascript|node|node-fetch|node-axios|go|python|ansible|cfml|php-requests|rust|r|parser] \"curl ...\"")
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
	case "node", "node-fetch":
		code := fetchgen.Generate(reqs[0])
		fmt.Print(code)
	case "node-axios":
		code := axgen.Generate(reqs[0])
		fmt.Print(code)
	case "go":
		code := gogen.Generate(reqs[0])
		fmt.Print(code)
	case "python":
		code, err := pygen.GenerateCommand(cmd)
		if err != nil {
			code = pygen.Generate(reqs[0])
		}
		fmt.Print(code)
	case "ansible":
		code := ansiblegen.Generate(reqs[0])
		fmt.Print(code)
	case "cfml":
		code := cfmlgen.Generate(reqs[0])
		fmt.Print(code)
	case "php-requests":
		code, err := phprequestsgen.GenerateCommand(cmd)
		if err != nil {
			code = phprequestsgen.Generate(reqs[0])
		}
		fmt.Print(code)
	case "rust":
		code := rustgen.Generate(reqs[0])
		fmt.Print(code)
	case "r":
		code := rgen.Generate(reqs[0])
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
