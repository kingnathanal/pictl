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

If a filesystem error interrupts installation, completed targets are listed before the error. Other targets may be missing or partially written. Run `pictl skill status` to inspect all installations before retrying.

---

## Roadmap

`pictl` is intended to stay useful for any Raspberry Pi running Raspberry Pi OS or Debian with SSH enabled. K3s support can be added later as an optional module, but the core tool should remain focused on generic fleet operations.

### M3 — Fleet targeting and command execution

- [x] Add global config path handling with `--config` and `PICTL_CONFIG`
- [x] Use Cobra `RunE` so command errors return cleanly to the root command
- [x] Add `--node` targeting for a single Pi
- [x] Add `--role` targeting for groups of Pis
- [x] Add `pictl exec` to run a command across selected nodes
- [x] Add safer command output formatting for multi-node results
- [ ] Add total and per-node timing output for long-running commands like `update`
- [ ] Add `pictl update --parallel` for intentionally concurrent updates

### M3.1 — Orchestration & DX Improvements

- [ ] Support `PICTL_SSH_KEY` environment variable for global key path
- [ ] Support custom SSH ports in `cluster.yaml` (defaulting to 22)
- [ ] Implement a unified internal "Fleet Runner" to reduce boilerplate for concurrent operations
- [ ] Refactor `ping.go` to use the channel pattern for consistency with `stats.go`

### M3.2 — Output & Scripting

- [ ] Add machine-readable output with `--json` for all commands
- [ ] Add `--quiet` or `--no-header` flags for cleaner scripting integration
- [ ] Ensure `exec` and `update` return appropriate exit codes based on fleet success

### M3.5 — Fleet config management

- [ ] Add `pictl node list`
- [ ] Add `pictl node add`
- [ ] Add duplicate validation for node name and IP address
- [ ] Add `pictl node remove`
- [ ] Add `pictl node update` for changing role, user, or IP address
- [ ] Keep `cluster.yaml` as the source of truth while allowing commands to edit it
- [ ] Optionally validate SSH reachability when adding a node

### M4 — Fleet health checks

- [x] Add `pictl check` for pass/warn/fail health checks
- [x] Detect undervoltage and throttling with `vcgencmd get_throttled`
- [x] Detect low disk space and high temperature
- [x] Detect failed systemd units
- [x] Detect reboot-required state
- [ ] Add summary output that makes unhealthy nodes obvious
- [ ] Add `--sort` flag to control output order (alphabetical by default, options: name, ip, role, status)

### M4.1 — Advanced Pi Health

- [ ] Detect "Reboot Required" state via `/var/run/reboot-required`
- [ ] Report detailed throttling history (undervoltage vs overheating)
- [ ] Check for failed or degraded systemd units across the fleet

### M4.2 — Live fleet monitor

- [ ] Add `pictl top` — real-time fleet view using `top -bn1` over SSH, refreshed on an interval
- [ ] Add `pictl watch <command>` — re-run any pictl command on a set interval with terminal refresh
- [ ] Flag runaway processes consuming above a threshold (e.g. >80% CPU) per node
- [ ] Support `--interval` flag to control refresh rate (default: 5s)

New Go concepts introduced: tickers, terminal control sequences, looping concurrent SSH collection.

### M4.3 — Power and cooling diagnostics

- [ ] Add `pictl power status` — decode the persistent `vcgencmd get_throttled` bits (undervoltage occurred, throttling occurred), beyond the current-state bits `check` already reports
- [ ] Report Pi 5 PMIC rail voltage/current via `vcgencmd pmic_read_adc` where available
- [ ] Add `pictl cooling status` — fan RPM, active cooling level, and thermal trip points for PoE HAT / official fan nodes
- [ ] Add `pictl clock status` — RTC presence, NTP sync state, and clock drift between each node and the controller

New Go concepts introduced: bitmask decoding into typed flags, sysfs globbing across multiple small files, time drift calculations with `time.Duration`.

### M5 — Inventory and audit

- [ ] Add `pictl inventory`
- [ ] Report Pi model, serial number, RAM, storage, OS, kernel, architecture, and network interfaces
- [ ] Decode the `/proc/cpuinfo` revision code into exact model, RAM size, and board revision for mixed Pi 4/5 fleets
- [ ] Report bootloader EEPROM version and configured boot order (SD/USB/NVMe) via `rpi-eeprom-config` / `vcgencmd bootloader_version`, flagging drift across nodes
- [ ] Report storage health: read-only root detection, SD card corruption signals, and NVMe/SSD SMART status where available
- [ ] Add machine-readable output with `--json`
- [ ] Track configuration drift across the fleet

### M6 — Services and logs

- [ ] Add `pictl service status <unit>`
- [ ] Add `pictl service restart <unit>`
- [ ] Add `pictl logs <unit>` using `journalctl`
- [ ] Support `--since`, `--lines`, `--node`, and `--role`

### M7 — Rolling maintenance

- [ ] Add `pictl reboot`
- [ ] Add `pictl reboot --rolling`
- [ ] Add `pictl update --rolling`
- [ ] Add `pictl overlay status|enable|disable` — audit and toggle the Pi read-only overlay filesystem (changes held in RAM) for reliability-critical nodes; a write command requiring sequential execution, a confirmation prompt, and a reboot to apply
- [ ] Add confirmation prompts for disruptive operations

### M8 — Discovery and auth improvements

- [ ] Add `pictl discover --subnet 192.168.1.0/24`
- [ ] Detect SSH hosts on port 22
- [ ] Attempt Raspberry Pi identification through SSH when credentials are available
- [ ] Show discovered candidates before adding them to config
- [ ] Add optional password prompt auth without storing plaintext passwords in YAML
- [ ] Add SSH agent and passphrase-aware key support

### Optional modules

- [ ] `pictl k3s install`
- [ ] `pictl k3s status`
- [ ] `pictl k3s kubeconfig`
- [ ] `pictl docker ps`
- [ ] `pictl monitoring install`
- [ ] `pictl kiosk status`
- [ ] `pictl config audit` — audit I2C/SPI/1-Wire/serial-console enablement (`raspi-config nonint get_*`) across the fleet to catch interface drift before attaching sensors or reproducing a node setup

### AI integrations

- [x] Add `pictl skill install`
- [x] Install the bundled skill for Agents, Claude, and Codex project directories
- [x] Add manifest-backed automatic upgrades and modification protection
- [x] Add `pictl skill status`
- [ ] Add selectable installation targets
