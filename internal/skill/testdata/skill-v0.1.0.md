---
name: pictl
description: Use the pictl CLI to inspect, check, update, and run commands across a fleet of Raspberry Pis over SSH. Use when a user asks to operate or troubleshoot a Raspberry Pi fleet managed by pictl.
metadata:
  version: "0.1.0"
---

# pictl

Use `pictl` to manage Raspberry Pi OS or Debian nodes defined in `cluster.yaml`.

## Safety

- Start with read-only commands such as `pictl ping`, `pictl info`, or `pictl check`.
- Use `--node <name>` or `--role <role>` to limit scope before acting on the whole fleet.
- Explain the intended effect and ask for confirmation before updates or arbitrary commands that modify remote systems.
- Never display or copy SSH private keys, passwords, or other secrets from configuration.
- Do not assume K3s or another workload is installed unless the user confirms it.

## Configuration

By default, `pictl` reads `cluster.yaml` from the current directory. Use `--config <path>` or the `PICTL_CONFIG` environment variable to select another configuration file.

Node definitions include a name, IP address, role, and SSH user. The role is an optional grouping label and is not limited to Kubernetes roles.

## Commands

### Connectivity

```bash
pictl ping
pictl ping --node pi03-lan
pictl ping --role worker
```

### Fleet information

```bash
pictl info
pictl info --node pi03-lan
pictl info --role worker
```

### Health checks

`check` reports disk usage, CPU temperature, throttling, failed systemd units, and reboot-required state.

```bash
pictl check
pictl check --node pi03-lan
pictl check --role worker
```

### Package updates

`update` runs APT updates on selected nodes. Treat it as a disruptive command.

```bash
pictl update --node pi03-lan
pictl update --role worker
```

Prefer one node first, verify its health, and then continue with a larger target.

### Remote command execution

`exec` runs an arbitrary shell command on selected nodes. Prefer read-only diagnostic commands unless the user explicitly requests a change.

```bash
pictl exec --node pi03-lan "uptime"
pictl exec --role worker "uname -a"
```

Quote the remote command so the local shell passes it as one argument.

## Operating workflow

1. Confirm the configuration path and intended node or role.
2. Run `pictl ping` for the target.
3. Run `pictl check` or `pictl info` to establish current health.
4. Perform the requested operation using the narrowest target.
5. Run the relevant health command again and report per-node failures clearly.
