package cli

import "testing"

func TestVersionCmd(t *testing.T) {
	out, err := executeCommand(NewVersionCmd("v1.2.3"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "v1.2.3\n" {
		t.Errorf("output = %q, want %q", out, "v1.2.3\n")
	}
}
