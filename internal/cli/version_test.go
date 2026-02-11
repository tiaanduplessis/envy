package cli

import "testing"

func TestVersionCmd_Full(t *testing.T) {
	info := VersionInfo{
		Version:   "v1.2.3",
		Commit:    "abc1234",
		Date:      "2026-02-10T12:00:00Z",
		GoVersion: "go1.25.7",
	}
	out, err := executeCommand(NewVersionCmd(info))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "envy v1.2.3\ncommit: abc1234\nbuilt:  2026-02-10T12:00:00Z\ngo:     go1.25.7\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestVersionCmd_DevBuild(t *testing.T) {
	info := VersionInfo{Version: "dev"}
	out, err := executeCommand(NewVersionCmd(info))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "envy dev\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}
