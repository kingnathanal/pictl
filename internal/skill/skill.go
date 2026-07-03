package skill

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	Version      = "0.2.0"
	manifestName = ".pictl-skill.json"
)

var DefaultTargets = []string{
	filepath.Join(".agents", "skills", "pictl"),
	filepath.Join(".claude", "skills", "pictl"),
	filepath.Join(".codex", "skills", "pictl"),
}

var legacyManagedChecksums = map[string]string{
	"68a342cc5af9d3b0e05c739c1f8b21c0d622ca38bbcd06051125e550dc6a177f": "0.1.0",
}

//go:embed templates/pictl/SKILL.md
var content []byte

type Manifest struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

type InstallResult struct {
	Path   string
	Action string
}

type StatusResult struct {
	Target           string
	Path             string
	State            string
	InstalledVersion string
	AvailableVersion string
	Detail           string
}

type installPlan struct {
	target        string
	skillPath     string
	manifestPath  string
	action        string
	writeSkill    bool
	writeManifest bool
}

func InstallAll(baseDir string, force bool) ([]InstallResult, error) {
	plans := make([]installPlan, 0, len(DefaultTargets))
	for _, target := range DefaultTargets {
		plan, err := planInstall(baseDir, target, force)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	results := make([]InstallResult, 0, len(plans))
	for _, plan := range plans {
		if err := applyPlan(plan); err != nil {
			return results, err
		}
		results = append(results, InstallResult{Path: plan.skillPath, Action: plan.action})
	}
	return results, nil
}

func InspectAll(baseDir string) []StatusResult {
	results := make([]StatusResult, 0, len(DefaultTargets))
	for _, target := range DefaultTargets {
		results = append(results, inspect(baseDir, target))
	}
	return results
}

func planInstall(baseDir, target string, force bool) (installPlan, error) {
	dir := filepath.Join(baseDir, target)
	plan := installPlan{
		target:       target,
		skillPath:    filepath.Join(dir, "SKILL.md"),
		manifestPath: filepath.Join(dir, manifestName),
	}
	status := inspect(baseDir, target)

	switch status.State {
	case "not installed":
		plan.action = "installed"
		plan.writeSkill = true
		plan.writeManifest = true
	case "current":
		plan.action = "unchanged"
	case "unmanaged current":
		plan.action = "managed"
		plan.writeManifest = true
	case "upgrade available":
		plan.action = "updated"
		plan.writeSkill = true
		plan.writeManifest = true
	case "modified", "unmanaged", "incomplete", "invalid manifest", "newer", "version conflict":
		if !force {
			return installPlan{}, fmt.Errorf("%s: %s; use --force to replace it", plan.skillPath, status.Detail)
		}
		plan.action = "replaced"
		plan.writeSkill = true
		plan.writeManifest = true
	default:
		return installPlan{}, fmt.Errorf("%s: unknown skill state %q", plan.skillPath, status.State)
	}

	return plan, nil
}

func inspect(baseDir, target string) StatusResult {
	dir := filepath.Join(baseDir, target)
	skillPath := filepath.Join(dir, "SKILL.md")
	manifestPath := filepath.Join(dir, manifestName)
	result := StatusResult{
		Target:           target,
		Path:             skillPath,
		AvailableVersion: Version,
	}

	skillData, skillExists, skillErr := readOptional(skillPath)
	if skillErr != nil {
		result.State = "incomplete"
		result.Detail = skillErr.Error()
		return result
	}
	manifestData, manifestExists, manifestErr := readOptional(manifestPath)
	if manifestErr != nil {
		result.State = "incomplete"
		result.Detail = manifestErr.Error()
		return result
	}

	switch {
	case !skillExists && !manifestExists:
		result.State = "not installed"
		result.Detail = "skill and manifest are missing"
		return result
	case skillExists && !manifestExists:
		if bytes.Equal(skillData, content) {
			result.State = "unmanaged current"
			result.InstalledVersion = Version
			result.Detail = "current skill has no management manifest"
		} else if legacyVersion, ok := legacyManagedChecksums[checksum(skillData)]; ok {
			result.State = "upgrade available"
			result.InstalledVersion = legacyVersion
			result.Detail = fmt.Sprintf("upgrade legacy managed skill %s to %s", legacyVersion, Version)
		} else {
			result.State = "unmanaged"
			result.Detail = "skill has no management manifest"
		}
		return result
	case !skillExists && manifestExists:
		result.State = "incomplete"
		result.Detail = "manifest exists but SKILL.md is missing"
		return result
	}

	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		result.State = "invalid manifest"
		result.Detail = fmt.Sprintf("cannot parse %s: %v", manifestPath, err)
		return result
	}
	result.InstalledVersion = manifest.Version
	if manifest.Version == "" || manifest.SHA256 == "" {
		result.State = "invalid manifest"
		result.Detail = "manifest requires version and sha256"
		return result
	}
	if checksum(skillData) != manifest.SHA256 {
		result.State = "modified"
		result.Detail = "SKILL.md differs from the managed checksum"
		return result
	}

	comparison, err := compareVersions(manifest.Version, Version)
	if err != nil {
		result.State = "invalid manifest"
		result.Detail = err.Error()
		return result
	}
	switch {
	case comparison < 0:
		result.State = "upgrade available"
		result.Detail = fmt.Sprintf("upgrade %s to %s", manifest.Version, Version)
	case comparison > 0:
		result.State = "newer"
		result.Detail = fmt.Sprintf("installed version %s is newer than bundled version %s", manifest.Version, Version)
	case manifest.SHA256 != checksum(content):
		result.State = "version conflict"
		result.Detail = fmt.Sprintf("version %s has different managed content; the bundled skill version must be incremented", Version)
	default:
		result.State = "current"
		result.Detail = "managed skill is current"
	}
	return result
}

func applyPlan(plan installPlan) error {
	if !plan.writeSkill && !plan.writeManifest {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(plan.skillPath), 0o755); err != nil {
		return fmt.Errorf("creating skill directory for %s: %w", plan.skillPath, err)
	}
	if plan.writeSkill {
		if err := writeAtomic(plan.skillPath, content); err != nil {
			return fmt.Errorf("installing skill at %s: %w", plan.skillPath, err)
		}
	}
	if plan.writeManifest {
		data, err := manifestData()
		if err != nil {
			return fmt.Errorf("creating skill manifest for %s: %w", plan.skillPath, err)
		}
		if err := writeAtomic(plan.manifestPath, data); err != nil {
			return fmt.Errorf("installing skill manifest at %s: %w", plan.manifestPath, err)
		}
	}
	return nil
}

func manifestData() ([]byte, error) {
	data, err := json.MarshalIndent(Manifest{Version: Version, SHA256: checksum(content)}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func writeAtomic(path string, data []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".pictl-skill-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func readOptional(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		return data, true, nil
	case errors.Is(err, os.ErrNotExist):
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("reading %s: %w", path, err)
	}
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func compareVersions(left, right string) (int, error) {
	leftParts, err := parseVersion(left)
	if err != nil {
		return 0, fmt.Errorf("invalid installed skill version %q: %w", left, err)
	}
	rightParts, err := parseVersion(right)
	if err != nil {
		return 0, fmt.Errorf("invalid bundled skill version %q: %w", right, err)
	}
	for i := range leftParts {
		if leftParts[i] < rightParts[i] {
			return -1, nil
		}
		if leftParts[i] > rightParts[i] {
			return 1, nil
		}
	}
	return 0, nil
}

func parseVersion(version string) ([3]int, error) {
	var parsed [3]int
	parts := strings.Split(version, ".")
	if len(parts) != len(parsed) {
		return parsed, fmt.Errorf("expected major.minor.patch")
	}
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return parsed, fmt.Errorf("version component %q is not a non-negative integer", part)
		}
		parsed[i] = value
	}
	return parsed, nil
}
