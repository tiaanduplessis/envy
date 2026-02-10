package dotenv

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Format specifies the output format for writing env variables.
type Format string

const (
	FormatDotenv Format = "dotenv"
	FormatExport Format = "export"
)

// WriteOptions controls the output of Write.
type WriteOptions struct {
	Format Format
	Header string
	Sorted bool
}

// Write writes key-value pairs to w in the specified format.
func Write(w io.Writer, vars map[string]string, opts WriteOptions) error {
	if opts.Format == "" {
		opts.Format = FormatDotenv
	}

	if opts.Header != "" {
		for _, line := range strings.Split(opts.Header, "\n") {
			if _, err := fmt.Fprintf(w, "# %s\n", line); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	if opts.Sorted {
		sort.Strings(keys)
	}

	for _, key := range keys {
		value := vars[key]
		quoted := quoteValue(value)

		var line string
		switch opts.Format {
		case FormatExport:
			line = fmt.Sprintf("export %s=%s", key, quoted)
		default:
			line = fmt.Sprintf("%s=%s", key, quoted)
		}

		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

// quoteValue decides whether a value needs quoting and returns the
// appropriately formatted string.
func quoteValue(s string) string {
	if s == "" {
		return `""`
	}

	if needsQuoting(s) {
		return `"` + escapeDoubleQuoted(s) + `"`
	}

	return s
}

func needsQuoting(s string) bool {
	for _, ch := range s {
		switch ch {
		case ' ', '\t', '#', '"', '\'', '\\', '\n', '\r',
			'$', '`', '!', '(', ')', '{', '}', '|', '&', ';', '<', '>':
			return true
		}
	}
	return false
}

func escapeDoubleQuoted(s string) string {
	var b strings.Builder
	for _, ch := range s {
		switch ch {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}
