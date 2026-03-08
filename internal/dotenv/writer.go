package dotenv

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Format string

const (
	FormatDotenv Format = "dotenv"
	FormatExport Format = "export"
)

type WriteOptions struct {
	Format       Format
	Header       string
	Sorted       bool
	DisabledVars map[string]string
}

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
		if _, err := fmt.Fprintln(w, formatLine(key, value, opts.Format, false)); err != nil {
			return err
		}
	}

	disabledKeys := make([]string, 0, len(opts.DisabledVars))
	for k := range opts.DisabledVars {
		disabledKeys = append(disabledKeys, k)
	}
	if opts.Sorted {
		sort.Strings(disabledKeys)
	}
	if len(disabledKeys) > 0 && (len(keys) > 0 || opts.Header != "") {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	for _, key := range disabledKeys {
		if _, err := fmt.Fprintln(w, formatLine(key, opts.DisabledVars[key], opts.Format, true)); err != nil {
			return err
		}
	}

	return nil
}

func formatLine(key, value string, format Format, disabled bool) string {
	quoted := quoteValue(value)

	var line string
	switch format {
	case FormatExport:
		line = fmt.Sprintf("export %s=%s", key, quoted)
	default:
		line = fmt.Sprintf("%s=%s", key, quoted)
	}

	if disabled {
		return "# " + line
	}

	return line
}

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
