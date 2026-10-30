package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kingnathanal/pictl/internal/skill"
	"github.com/spf13/cobra"
)

func TestSkillInstallReportsPartialSuccess(t *testing.T) {
	baseDir := t.TempDir()
	t.Chdir(baseDir)

	// A file blocks the second target deterministically, without relying on permissions.
	if err := os.WriteFile(filepath.Join(baseDir, ".claude"), []byte("blocked"), 0o644); err != nil {
		t.Fatalf("blocking second target: %v", err)
	}
	previousForce := forceSkillInstall
	forceSkillInstall = true
	t.Cleanup(func() { forceSkillInstall = previousForce })

	var output bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&output)
	err := skillInstallCMD.RunE(command, nil)
	if err == nil {
		t.Fatal("install error = nil, want partial installation error")
	}
	if !strings.Contains(err.Error(), "installation incomplete (1 of 3 targets completed)") || !strings.Contains(err.Error(), "pictl skill status") {
		t.Fatalf("install error lacks partial installation details: %v", err)
	}
	firstSkill := filepath.Join(baseDir, ".agents", "skills", "pictl", "SKILL.md")
	if _, err := os.Stat(firstSkill); err != nil {
		t.Fatalf("first target was not installed: %v", err)
	}
	if !strings.Contains(output.String(), "installed "+firstSkill) {
		t.Fatalf("output does not report completed installation: %q", output.String())
	}
	lastSkill := filepath.Join(baseDir, ".codex", "skills", "pictl", "SKILL.md")
	if _, err := os.Stat(lastSkill); !os.IsNotExist(err) {
		t.Fatalf("last target should remain uninstalled, got: %v", err)
	}
}

func TestSkillInstallPreflightConflictMakesNoChanges(t *testing.T) {
	baseDir := t.TempDir()
	t.Chdir(baseDir)
	localSkill := filepath.Join(baseDir, ".claude", "skills", "pictl", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(localSkill), 0o755); err != nil {
		t.Fatalf("creating local skill directory: %v", err)
	}
	if err := os.WriteFile(localSkill, []byte("local changes"), 0o644); err != nil {
		t.Fatalf("writing local skill: %v", err)
	}
	previousForce := forceSkillInstall
	forceSkillInstall = false
	t.Cleanup(func() { forceSkillInstall = previousForce })

	var output bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&output)
	if err := skillInstallCMD.RunE(command, nil); err == nil {
		t.Fatal("install error = nil, want local modification protection")
	}
	if output.Len() != 0 {
		t.Fatalf("preflight conflict reported completed installations: %q", output.String())
	}
	for _, status := range skill.InspectAll(baseDir) {
		if status.State != "not installed" && status.State != "unmanaged" {
			t.Errorf("unexpected state after preflight conflict for %s: %s", status.Target, status.State)
		}
	}
}
