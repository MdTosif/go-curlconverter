package parser

import "strings"

type Warning [2]string
type Warnings []Warning

func underlineRange(source string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if start > len(source) {
		start = len(source)
	}
	if end > len(source) {
		end = len(source)
	}
	if start == end && end < len(source) {
		end++
	}

	lineStart := 0
	if start > 0 {
		lineStart = strings.LastIndex(source[:start], "\n") + 1
	}
	lineEnd := len(source)
	if idx := strings.Index(source[start:], "\n"); idx != -1 {
		lineEnd = start + idx
	}
	if lineEnd < end {
		end = lineEnd
	}
	if start == end && end < len(source) {
		end++
	}

	line := source[lineStart:lineEnd]
	underline := strings.Repeat(" ", start-lineStart) + strings.Repeat("^", max(1, end-start))
	return line + "\n" + underline
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func scanShellWarnings(source string) Warnings {
	var warnings Warnings
	inSingle, inDouble, esc := false, false, false
	canStartComment := true
	commentLine := false

	for i := 0; i < len(source); i++ {
		ch := source[i]

		if commentLine {
			if ch == '\n' {
				commentLine = false
				canStartComment = true
			}
			continue
		}
		if esc {
			if ch == '\n' || ch == '\r' {
				esc = false
				canStartComment = true
				continue
			}
			esc = false
			canStartComment = false
			continue
		}
		if ch == '\\' {
			if !inSingle {
				if isImproperLineContinuation(source, i) {
					warnings = append(warnings, Warning{
						"unescaped-newline",
						"The trailling '\\' is followed by whitespace, so it won't escape the newline after it:\n" +
							underlineImproperBackslashLine(source, i),
					})
				}
				esc = true
				continue
			}
			canStartComment = false
			continue
		}
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			canStartComment = false
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			canStartComment = false
			continue
		}
		if !inSingle && !inDouble && ch == '#' && canStartComment {
			commentLine = true
			continue
		}
		if !inSingle && !inDouble && (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') {
			canStartComment = true
			continue
		}
		if !inSingle && ch == '`' {
			end := findBacktickEnd(source, i+1)
			warnings = append(warnings, Warning{
				"expansion",
				"found command substitution expression\n" + underlineRange(source, i, end),
			})
			if end > i {
				i = end - 1
			}
			canStartComment = false
			continue
		}
		if !inSingle && ch == '$' && i+1 < len(source) {
			next := source[i+1]
			switch {
			case next == '{':
				end := findClosingDelimiter(source, i+2, '{', '}')
				warnings = append(warnings, Warning{
					"expansion",
					"found expansion expression\n" + underlineRange(source, i, end),
				})
				if end > i {
					i = end - 1
				}
			case next == '(':
				end := findClosingDelimiter(source, i+2, '(', ')')
				warnings = append(warnings, Warning{
					"expansion",
					"found command substitution expression\n" + underlineRange(source, i, end),
				})
				if end > i {
					i = end - 1
				}
			case isSpecialVariableChar(next):
				warnings = append(warnings, Warning{
					"expansion",
					"found environment variable\n" + underlineRange(source, i, i+2),
				})
				warnings = append(warnings, Warning{
					"special_variable_name",
					source[i:i+2] + " is a special Bash variable\n" + underlineRange(source, i+1, i+2),
				})
				i++
			case isVariableStart(next):
				end := i + 2
				for end < len(source) && isVariablePart(source[end]) {
					end++
				}
				warnings = append(warnings, Warning{
					"expansion",
					"found environment variable\n" + underlineRange(source, i, end),
				})
				i = end - 1
			}
			canStartComment = false
			continue
		}
		if !inSingle && !inDouble && ch == '&' {
			if i+1 < len(source) && source[i+1] == '&' {
				// handled as part of the command boundary warnings elsewhere
			} else {
				warnings = append(warnings, Warning{
					"background",
					"found background operator\n" + underlineRange(source, i, i+1),
				})
				canStartComment = false
				continue
			}
		}
		if !inSingle && !inDouble && ch == '|' {
			end := i + 1
			if end < len(source) && source[end] == '|' {
				end++
			}
			warnings = append(warnings, Warning{
				"pipeline",
				"found pipeline operator\n" + underlineRange(source, i, end),
			})
			i = end - 1
			canStartComment = false
			continue
		}
		if !inSingle && !inDouble && (ch == '>' || ch == '<') {
			end := i + 1
			for end < len(source) && source[end] == ch {
				end++
			}
			warnings = append(warnings, Warning{
				"redirect",
				"found shell redirection operator\n" + underlineRange(source, i, end),
			})
			i = end - 1
			canStartComment = false
			continue
		}

		canStartComment = false
	}

	return warnings
}

func findBacktickEnd(source string, start int) int {
	for i := start; i < len(source); i++ {
		if source[i] == '\\' {
			i++
			continue
		}
		if source[i] == '`' {
			return i + 1
		}
	}
	return len(source)
}

func findClosingDelimiter(source string, start int, open, close byte) int {
	depth := 1
	inSingle, inDouble, esc := false, false, false
	for i := start; i < len(source); i++ {
		ch := source[i]
		if esc {
			esc = false
			continue
		}
		if ch == '\\' && !inSingle {
			esc = true
			continue
		}
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
		if inSingle || inDouble {
			continue
		}
		if ch == open {
			depth++
			continue
		}
		if ch == close {
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return len(source)
}

func isSpecialVariableChar(ch byte) bool {
	switch ch {
	case '?', '#', '$', '!', '-', '*', '@':
		return true
	default:
		return false
	}
}

func isVariableStart(ch byte) bool {
	return ch == '_' ||
		(ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z')
}

func isVariablePart(ch byte) bool {
	return isVariableStart(ch) || (ch >= '0' && ch <= '9')
}

func isImproperLineContinuation(source string, idx int) bool {
	if idx < 0 || idx >= len(source) || source[idx] != '\\' {
		return false
	}
	sawWhitespace := false
	for i := idx + 1; i < len(source); i++ {
		switch source[i] {
		case ' ', '\t':
			sawWhitespace = true
		case '\n', '\r':
			return sawWhitespace
		default:
			return false
		}
	}
	return false
}

func underlineImproperBackslashLine(source string, idx int) string {
	lineStart := 0
	if idx > 0 {
		lineStart = strings.LastIndex(source[:idx], "\n") + 1
	}
	lineEnd := len(source)
	if newline := strings.Index(source[idx:], "\n"); newline != -1 {
		lineEnd = idx + newline
	}
	line := source[lineStart:lineEnd]
	end := idx + 1
	for end < len(source) && (source[end] == ' ' || source[end] == '\t') {
		end++
	}
	underline := strings.Repeat(" ", idx-lineStart) + strings.Repeat("^", max(1, end-idx))
	return line + "\n" + underline
}
