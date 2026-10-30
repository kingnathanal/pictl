# AGENTS.md — pictl

This file describes the project purpose, architecture, conventions, and roadmap for AI agents and contributors working on this codebase.

---

## What This Project Is

`pictl` is a CLI tool written in Go for managing a fleet of Raspberry Pis over SSH. It connects to each Pi using key-based auth, runs commands remotely, and presents results in clean terminal tables.

The core project should stay useful for any Raspberry Pi running Raspberry Pi OS or Debian with SSH enabled. K3s support may exist as an optional module later, but Kubernetes is not the main focus of the tool.

It is also a structured Go learning project — each milestone introduces new Go concepts intentionally.

---

## Learning-First Collaboration

This project is a Go learning pad. When helping the owner work on `pictl`, default to guiding them through the Go implementation instead of writing the code directly.

### Default behavior
- Explain the relevant Go concepts before proposing code
- Point to the files and functions that should change
- Describe the implementation steps in small, practical chunks
- Provide focused examples or pseudocode when useful
- Encourage the owner to make the actual code changes

### When code changes are allowed
Only edit or create Go code when the owner explicitly asks the agent to implement the change, fix a bug directly, or generate code.

Documentation-only updates are acceptable when requested, as with this file and the README.

### Teaching style
- Keep explanations tied to the current codebase
- Prefer simple, idiomatic Go over clever abstractions
- Explain why a pattern is used, especially for goroutines, channels, errors, structs, interfaces, and package boundaries
- Call out tradeoffs when there are multiple reasonable approaches
- Avoid skipping over syntax that is likely to be new to a Go learner

---

## Fleet Topology

| Node | Role | Interface |
|---|---|---|
| pi00-wifi | Worker | WiFi |
| pi01-wifi | Worker | WiFi |
| pi02-wifi | Worker | WiFi |
| pi03-lan | Worker | LAN |
| pi04-lan | Worker | LAN |
| pi05-lan | Worker | LAN |

- **OS:** Debian GNU/Linux 13 (trixie) — ARM64
- **SSH User:** `kingnathanal`
- **Auth:** SSH key only — no password auth
- **SSH Key Path:** configured in `cluster.yaml` via `ssh_key_path`
- **K3s:** optional future module, not required for core fleet management

---

## Project Structure

```
pictl/
├── cmd/                    # Cobra CLI commands — thin, no business logic
│   ├── root.go             # Root command + Execute()
│   ├── ping.go             # pictl ping
│   ├── info.go             # pictl info
│   ├── update.go           # pictl update
│   ├── exec.go             # pictl exec
│   ├── check.go            # pictl check
│   └── skill.go            # pictl skill install
├── internal/               # All business logic — not importable outside module
│   ├── config/
│   │   └── config.go       # cluster.yaml loader, Config + Node structs
│   ├── ssh/
│   │   ├── common.go       # NewClient(), RunCommand() — shared SSH primitives
│   │   ├── ping.go         # TCP reachability check, PingAll(), printTable()
│   │   ├── stats.go        # Node metrics collection, CollectAll(), PrintStatsTable()
│   │   └── update.go       # Apt update logic, UpdateAll(), PrintUpdateResults()
│   └── skill/
│       ├── skill.go        # Embedded skill content and installer
│       └── templates/pictl/SKILL.md
├── cluster.yaml            # Node definitions — source of truth for the fleet
├── go.mod
├── go.sum
└── main.go                 # Entry point — calls cmd.Execute()
```

---

## Architecture Decisions

### CMD layer is always thin
`cmd/` files load config, call into `internal/`, and print results. No SSH logic, no parsing, no business decisions live in `cmd/`.

### Exported vs unexported
- **Exported (capital):** functions called from `cmd/` or other packages
- **Unexported (lowercase):** internal helpers — parsers, single-node workers, formatters

### Concurrency model
- **Read-only commands** (`ping`, `info`): run concurrently via goroutines + channels
- **Write commands** (`update`, future `install`): run sequentially to avoid system lock conflicts (e.g. dpkg lock)
- Channel pattern used for result collection — not shared slice + WaitGroup

### SSH connection model
- `NewClient()` and `RunCommand()` live in `internal/ssh/common.go`
- Every command that needs SSH calls `NewClient()` per node, defers `client.Close()`
- `RunCommand()` returns `(string, error)` — output is always captured even on non-zero exit

### Import aliases
- Only used when a name collision exists (e.g. `internalssh` when both `internal/ssh` and `golang.org/x/crypto/ssh` are imported in the same file)
- Never aliased gratuitously

### Config is file-driven
- `cluster.yaml` is the single source of truth for node IPs, names, roles, SSH user, and key path
- Loaded at command runtime via `config.LoadConfig()`
- `cluster.yaml` should be `.gitignore`'d if sensitive details are added in the future

### AI skill maintenance
- The bundled user-facing skill lives at `internal/skill/templates/pictl/SKILL.md`
- `internal/skill.Version` is the authoritative skill version and must match the version in the skill frontmatter
- Installed skills are tracked by `.pictl-skill.json`, which records the managed version and SHA-256 checksum used for safe upgrades
- When a command, flag, targeting rule, configuration behavior, or safety requirement changes, review and update the bundled skill in the same change
- Increment the skill version for every skill content change: patch for corrections, minor for new commands or workflows, and major for incompatible behavior
- Keep the skill focused on using `pictl`; repository development guidance belongs in this `AGENTS.md` file

---

## Coding Conventions

| Convention | Rule |
|---|---|
| Error handling | Always check errors — never `_` an error silently |
| Error wrapping | Use `fmt.Errorf("context: %w", err)` to preserve error chain |
| SSH sessions | Always `defer session.Close()` and `defer client.Close()` |
| Apt commands | Always prefix with `sudo DEBIAN_FRONTEND=noninteractive` to suppress interactive prompts |
| Output | Print tables with fixed-width `%-Ns` formatting for alignment |
| Struct field visibility | Exported fields in structs that cross package boundaries, unexported otherwise |
| File naming | Descriptive and flat — `common.go`, `stats.go`, `update.go`. No `helpers.go` or `utils.go` |

---

## Commands

### `pictl ping`
- Checks TCP reachability on port 22 for all nodes
- Runs concurrently via goroutines
- Output: node name, IP, status (UP/DOWN), latency

### `pictl info`
- SSHs into each node and collects: CPU usage %, CPU temperature, memory used/total, disk used/total/percent, hostname, OS
- CPU usage measured via two `/proc/stat` snapshots in a single SSH session with `sleep 1` between them
- Runs concurrently via goroutines + channels
- Completes in ~2 seconds for the full fleet

### `pictl update`
- Runs `sudo DEBIAN_FRONTEND=noninteractive apt update && apt upgrade -y` on each node
- Runs **sequentially** to avoid dpkg lock conflicts
- Output: node name, IP, status (OK/FAILED), apt summary line

---

## Environment Notes

- Pis run Debian trixie (testing) — some packages like `rpi-chromium-mods` may have config file conflicts during upgrade; `DEBIAN_FRONTEND=noninteractive` handles this automatically
- `sudo` is passwordless for the SSH user on all nodes
- `HostKeyCallback` is set to `ssh.InsecureIgnoreHostKey()` — acceptable for a trusted home lab
- Nodes are on a local `192.168.1.x` network — not reachable from the internet

---

## Roadmap

The canonical product roadmap and release acceptance checklist live in [Readme.md](Readme.md#roadmap). Keep milestone status there so contributor guidance and user-facing release scope do not diverge.

The v0.1.0 MVP promise is: **Connect your existing Raspberry Pis and see their health from one terminal.** Existing operational commands are a foundation; they do not establish public-release readiness.

Implement the remaining MVP milestones in this order:

1. **First Pi onboarding:** initialize configuration outside the application checkout, add / list / update / remove nodes safely, verify authenticated access, and document the first successful health check.
2. **Core reliability and safe operation:** fix complete role selection, define combined targeting and empty-selection errors, validate configuration, verify host keys, bound waits, preserve failure output, report meaningful exit statuses, and confirm maintenance targets.
3. **Releases, installation, and upgrades:** identify application versions, validate CI and packaged controller binaries, publish checksums and release notes, and prove configuration-preserving upgrades and documented rollback.

Release v0.1.0 only after the README acceptance checks pass: a first-time user reaches authenticated health information within ten minutes, two nodes sharing a role are both targeted, failure behavior is trustworthy, and packaged installation / upgrade / rollback have been exercised.

Discovery moves immediately after the MVP, ahead of advanced fleet features. Manual addition must work first; discovery returns candidates and reuses the same onboarding flow. An open SSH port does not identify a Raspberry Pi. SSH agent / encrypted-key support and custom ports follow in that milestone.

Monitoring, advanced diagnostics, inventory, services, rolling maintenance, scripting enhancements, package managers, self-updating, and optional Kubernetes / Docker integrations remain later work. Do not expand v0.1.0 to include them without an explicit scope change.

Roadmap entries describe planned behavior until implemented. In particular, the current config path rules and SSH trust behavior above describe the existing code; the MVP checklist defines the required changes. When implementing a milestone that changes CLI behavior, update the bundled user-facing skill and its version in the same change. A roadmap-only edit does not change the CLI or skill version.

---

## Go Concepts by Milestone

| Milestone | Concepts Introduced |
|---|---|
| Existing foundation | Structs, goroutines, channels, YAML parsing, Cobra commands, SSH sessions, metric parsing |
| MVP 1 — First Pi onboarding | Configuration validation, filesystem paths, safe writes, nested commands, actionable errors |
| MVP 2 — Core reliability | Slice filtering, table-driven tests, SSH host-key callbacks, deadlines, cancellation, resource cleanup, result aggregation |
| MVP 3 — Releases and upgrades | Build-time version injection, platform targets, CI automation, compatibility and migration design |
| Next release — Discovery and authentication | CIDR iteration, bounded workers, candidate identification, SSH agents and encrypted keys |
| Later — Scripting and fleet usability | JSON output, shared orchestration, timing and concurrency controls |
| Later — Health and inventory | Bitmasks, sysfs parsing, timers, normalized records, drift comparisons |
| Later — Services and maintenance | Nested commands, bounded logs, sequential orchestration, readiness polling |
