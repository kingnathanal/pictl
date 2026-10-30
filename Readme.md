# pictl

A lightweight CLI tool written in Go for managing a fleet of Raspberry Pis over SSH. Built as a hands-on Go learning project: real tool, real devices, real operational workflows.

---

## Features

- **Ping sweep** — check SSH reachability across all nodes concurrently
- **Node stats** — pull CPU usage, CPU temp, memory, disk, hostname, and OS info from every Pi in one shot
- **Apt update** — run `apt update && apt upgrade` across the fleet
- **AI skill installer** — teach supported AI coding agents how to operate `pictl` safely

---

## Project Structure

```
pictl/
├── cmd/
│   ├── root.go          # Cobra root command
│   ├── ping.go          # pictl ping
│   ├── info.go          # pictl info
│   ├── update.go        # pictl update
│   ├── exec.go          # pictl exec
│   ├── check.go         # pictl check
│   └── skill.go         # pictl skill install
├── internal/
│   ├── config/
│   │   └── config.go    # cluster.yaml loader
│   ├── ssh/
│   │   ├── client.go    # SSH connection + command runner
│   │   ├── ping.go      # TCP reachability check
│   │   ├── stats.go     # Node metrics collection
│   │   └── update.go    # Apt update logic
│   └── skill/
│       ├── skill.go     # Embedded skill installer
│       └── templates/pictl/SKILL.md
├── cluster.yaml         # Fleet node definitions
├── go.mod
├── go.sum
└── main.go
```

---

## Prerequisites

- Go 1.21+
- SSH key-based auth configured on each Pi (`ssh-copy-id pi@<ip>`)
- Raspberry Pi OS / Debian on each node
- SSH enabled on each Pi (`sudo systemctl enable ssh --now`)

---

## Installation

```bash
# Clone the repo
git clone https://github.com/yourusername/pictl.git
cd pictl

# Install dependencies
go mod tidy

# Build the binary
go build -o pictl .

# Optional: move to PATH
mv pictl /usr/local/bin/pictl
```

---

## Configuration

Create a `cluster.yaml` in the same directory you run `pictl` from:


| Field | Description |
|---|---|
| `ssh_key_path` | Path to your SSH private key. Supports `~` expansion. |
| `name` | Display name for the node |
| `ip` | Node IP address |
| `role` | Optional grouping label such as `worker`, `kiosk`, `camera`, `sensor`, or `control-plane` |
| `user` | SSH user on the Pi (typically `pi`) |

---

## Usage

### `pictl ping`
Check SSH reachability on port 22 for all nodes concurrently.

```bash
./pictl ping
```

```
NODE            IP               STATUS     LATENCY
------------------------------------------------------
pi-cp-01        192.168.1.101    ✅ UP      12ms
pi-cp-02        192.168.1.102    ✅ UP      14ms
pi-worker-01    192.168.1.104    ❌ DOWN    —
```

---

### `pictl info`
SSH into each node and collect CPU usage, CPU temperature, memory, disk, hostname, and OS. All nodes are queried concurrently — typically completes in ~2 seconds.

```bash
./pictl info
```

```
Collecting node stats (this takes ~2s for CPU measurement)...

NODE            IP               HOSTNAME     OS                          CPU%    CPU TEMP    MEM USED/TOTAL      DISK
-----------------------------------------------------------------------------------------------------------------------
pi-cp-01        192.168.1.101    pi-cp-01     Debian GNU/Linux 13         0.2%    57.3°C      670MB / 8062MB      8.3G/117G (8%)
pi-cp-02        192.168.1.102    pi-cp-02     Debian GNU/Linux 13         0.0%    56.2°C      669MB / 8062MB      8.3G/117G (8%)
pi-worker-01    192.168.1.104    pi-worker-01 Debian GNU/Linux 13         0.2%    55.1°C      660MB / 8062MB      8.3G/117G (8%)
```

> CPU usage is measured using two `/proc/stat` snapshots 1 second apart within a single SSH session, then calculating the delta.

---

### `pictl update`
Run `apt update && apt upgrade -y` on each node sequentially to avoid dpkg lock conflicts.

```bash
./pictl update
./pictl update --node pi03-lan
./pictl update --role worker
```

---

### `pictl exec`
Run an arbitrary shell command across selected nodes.

```bash
./pictl exec "uname -a"
./pictl exec --node pi03-lan "hostname"
./pictl exec --role worker "uptime"
```

```
Node: pi03-lan        IP: 192.168.1.134
--------------------------------------------------
pi03-lan
```

---

### `pictl check`
Run health checks on each node and report pass/warn/fail results. Checks: disk space, CPU temperature, throttling, failed systemd units, and reboot-required state.

```bash
./pictl check
./pictl check --node pi03-lan
./pictl check --role worker
```

```
Node: pi00-wifi (IP: 192.168.1.126)
--------------------------------------------------
✅ Failed Units: No failed units
✅ Disk Space: 10% used
✅ Temperature: 33.60°C
✅ Throttling: No throttling detected
✅ Reboot Required: No reboot required
```

---

### Targeting flags
All commands support `--node` and `--role` flags to target a subset of the fleet.

```bash
./pictl ping --node pi03-lan
./pictl info --role worker
./pictl check --node pi05-lan
```

---

### `pictl skill install`
Install the bundled `pictl` AI skill into the current project's supported skill directories:

- `.agents/skills/pictl/SKILL.md`
- `.claude/skills/pictl/SKILL.md`
- `.codex/skills/pictl/SKILL.md`

Each installation includes a `.pictl-skill.json` manifest containing the installed version and SHA-256 checksum.

```bash
./pictl skill install
./pictl skill status
./pictl skill install --force
```

Running the command again leaves matching files unchanged. Older managed skills are upgraded automatically when their checksum confirms they have not been edited. Modified, unmanaged, inconsistent, or newer skills are protected unless `--force` is provided.

Dangling `SKILL.md` or manifest symlinks are reported as incomplete and require `--force` to replace. Forced replacement replaces the link itself without writing to its missing destination.

If a filesystem error interrupts installation, completed targets are listed before the error. Other targets may be missing or partially written. Run `pictl skill status` to inspect all installations before retrying.

---

## Roadmap

The v0.1.0 MVP promise is: **Connect your existing Raspberry Pis and see their health from one terminal.**

The release targets users with network access to SSH-enabled Raspberry Pi OS / Debian devices and an existing SSH key. OS imaging, enabling SSH, and provisioning keys remain prerequisites documented in the quickstart. The CLI runs on the controller; users do not need to install it on every Pi.

This is the canonical product roadmap. Checked items below describe existing capabilities, not a claim that the public MVP is release-ready. Unchecked commands and behaviors are planned and must not be treated as available today. All three MVP milestones and the release acceptance checks must be complete before v0.1.0.

### Existing foundation

- [x] TCP reachability checks with `pictl ping`
- [x] Node metrics with `pictl info`
- [x] Basic health checks for disk space, temperature, throttling, failed systemd units, and reboot-required state
- [x] Remote command execution with `pictl exec`
- [x] Sequential apt maintenance with `pictl update`
- [x] Configuration overrides through `--config` and `PICTL_CONFIG`
- [x] `--node` targeting and an initial `--role` implementation; role completeness remains an MVP blocker below
- [x] Bundled AI skill installation, status reporting, and manifest-backed upgrade protection

### MVP 1 — First Pi onboarding

Goal: reach a successful authenticated health check without hand-writing YAML.

- [ ] Add `pictl init` to create a valid empty configuration without overwriting existing files
- [ ] Establish and document a predictable user configuration location outside the application checkout; preserve explicit `--config` / `PICTL_CONFIG` overrides and document compatibility with existing `cluster.yaml` files
- [ ] Keep node definitions, SSH settings, and trusted host keys separate from the installed binary so upgrades preserve them
- [ ] Add `pictl node add` with name, address, SSH user, key configuration, and optional role
- [ ] Validate required fields, malformed configuration, and duplicate node names / addresses; reject invalid edits without damaging existing configuration
- [ ] Add `pictl node list`, `pictl node update`, and `pictl node remove`, including changing an address after DHCP changes
- [ ] Save configuration safely and require deliberate confirmation for destructive changes
- [ ] Provide an explicit authenticated connection check during onboarding; distinguish network failure, unreadable / unsupported key, authentication failure, and host-key mismatch, with actionable next steps
- [ ] Explain that `ping` only checks TCP reachability and does not verify authentication or Raspberry Pi identity
- [ ] Guide users from adding a node to `pictl info --node <name>` and `pictl check --node <name>`
- [ ] Document the supported key format and port-22 limitation; keep password auth, encrypted-key handling, and SSH agent integration out of the initial release
- [ ] Write a complete quickstart covering prerequisites, installation, configuration, adding the first Pi, checking access, and troubleshooting; correct the placeholder clone URL and Go version requirement

### MVP 2 — Core reliability and safe operation

Goal: users can trust which devices are selected and whether an operation succeeded.

- [ ] Fix `--role` selection to include every matching node
- [ ] Define and test combined `--node` / `--role` behavior; reject unknown targets and empty operational selections with useful errors
- [ ] Validate configuration before connecting or running remote commands
- [ ] Verify SSH host keys by default, support deliberate first-use trust, and reject unexpected key changes
- [ ] Bound connection, authentication, and remote-command waits; allow users to cancel operations cleanly
- [ ] Close SSH clients on every execution path and retain diagnostic command output when remote commands fail
- [ ] Define and implement meaningful exit statuses for invalid input, connection / command failures, partial fleet failures, and failed health checks; distinguish warnings from failures
- [ ] Print clear per-node errors and a fleet summary; avoid reporting failed or unavailable measurements as successful zero values
- [ ] Show the selected nodes before `update`, require confirmation, and provide an explicit noninteractive confirmation option; handle sudo failures without hanging
- [ ] Add focused tests for configuration, all matching role members, unknown / combined targets, SSH trust and authentication failures, timeouts, command output, and partial failures
- [ ] Validate `ping`, `info`, `check`, and a harmless `exec` against representative supported devices; validate maintenance behavior in a disposable test environment

### MVP 3 — Releases, installation, and upgrades

Goal: a user can install a named release, upgrade it, and recover without losing fleet configuration.

- [ ] Add `pictl --version` with the application release version and clearly labeled development builds
- [ ] Adopt version tags starting at `v0.1.0`; document patch / minor conventions and compatibility expectations before v1.0
- [ ] Add pull-request CI for build, tests, and vet; require passing checks on the release commit
- [ ] Select and validate controller platforms, initially targeting Linux amd64 / arm64 and macOS arm64; publish support only for validated targets
- [ ] Automate tagged builds and packaging with GitHub Actions and a release tool such as GoReleaser
- [ ] Produce a draft GitHub Release containing versioned binaries, checksums, release notes, and installation instructions
- [ ] Smoke-test packaged artifacts on supported targets, including reported version and first-Pi workflow, before publishing the release
- [ ] Document checksum verification, installing to PATH, and the source-build alternative using the Go version required by `go.mod`
- [ ] Document upgrading by replacing the binary, checking its version, listing saved nodes, and making a read-only connection check
- [ ] Preserve existing configuration across upgrades; document compatibility and any explicit migration steps, with backups before incompatible changes
- [ ] Document reinstalling the previous release and the configuration compatibility needed for rollback
- [ ] Keep application and bundled AI skill versions distinct; explain refreshing installed skills with `pictl skill install` after upgrading the application

### v0.1.0 release acceptance

- [ ] A new user with an SSH-enabled Pi and supported key can install a packaged release, initialize configuration, add the Pi, authenticate, and retrieve its health within ten minutes using only the quickstart
- [ ] Adding a second Pi with the same role causes role-targeted commands to select both devices
- [ ] Missing configuration, invalid targets, unreachable devices, authentication failures, and changed host keys produce actionable errors and documented exit statuses
- [ ] Partial fleet failures remain visible in both output and exit status; maintenance requires deliberate confirmation
- [ ] Upgrading from a release candidate preserves fleet configuration and trusted host keys; the documented rollback works with the tested configuration format
- [ ] Release artifacts pass the published-platform smoke checks, and checksums, release notes, and installation / upgrade instructions are available

### Next release — Discovery and easier authentication

Discovery follows manual onboarding and reuses its validation and addition flow. It is not required for v0.1.0, and the initial release must not promise automatic Pi discovery.

- [ ] Add `pictl discover --subnet <CIDR>` with bounded concurrency and timeouts for an explicitly selected network
- [ ] Detect SSH candidates and show results before adding anything to configuration
- [ ] Identify Raspberry Pis through authenticated SSH where possible; label unverified candidates as unknown rather than equating an open port with a Pi
- [ ] Let users select candidates and add them through the existing onboarding flow with duplicate checks
- [ ] Explain discovery limitations: SSH-disabled devices, inaccessible networks, and devices that cannot be authenticated
- [ ] Add custom SSH ports consistently across configuration, discovery, ping, and SSH operations
- [ ] Add SSH agent and passphrase-aware key support; consider optional password prompts without storing plaintext passwords
- [ ] Consider `PICTL_SSH_KEY` as a documented key-path override

### Later — Fleet usability and scripting

- [ ] Add machine-readable `--json`, `--quiet` / `--no-header`, and configurable result sorting
- [ ] Add total and per-node timing for long-running commands
- [ ] Consider a shared fleet runner and consistent concurrency patterns when needed by new features
- [ ] Add selectable AI skill installation targets
- [ ] Add package-manager distribution and update notifications after the binary upgrade path is proven
- [ ] Consider a built-in `pictl upgrade` only with verified downloads, permission handling, interrupted-update recovery, and package-manager ownership rules

### Later — Health, monitoring, and inventory

- [ ] Decode current and historical undervoltage / throttling bits separately and report Pi 5 PMIC rails where available
- [ ] Add `pictl cooling status` for fan RPM, cooling level, and thermal trip points
- [ ] Add `pictl clock status` for RTC presence, NTP synchronization, and controller / node drift
- [ ] Add `pictl top` / `pictl watch` with configurable refresh intervals and high-CPU alerts
- [ ] Add `pictl inventory` for model, serial number, RAM, storage, OS, kernel, architecture, and interfaces
- [ ] Decode board revision identifiers and audit bootloader EEPROM versions / boot order
- [ ] Report read-only roots, storage corruption signals, and SMART health where supported
- [ ] Track configuration drift and expand systemd health diagnostics

### Later — Services and maintenance

- [ ] Add `pictl service status`, `pictl service restart`, and bounded `pictl logs` with targeting, time, and line filters
- [ ] Add reboot and rolling reboot / update workflows with readiness checks and failure handling
- [ ] Consider explicitly requested parallel updates with controlled concurrency
- [ ] Add `pictl overlay status|enable|disable` with sequential execution, confirmation, and reboot handling

### Optional modules

- [ ] K3s installation, status, and kubeconfig workflows
- [ ] Docker container inspection
- [ ] Monitoring installation
- [ ] Kiosk status
- [ ] Hardware-interface configuration audits for I2C, SPI, 1-Wire, and serial consoles

These modules do not block the core fleet-management MVP.
