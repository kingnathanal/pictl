# 🥧 pictl

A lightweight CLI tool written in Go for managing a Raspberry Pi K3s home lab cluster. Built as a hands-on Go learning project — real tool, real cluster, real code.

---

## Features

- **Ping sweep** — check SSH reachability across all nodes concurrently
- **Node stats** — pull CPU usage, CPU temp, memory, disk, hostname, and OS info from every Pi in one shot
- **Apt update** — run `apt update && apt upgrade` across the entire cluster in parallel

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
├── cluster.yaml         # Node definitions
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
| `role` | `control-plane` or `worker` (informational for now) |
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

More on the way