package skill

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedVersionMatchesVersionConstant(t *testing.T) {
	want := []byte("version: \"" + Version + "\"")
	if !bytes.Contains(content, want) {
		t.Fatalf("embedded skill does not contain %q", want)
	}
}

func TestInstallAll(t *testing.T) {
	baseDir := t.TempDir()

	results, err := InstallAll(baseDir, false)
	if err != nil {
		t.Fatalf("InstallAll() error = %v", err)
	}
	if len(results) != len(DefaultTargets) {
		t.Fatalf("InstallAll() installed %d targets, want %d", len(results), len(DefaultTargets))
	}

	for _, result := range results {
		if result.Action != "installed" {
			t.Errorf("first install of %s action = %q, want installed", result.Path, result.Action)
		}
		data, err := os.ReadFile(result.Path)
		if err != nil {
			t.Errorf("reading installed skill %s: %v", result.Path, err)
			continue
		}
		if string(data) != string(content) {
			t.Errorf("installed skill %s does not match embedded content", result.Path)
		}
		manifestPath := filepath.Join(filepath.Dir(result.Path), manifestName)
		if _, err := os.Stat(manifestPath); err != nil {
			t.Errorf("installed manifest %s: %v", manifestPath, err)
		}
	}
}

func TestInstallAllIsIdempotent(t *testing.T) {
	baseDir := t.TempDir()
	if _, err := InstallAll(baseDir, false); err != nil {
		t.Fatalf("first InstallAll() error = %v", err)
	}

	results, err := InstallAll(baseDir, false)
	if err != nil {
		t.Fatalf("second InstallAll() error = %v", err)
	}
	for _, result := range results {
		if result.Action != "unchanged" {
			t.Errorf("second install of %s action = %q, want unchanged", result.Path, result.Action)
		}
	}
}

func TestInstallAllRequiresForceToReplaceChanges(t *testing.T) {
	baseDir := t.TempDir()
	path := filepath.Join(baseDir, DefaultTargets[1], "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating target directory: %v", err)
	}
	if err := os.WriteFile(path, []byte("local changes"), 0o644); err != nil {
		t.Fatalf("writing modified skill: %v", err)
	}

	if _, err := InstallAll(baseDir, false); err == nil {
		t.Fatal("InstallAll() error = nil, want overwrite protection error")
	}
	firstTarget := filepath.Join(baseDir, DefaultTargets[0], "SKILL.md")
	if _, err := os.Stat(firstTarget); !os.IsNotExist(err) {
		t.Fatalf("partial install created %s before detecting conflict", firstTarget)
	}

	if _, err := InstallAll(baseDir, true); err != nil {
		t.Fatalf("InstallAll() with force error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading replaced skill: %v", err)
	}
	if string(data) != string(content) {
		t.Fatal("forced install did not restore embedded skill content")
	}
}

func TestInstallAllProtectsDanglingSymlinks(t *testing.T) {
	for _, filename := range []string{"SKILL.md", manifestName} {
		t.Run(filename, func(t *testing.T) {
			baseDir := t.TempDir()
			dir := filepath.Join(baseDir, DefaultTargets[1])
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("creating skill directory: %v", err)
			}
			path := filepath.Join(dir, filename)
			destination := filepath.Join(baseDir, "missing", filename)
			if err := os.Symlink(destination, path); err != nil {
				t.Fatalf("creating dangling symlink: %v", err)
			}

			if status := InspectAll(baseDir)[1]; status.State != "incomplete" {
				t.Errorf("dangling symlink state = %q, want incomplete", status.State)
			}
			results, err := InstallAll(baseDir, false)
			if err == nil || !strings.Contains(err.Error(), "--force") {
				t.Fatalf("InstallAll() error = %v, want explicit replacement requirement", err)
			}
			if len(results) != 0 {
				t.Fatalf("InstallAll() completed %d targets before detecting symlink", len(results))
			}
			link, err := os.Readlink(path)
			if err != nil || link != destination {
				t.Fatalf("dangling symlink changed: destination = %q, error = %v", link, err)
			}
			if _, err := os.Lstat(filepath.Join(baseDir, DefaultTargets[0])); !os.IsNotExist(err) {
				t.Fatalf("first target should remain absent after preflight failure: %v", err)
			}

			results, err = InstallAll(baseDir, true)
			if err != nil {
				t.Fatalf("forced InstallAll() error = %v", err)
			}
			if results[1].Action != "replaced" {
				t.Errorf("forced install action = %q, want replaced", results[1].Action)
			}
			info, err := os.Lstat(path)
			if err != nil {
				t.Fatalf("checking replaced symlink: %v", err)
			}
			if !info.Mode().IsRegular() {
				t.Errorf("forced replacement mode = %v, want regular file", info.Mode())
			}
			if status := InspectAll(baseDir)[1]; status.State != "current" {
				t.Errorf("forced installation state = %q, want current", status.State)
			}
			if _, err := os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatalf("forced replacement should leave missing destination untouched: %v", err)
			}
		})
	}
}

func TestInstallAllPreservesCurrentSymlinks(t *testing.T) {
	for _, filename := range []string{"SKILL.md", manifestName} {
		t.Run(filename, func(t *testing.T) {
			baseDir := t.TempDir()
			target := DefaultTargets[1]
			writeManagedSkill(t, baseDir, target, Version, content)
			path := filepath.Join(baseDir, target, filename)
			destination := filepath.Join(baseDir, filename)
			if err := os.Rename(path, destination); err != nil {
				t.Fatalf("moving managed file to symlink destination: %v", err)
			}
			if err := os.Symlink(destination, path); err != nil {
				t.Fatalf("creating current symlink: %v", err)
			}

			results, err := InstallAll(baseDir, false)
			if err != nil {
				t.Fatalf("InstallAll() error = %v", err)
			}
			if results[1].Action != "unchanged" {
				t.Errorf("current symlink action = %q, want unchanged", results[1].Action)
			}
			link, err := os.Readlink(path)
			if err != nil || link != destination {
				t.Fatalf("current symlink changed: destination = %q, error = %v", link, err)
			}
		})
	}
}

func TestInstallAllAutomaticallyUpgradesManagedSkill(t *testing.T) {
	baseDir := t.TempDir()
	target := DefaultTargets[0]
	oldContent := []byte("old managed skill")
	writeManagedSkill(t, baseDir, target, "0.0.1", oldContent)

	results, err := InstallAll(baseDir, false)
	if err != nil {
		t.Fatalf("InstallAll() error = %v", err)
	}
	if results[0].Action != "updated" {
		t.Fatalf("upgrade action = %q, want updated", results[0].Action)
	}
	data, err := os.ReadFile(results[0].Path)
	if err != nil {
		t.Fatalf("reading upgraded skill: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatal("upgraded skill does not match embedded content")
	}
}

func TestInstallAllAutomaticallyUpgradesLegacySkillWithoutManifest(t *testing.T) {
	baseDir := t.TempDir()
	target := DefaultTargets[0]
	legacyContent := legacyVersion010Content(t)
	dir := filepath.Join(baseDir, target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating legacy skill directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), legacyContent, 0o644); err != nil {
		t.Fatalf("writing legacy skill: %v", err)
	}

	results, err := InstallAll(baseDir, false)
	if err != nil {
		t.Fatalf("InstallAll() error = %v", err)
	}
	if results[0].Action != "updated" {
		t.Fatalf("legacy upgrade action = %q, want updated", results[0].Action)
	}
}

func TestInstallAllProtectsModifiedManagedSkill(t *testing.T) {
	baseDir := t.TempDir()
	target := DefaultTargets[1]
	writeManagedSkill(t, baseDir, target, Version, content)
	skillPath := filepath.Join(baseDir, target, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("local changes"), 0o644); err != nil {
		t.Fatalf("modifying managed skill: %v", err)
	}

	if _, err := InstallAll(baseDir, false); err == nil {
		t.Fatal("InstallAll() error = nil, want modified skill error")
	}
	firstTarget := filepath.Join(baseDir, DefaultTargets[0], "SKILL.md")
	if _, err := os.Stat(firstTarget); !os.IsNotExist(err) {
		t.Fatalf("partial install created %s before detecting modification", firstTarget)
	}
}

func TestInspectAllReportsCurrentAndUpgrade(t *testing.T) {
	baseDir := t.TempDir()
	writeManagedSkill(t, baseDir, DefaultTargets[0], Version, content)
	writeManagedSkill(t, baseDir, DefaultTargets[1], "0.0.1", []byte("old skill"))

	results := InspectAll(baseDir)
	if results[0].State != "current" {
		t.Errorf("current target state = %q, want current", results[0].State)
	}
	if results[1].State != "upgrade available" {
		t.Errorf("old target state = %q, want upgrade available", results[1].State)
	}
	if results[2].State != "not installed" {
		t.Errorf("missing target state = %q, want not installed", results[2].State)
	}
}

func TestInstallAllBlocksDowngrade(t *testing.T) {
	baseDir := t.TempDir()
	writeManagedSkill(t, baseDir, DefaultTargets[0], "99.0.0", []byte("future skill"))

	if _, err := InstallAll(baseDir, false); err == nil {
		t.Fatal("InstallAll() error = nil, want downgrade protection error")
	}
}

func writeManagedSkill(t *testing.T, baseDir, target, version string, skillData []byte) {
	t.Helper()
	dir := filepath.Join(baseDir, target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating managed skill directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), skillData, 0o644); err != nil {
		t.Fatalf("writing managed skill: %v", err)
	}
	manifest, err := json.Marshal(Manifest{Version: version, SHA256: checksum(skillData)})
	if err != nil {
		t.Fatalf("marshalling manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifestName), manifest, 0o644); err != nil {
		t.Fatalf("writing managed skill manifest: %v", err)
	}
}

func legacyVersion010Content(t *testing.T) []byte {
	t.Helper()
	legacy, err := os.ReadFile(filepath.Join("testdata", "skill-v0.1.0.md"))
	if err != nil {
		t.Fatalf("reading legacy v0.1.0 fixture: %v", err)
	}
	if checksum(legacy) != "68a342cc5af9d3b0e05c739c1f8b21c0d622ca38bbcd06051125e550dc6a177f" {
		t.Fatal("legacy v0.1.0 fixture checksum changed")
	}
	return legacy
}
