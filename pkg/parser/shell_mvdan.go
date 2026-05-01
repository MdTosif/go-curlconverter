package parser

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
	"mvdan.cc/sh/v3/syntax"
)

func scanShellWarningsAST(source string) (Warnings, error) {
	parser := syntax.NewParser()
	file, err := parser.Parse(strings.NewReader(source), "")
	if err != nil {
		return nil, err
	}

	var warnings Warnings
	syntax.Walk(file, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.CmdSubst:
			warnings = append(warnings, Warning{
				"expansion",
				"found command substitution expression\n" + underlineNodeRange(source, n),
			})
			return false
		case *syntax.ParamExp:
			text := printSyntaxNode(n)
			if n.Short {
				warnings = append(warnings, Warning{
					"expansion",
					"found environment variable\n" + underlineNodeRange(source, n),
				})
				if len(text) == 2 && strings.HasPrefix(text, "$") && isSpecialVariableChar(text[1]) {
					warnings = append(warnings, Warning{
						"special_variable_name",
						text + " is a special Bash variable\n" + underlineRange(source, posOffset(n.Pos())+1, posOffset(n.End())),
					})
				}
				return true
			}
			warnings = append(warnings, Warning{
				"expansion",
				"found expansion expression\n" + underlineNodeRange(source, n),
			})
		case *syntax.BinaryCmd:
			op := n.Op.String()
			if op == "|" || op == "||" {
				start := posOffset(n.OpPos)
				warnings = append(warnings, Warning{
					"pipeline",
					"found pipeline operator\n" + underlineRange(source, start, start+len(op)),
				})
			}
		case *syntax.Stmt:
			if n.Background {
				start := posOffset(n.Semicolon)
				warnings = append(warnings, Warning{
					"background",
					"found background operator\n" + underlineRange(source, start, start+1),
				})
			}
			for _, redir := range n.Redirs {
				op := redir.Op.String()
				start := posOffset(redir.OpPos)
				warnings = append(warnings, Warning{
					"redirect",
					"found shell redirection operator\n" + underlineRange(source, start, start+len(op)),
				})
			}
		}
		return true
	})
	return warnings, nil
}

func extractRawCommands(source string) ([][]string, error) {
	parser := syntax.NewParser()
	file, err := parser.Parse(strings.NewReader(source), "")
	if err != nil {
		return nil, err
	}

	commands := make([][]string, 0)
	syntax.Walk(file, func(node syntax.Node) bool {
		// Command substitutions are shell expressions, not top-level command chains.
		// Skip their subtree so nested calls are not treated as standalone commands.
		if _, ok := node.(*syntax.CmdSubst); ok {
			return false
		}
		call, ok := node.(*syntax.CallExpr)
		if !ok {
			return true
		}
		toks := wordsToStrings(call.Args)
		if len(toks) > 0 {
			commands = append(commands, toks)
		}
		return true
	})
	return commands, nil
}

func parseExtractedCommands(commands [][]string, warnings Warnings) ([]*request.Request, Warnings, error) {
	commands = mergeOptionOnlyCommands(commands)
	curlCommands := make([][]string, 0, len(commands))
	for _, command := range commands {
		if len(command) == 0 {
			continue
		}
		if strings.TrimSpace(command[0]) == "curl" {
			curlCommands = append(curlCommands, command)
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(command[0]), "-") {
			continue
		}
		warnings = append(warnings, Warning{
			"ignored-command",
			"ignoring non-curl command starting with " + strconv.Quote(clip(command[0])),
		})
	}
	if len(curlCommands) == 0 {
		return nil, warnings, errors.New("command must start with 'curl'")
	}

	reqs := make([]*request.Request, 0, len(curlCommands))
	for _, command := range curlCommands {
		parsed, err := parseTokenArgs(command)
		if err != nil {
			return nil, warnings, err
		}
		reqs = append(reqs, parsed...)
	}
	return reqs, warnings, nil
}

func wordsToStrings(words []*syntax.Word) []string {
	out := make([]string, 0, len(words))
	for _, word := range words {
		token := wordToString(word)
		if token == "" && len(word.Parts) == 0 {
			continue
		}
		out = append(out, token)
	}
	return out
}

func wordToString(word *syntax.Word) string {
	if word == nil {
		return ""
	}
	var b strings.Builder
	for _, part := range word.Parts {
		b.WriteString(wordPartToString(part, false))
	}
	return b.String()
}

func wordPartToString(part syntax.WordPart, inDoubleQuote bool) string {
	switch p := part.(type) {
	case *syntax.Lit:
		return p.Value
	case *syntax.SglQuoted:
		return p.Value
	case *syntax.DblQuoted:
		var b strings.Builder
		for _, sub := range p.Parts {
			b.WriteString(wordPartToString(sub, true))
		}
		return b.String()
	case *syntax.ParamExp:
		return printSyntaxNode(p)
	case *syntax.CmdSubst:
		return printSyntaxNode(p)
	case *syntax.ArithmExp:
		return printSyntaxNode(p)
	case *syntax.ProcSubst:
		return printSyntaxNode(p)
	case *syntax.ExtGlob:
		return printSyntaxNode(p)
	default:
		if inDoubleQuote {
			return printSyntaxNode(part)
		}
		return strings.TrimSpace(printSyntaxNode(part))
	}
}

func printSyntaxNode(node syntax.Node) string {
	var b strings.Builder
	printer := syntax.NewPrinter()
	if err := printer.Print(&b, node); err != nil {
		return ""
	}
	return b.String()
}

func mergeOptionOnlyCommands(commands [][]string) [][]string {
	if len(commands) == 0 {
		return commands
	}
	merged := make([][]string, 0, len(commands))
	for _, cmd := range commands {
		if len(cmd) > 0 &&
			strings.HasPrefix(cmd[0], "-") &&
			len(merged) > 0 &&
			len(merged[len(merged)-1]) > 0 &&
			merged[len(merged)-1][0] == "curl" {
			merged[len(merged)-1] = append(merged[len(merged)-1], cmd...)
			continue
		}
		merged = append(merged, cmd)
	}
	return merged
}

func posOffset(pos syntax.Pos) int {
	if !pos.IsValid() {
		return 0
	}
	return int(pos.Offset())
}

func underlineNodeRange(source string, node syntax.Node) string {
	return underlineRange(source, posOffset(node.Pos()), posOffset(node.End()))
}
