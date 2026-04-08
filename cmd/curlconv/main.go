package main

import (
	"fmt"
	"os"
	"strings"

	jsgen "github.com/mdtosif/go-curlconverter/pkg/generator/javascript"
	"github.com/mdtosif/go-curlconverter/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: curlconv \"curl ...\"")
		os.Exit(2)
	}

	cmd := strings.Join(os.Args[1:], " ")
	req, err := parser.Parse(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	code := jsgen.Generate(req)
	fmt.Println(code)
}
