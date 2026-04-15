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

func scanImproperLineContinuationWarnings(source string) Warnings {
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
		canStartComment = false
	}
	return warnings
}

func isSpecialVariableChar(ch byte) bool {
	switch ch {
	case '?', '#', '$', '!', '-', '*', '@':
		return true
	default:
		return false
	}
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
