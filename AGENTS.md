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
│   └── update.go           # pictl update
├── internal/               # All business logic — not importable outside module
│   ├── config/
│   │   └── config.go       # cluster.yaml loader, Config + Node structs
│   └── ssh/
│       ├── common.go       # NewClient(), RunCommand() — shared SSH primitives
│       ├── ping.go         # TCP reachability check, PingAll(), printTable()
│       ├── stats.go        # Node metrics collection, CollectAll(), PrintStatsTable()
│       └── update.go       # Apt update logic, UpdateAll(), PrintUpdateResults()
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

### M3 — Fleet targeting and command execution

- Add `--node` targeting for a single Pi
- Add `--role` targeting for groups of Pis
- Add `pictl exec` to run a command across selected nodes
- Add safer command output formatting for multi-node results

New Go concepts introduced: shared command filters, CLI flags, command argument handling, concurrent remote execution with structured results.

### M4 — Fleet health checks

- Add `pictl check` for pass/warn/fail health checks
- Detect undervoltage and throttling with `vcgencmd get_throttled`
- Detect low disk space, high temperature, failed systemd units, and reboot-required state
- Add summary output that makes unhealthy nodes obvious

New Go concepts introduced: check structs, severity levels, composable health checks, parsing command output into typed results.

### M5 — Inventory and audit

- Add `pictl inventory`
- Report Pi model, serial number, RAM, storage, OS, kernel, architecture, and network interfaces
- Add machine-readable output with `--json`
- Track configuration drift across the fleet

New Go concepts introduced: JSON output, richer data models, normalized inventory records.

### M6 — Services and logs

- Add `pictl service status <unit>`
- Add `pictl service restart <unit>`
- Add `pictl logs <unit>` using `journalctl`
- Support `--since`, `--lines`, `--node`, and `--role`

New Go concepts introduced: subcommands with arguments, systemd-oriented workflows, streaming or bounded log output.

### M7 — Rolling maintenance

- Add `pictl reboot`
- Add `pictl reboot --rolling`
- Add `pictl update --rolling`
- Add confirmation prompts for disruptive operations

New Go concepts introduced: sequential orchestration, readiness polling, user confirmation, failure handling for disruptive operations.

### Optional modules

- [ ] `pictl k3s install`
- [ ] `pictl k3s status`
- [ ] `pictl k3s kubeconfig`
- [ ] `pictl docker ps`
- [ ] `pictl monitoring install`
- [ ] `pictl kiosk status`

---

## Go Concepts by Milestone

| Milestone | Concepts Introduced |
|---|---|
| M1 — Ping | Structs, goroutines, `sync.WaitGroup`, YAML parsing, Cobra CLI |
| M2 — Node stats | Channels, SSH client, string/strconv parsing, import aliasing, error wrapping, two-snapshot CPU delta |
| M2 — Update | Sequential vs concurrent tradeoffs, `DEBIAN_FRONTEND`, exit code handling |
| M3 — Targeting + exec | CLI flags, filtering slices by field, concurrent remote command execution |
| M3.1 — Orchestration | Unified workers, environment binding, consistent concurrency |
| M3.2 — Scripting | JSON marshalling, global flag management, exit code standards |
| M4 — Health checks | Typed checks, severity levels, parsing system command output |
| M4.1 — Advanced Health | Multi-stage command execution, existence checks |
| M5 — Inventory | JSON output, normalized data models, audit-friendly records |
| M6 — Services + logs | Nested commands, systemd workflows, bounded log retrieval |
| M7 — Rolling maintenance | Sequential orchestration, readiness checks, confirmation prompts |
tructs, goroutines, `sync.WaitGroup`, YAML parsing, Cobra CLI |
| M2 — Node stats | Channels, SSH client, string/strconv parsing, import aliasing, error wrapping, two-snapshot CPU delta |
| M2 — Update | Sequential vs concurrent tradeoffs, `DEBIAN_FRONTEND`, exit code handling |
| M3 — Targeting + exec | CLI flags, filtering slices by field, concurrent remote command execution |
| M4 — Health checks | Typed checks, severity levels, parsing system command output |
| M5 — Inventory | JSON output, normalized data models, audit-friendly records |
| M6 — Services + logs | Nested commands, systemd workflows, bounded log retrieval |
| M7 — Rolling maintenance | Sequential orchestration, readiness checks, confirmation prompts |
