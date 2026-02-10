package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestDeleteCmd_WithForce(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", nil, "")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "delete", "foo", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Deleted project") {
		t.Errorf("output = %q", out)
	}
	if store.Exists("foo") {
		t.Error("project should be deleted")
	}
}

func TestDeleteCmd_WithConfirmation(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", nil, "")
	store.Save(p)

	cmd := newDeleteCmd(store, strings.NewReader("y\n"))
	cmd.SetArgs([]string{"foo"})
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.Execute()

	if store.Exists("foo") {
		t.Error("project should be deleted after confirmation")
	}
}

func TestDeleteCmd_Declined(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", nil, "")
	store.Save(p)

	cmd := newDeleteCmd(store, strings.NewReader("n\n"))
	cmd.SetArgs([]string{"foo"})
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.Execute()

	if !store.Exists("foo") {
		t.Error("project should still exist after declining")
	}
	if !strings.Contains(buf.String(), "Aborted") {
		t.Errorf("expected abort message: %q", buf.String())
	}
}

func TestDeleteCmd_NonexistentProject(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	_, err := executeCommand(root, "delete", "nope", "--force")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}
