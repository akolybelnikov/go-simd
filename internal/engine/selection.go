package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Go's -bench grammar separates slash components and top-level alternatives,
// except inside parentheses, character classes, or escaped characters. Match
// registered leaf names only: a longer path has no registered descendant.
// Reference: Go 1.27 src/testing/match.go. No testing internals are accessed.
func compileSelection(expression string) (func(string) bool, error) {
	if expression == "" {
		return func(string) bool { return false }, nil
	}
	var branches [][]*regexp.Regexp
	var components []*regexp.Regexp
	start, parens, classes := 0, 0, 0
	appendPart := func(end int) error {
		var normalized strings.Builder
		for _, r := range expression[start:end] {
			switch {
			case unicode.IsSpace(r):
				normalized.WriteByte('_')
			case !strconv.IsPrint(r):
				q := strconv.QuoteRune(r)
				normalized.WriteString(q[1 : len(q)-1])
			default:
				normalized.WriteRune(r)
			}
		}
		compiled, err := regexp.Compile(normalized.String())
		if err != nil {
			return fmt.Errorf("invalid -bench expression: %w", err)
		}
		components = append(components, compiled)
		return nil
	}
	for pos := 0; pos < len(expression); pos++ {
		char := expression[pos]
		if char == '\\' {
			pos++
			continue
		}
		switch char {
		case '[':
			classes++
		case ']':
			classes = max(0, classes-1)
		case '(':
			if classes == 0 {
				parens++
			}
		case ')':
			if classes == 0 {
				parens--
			}
		case '/', '|':
			if classes != 0 || parens != 0 {
				continue
			}
			if err := appendPart(pos); err != nil {
				return nil, err
			}
			if char == '|' {
				branches = append(branches, components)
				components = nil
			}
			start = pos + 1
		}
	}
	if err := appendPart(len(expression)); err != nil {
		return nil, err
	}
	branches = append(branches, components)
	return func(name string) bool {
		parts := strings.Split(name, "/")
		for _, branch := range branches {
			if len(branch) > len(parts) {
				continue
			}
			matched := true
			for i, pattern := range branch {
				if !pattern.MatchString(parts[i]) {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
		return false
	}, nil
}
