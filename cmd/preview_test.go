package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPreviewCmd_OutputsToStdout(t *testing.T) {
	tmpDir := t.TempDir()

	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	todofile := filepath.Join(tmpDir, "todo.md")
	if _, err := os.Create(todofile); err != nil {
		t.Fatalf("failed to create todofile: %v", err)
	}

	os.Setenv("QAI_DATA_DIR", tmpDir)
	os.Setenv("QAI_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("QAI_DATA_DIR")
	defer os.Unsetenv("QAI_CONFIG_DIR")

	rootCmd.SetArgs([]string{"preview"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output to stdout, but got empty string")
	}

	t.Logf("Output: %s", output)
}

func TestPreviewCmd_PrintsToStdout(t *testing.T) {
	tmpDir := t.TempDir()

	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	todofile := filepath.Join(tmpDir, "todo.md")
	if _, err := os.Create(todofile); err != nil {
		t.Fatalf("failed to create todofile: %v", err)
	}

	os.Setenv("QAI_DATA_DIR", tmpDir)
	os.Setenv("QAI_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("QAI_DATA_DIR")
	defer os.Unsetenv("QAI_CONFIG_DIR")

	buf := new(bytes.Buffer)
	previewCmd.SetOut(buf)

	if err := previewCmd.Execute(); err != nil {
		t.Fatalf("preview command execution failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("preview should output to stdout, got empty string")
	}
}

func TestPreviewCmd_WithID_OutputsToStdout(t *testing.T) {
	tmpDir := t.TempDir()

	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	todofile := filepath.Join(tmpDir, "todo.md")
	if _, err := os.Create(todofile); err != nil {
		t.Fatalf("failed to create todofile: %v", err)
	}

	os.Setenv("QAI_DATA_DIR", tmpDir)
	os.Setenv("QAI_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("QAI_DATA_DIR")
	defer os.Unsetenv("QAI_CONFIG_DIR")

	buf := new(bytes.Buffer)
	previewCmd.SetOut(buf)
	previewCmd.SetArgs([]string{"preview", "1"})

	if err := previewCmd.Execute(); err != nil {
		t.Fatalf("preview command execution failed: %v", err)
	}

	output := buf.String()
	t.Logf("Output with ID: %s", output)
}
