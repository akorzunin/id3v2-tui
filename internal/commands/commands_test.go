package commands

import (
	"strings"
	"testing"
)

func TestRunCommand(t *testing.T) {
	executor := NewExecutor()

	output, err := executor.Run("echo", "hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !strings.Contains(output, "hello") {
		t.Errorf("expected 'hello' in output, got '%s'", output)
	}
}

func TestRunCommandError(t *testing.T) {
	executor := NewExecutor()

	_, err := executor.Run("nonexistent-command-xyz")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}
