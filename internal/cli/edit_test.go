package cli

import (
	"fmt"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestEditCmd_ProjectNotFound(t *testing.T) {
	store := setupTestStore(t)
	cmd := newEditCmd(store, func(path string) error { return nil })
	_, err := executeCommand(cmd, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if want := `project "nonexistent" not found`; err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestEditCmd_EditorCalledWithCorrectPath(t *testing.T) {
	store := setupTestStore(t)
	p := &config.Project{Name: "my-app", DefaultEnv: "dev"}
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	var calledWith string
	cmd := newEditCmd(store, func(path string) error {
		calledWith = path
		return nil
	})

	_, err := executeCommand(cmd, "my-app")
	if err != nil {
		t.Fatal(err)
	}

	want := store.ProjectPath("my-app")
	if calledWith != want {
		t.Errorf("editor called with %q, want %q", calledWith, want)
	}
}

func TestEditCmd_EditorError(t *testing.T) {
	store := setupTestStore(t)
	p := &config.Project{Name: "my-app", DefaultEnv: "dev"}
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	editorErr := fmt.Errorf("editor crashed")
	cmd := newEditCmd(store, func(path string) error {
		return editorErr
	})

	_, err := executeCommand(cmd, "my-app")
	if err != editorErr {
		t.Errorf("got %v, want %v", err, editorErr)
	}
}
