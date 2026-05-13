# pictl

A lightweight CLI tool written in Go for managing a fleet of Raspberry Pis over SSH. Built as a hands-on Go learning project: real tool, real devices, real operational workflows.

---

## Features

- **Ping sweep** — check SSH reachability across all nodes concurrently
- **Node stats** — pull CPU usage, CPU temp, memory, disk, hostname, and OS info from every Pi in one shot
- **Apt update** — run `apt update && apt upgrade` across the fleet

---

## Project Structure

```
pictl/
├── cmd/
│   ├── root.go          # Cobra root command
│   ├── ping.go          # pictl ping
│   ├── info.go          # pictl info
│   └── update.go        # pictl update
├── internal/
│   ├── config/
│   │   └── config.go    # cluster.yaml loader
│   └── ssh/
│       ├── client.go    # SSH connection + command runner
│       ├── ping.go      # TCP reachability check
│       ├── stats.go     # Node metrics collection
│       └── update.go    # Apt update logic
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

## Roadmap

`pictl` is intended to stay useful for any Raspberry Pi running Raspberry Pi OS or Debian with SSH enabled. K3s support can be added later as an optional module, but the core tool should remain focused on generic fleet operations.

### M3 — Fleet targeting and command execution

- [ ] Add `--node` targeting for a single Pi
- [ ] Add `--role` targeting for groups of Pis
- [ ] Add `pictl exec` to run a command across selected nodes
- [ ] Add safer command output formatting for multi-node results

### M4 — Fleet health checks

- [ ] Add `pictl check` for pass/warn/fail health checks
- [ ] Detect undervoltage and throttling with `vcgencmd get_throttled`
- [ ] Detect low disk space, high temperature, failed systemd units, and reboot-required state
- [ ] Add summary output that makes unhealthy nodes obvious

### M5 — Inventory and audit

- [ ] Add `pictl inventory`
- [ ] Report Pi model, serial number, RAM, storage, OS, kernel, architecture, and network interfaces
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
- [ ] Add confirmation prompts for disruptive operations

### Optional modules

- [ ] `pictl k3s install`
- [ ] `pictl k3s status`
- [ ] `pictl k3s kubeconfig`
- [ ] `pictl docker ps`
- [ ] `pictl monitoring install`
- [ ] `pictl kiosk status`
