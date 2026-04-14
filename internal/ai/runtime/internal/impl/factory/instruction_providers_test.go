package factory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectInstructionProviderResolveFromAgentsFile(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "child")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := "# AGENTS.md\n\nfrom temp file"
	if err := os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(nestedDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	got := NewProjectInstructionProvider().Resolve()
	if !strings.Contains(got, "from temp file") {
		t.Fatalf("expected provider to load AGENTS.md from file, got: %s", got)
	}
}

func TestProjectInstructionProviderFallbacksToDefault(t *testing.T) {
	tmpDir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	got := NewProjectInstructionProvider().Resolve()
	if !strings.Contains(got, "本文件定义本项目内 AI Agent 的强制开发规则") {
		t.Fatalf("expected fallback project instruction, got: %s", got)
	}
}
