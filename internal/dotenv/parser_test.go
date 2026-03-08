package dotenv

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "bare value",
			input: "KEY=value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "double-quoted value",
			input: `KEY="hello world"`,
			want:  map[string]string{"KEY": "hello world"},
		},
		{
			name:  "single-quoted value",
			input: `KEY='hello world'`,
			want:  map[string]string{"KEY": "hello world"},
		},
		{
			name:  "export prefix",
			input: "export KEY=value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "comment line",
			input: "# this is a comment\nKEY=value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "blank lines",
			input: "\n\nKEY=value\n\n",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "empty value",
			input: "KEY=",
			want:  map[string]string{"KEY": ""},
		},
		{
			name:  "value containing equals",
			input: "URL=postgres://host:5432/db?opt=val",
			want:  map[string]string{"URL": "postgres://host:5432/db?opt=val"},
		},
		{
			name:  "double-quoted with newline escape",
			input: `KEY="line1\nline2"`,
			want:  map[string]string{"KEY": "line1\nline2"},
		},
		{
			name:  "double-quoted with backslash escape",
			input: `KEY="path\\to\\file"`,
			want:  map[string]string{"KEY": "path\\to\\file"},
		},
		{
			name:  "double-quoted with escaped quote",
			input: `KEY="say \"hello\""`,
			want:  map[string]string{"KEY": `say "hello"`},
		},
		{
			name:  "inline comment on bare value",
			input: "KEY=value # this is a comment",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "hash in double-quoted value is literal",
			input: `KEY="value # not a comment"`,
			want:  map[string]string{"KEY": "value # not a comment"},
		},
		{
			name:  "hash in single-quoted value is literal",
			input: `KEY='value # not a comment'`,
			want:  map[string]string{"KEY": "value # not a comment"},
		},
		{
			name:  "multiple variables",
			input: "A=1\nB=2\nC=3",
			want:  map[string]string{"A": "1", "B": "2", "C": "3"},
		},
		{
			name:  "duplicate keys last wins",
			input: "KEY=first\nKEY=second",
			want:  map[string]string{"KEY": "second"},
		},
		{
			name:  "no trailing newline",
			input: "KEY=value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "trailing whitespace on bare value",
			input: "KEY=value   ",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "whitespace around key",
			input: "  KEY  =value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "export with double quotes",
			input: `export KEY="value"`,
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "single-quoted with backslash is literal",
			input: `KEY='no\nescape'`,
			want:  map[string]string{"KEY": `no\nescape`},
		},
		{
			name:    "unterminated double quote",
			input:   `KEY="unterminated`,
			wantErr: true,
		},
		{
			name:    "unterminated single quote",
			input:   `KEY='unterminated`,
			wantErr: true,
		},
		{
			name:    "no equals sign",
			input:   "JUSTKEY",
			wantErr: true,
		},
		{
			name:    "empty key",
			input:   "=value",
			wantErr: true,
		},
		{
			name:  "hash at start of bare value not a comment",
			input: "KEY=#value",
			want:  map[string]string{"KEY": "#value"},
		},
		{
			name:  "empty file",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "only comments",
			input: "# comment 1\n# comment 2",
			want:  map[string]string{},
		},
		{
			name:  "double-quoted empty value",
			input: `KEY=""`,
			want:  map[string]string{"KEY": ""},
		},
		{
			name:  "single-quoted empty value",
			input: `KEY=''`,
			want:  map[string]string{"KEY": ""},
		},
		{
			name:  "tab escape in double quotes",
			input: `KEY="col1\tcol2"`,
			want:  map[string]string{"KEY": "col1\tcol2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d vars, want %d: %v", len(got), len(tt.want), got)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestParseWithDisabled(t *testing.T) {
	input := strings.Join([]string{
		"# API_KEY=disabled",
		"# export TOKEN=\"secret value\"",
		"# just a comment",
		"ACTIVE=yes",
		"# not an assignment = keep ignored",
		"",
	}, "\n")

	got, err := ParseWithDisabled(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWithDisabled() error = %v", err)
	}

	if got.Vars["ACTIVE"] != "yes" {
		t.Fatalf("ACTIVE = %q, want %q", got.Vars["ACTIVE"], "yes")
	}
	if got.DisabledVars["API_KEY"] != "disabled" {
		t.Errorf("API_KEY = %q, want %q", got.DisabledVars["API_KEY"], "disabled")
	}
	if got.DisabledVars["TOKEN"] != "secret value" {
		t.Errorf("TOKEN = %q, want %q", got.DisabledVars["TOKEN"], "secret value")
	}
	if len(got.DisabledVars) != 2 {
		t.Errorf("got %d disabled vars, want 2: %v", len(got.DisabledVars), got.DisabledVars)
	}
}

func TestParse_IgnoresDisabledAssignments(t *testing.T) {
	got, err := Parse(strings.NewReader("# DISABLED=value\nACTIVE=yes\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got["ACTIVE"] != "yes" {
		t.Fatalf("ACTIVE = %q, want %q", got["ACTIVE"], "yes")
	}
	if _, ok := got["DISABLED"]; ok {
		t.Fatalf("DISABLED should not be returned by Parse(): %v", got)
	}
}
