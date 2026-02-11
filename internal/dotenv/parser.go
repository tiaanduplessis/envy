package dotenv

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Parse reads a .env file and returns a map of key-value pairs.
// It handles bare values, double-quoted values (with escape sequences),
// single-quoted values (literal), export prefixes, comments, and blank lines.
// Duplicate keys: last value wins.
func Parse(r io.Reader) (map[string]string, error) {
	vars := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		key, rest, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE, got %q", lineNum, line)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}

		value, err := parseValue(rest)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		vars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return vars, nil
}

func parseValue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	if strings.HasPrefix(raw, "\"") {
		return parseDoubleQuoted(raw[1:])
	}

	// Single-quoted value (literal, no escapes)
	if strings.HasPrefix(raw, "'") {
		end := strings.LastIndex(raw[1:], "'")
		if end == -1 {
			return "", fmt.Errorf("unterminated single quote")
		}
		return raw[1 : end+1], nil
	}

	value := stripInlineComment(raw)
	return strings.TrimRight(value, " \t"), nil
}

func parseDoubleQuoted(s string) (string, error) {
	var b strings.Builder
	i := 0
	for i < len(s) {
		ch := s[i]
		if ch == '"' {
			// End of quoted value — ignore anything after closing quote
			return b.String(), nil
		}
		if ch == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
			}
			i += 2
			continue
		}
		b.WriteByte(ch)
		i++
	}
	return "", fmt.Errorf("unterminated double quote")
}

func stripInlineComment(s string) string {
	// For bare values, # preceded by whitespace is a comment.
	// We scan for " #" or "\t#".
	for i := 0; i < len(s); i++ {
		if s[i] == '#' && i > 0 && (s[i-1] == ' ' || s[i-1] == '\t') {
			return s[:i-1]
		}
	}
	return s
}
