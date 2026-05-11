package ssh

import (
	"fmt"
	"strconv"
	"strings"
	//"time"

	"github.com/kingnathanal/pictl/internal/config"
	gossh "golang.org/x/crypto/ssh"
)

type NodeStats struct {
	Node        config.Node
	Hostname    string
	OS          string
	CPUUsage    float64
	CPUTemp     float64
	MemUsedMB   int
	MemTotalMB  int
	DiskUsed    string
	DiskTotal   string
	DiskPercent string
	Error       error
}

func CollectAll(nodes []config.Node, keyPath string) []NodeStats {

	resultCh := make(chan NodeStats, len(nodes))

	for _, node := range nodes {
		go func(n config.Node) {
			resultCh <- collectNode(n, keyPath)
		}(node)
	}

	results := make([]NodeStats, 0, len(nodes))
	for range nodes {
		results = append(results, <-resultCh)
	}

	return results
}

func collectNode(node config.Node, keyPath string) NodeStats {

	stats := NodeStats{Node: node}

	client, err := NewClient(node.User, node.IP, keyPath)
	if err != nil {
		stats.Error = err
		return stats
	}
	defer client.Close()

	stats.Hostname = runOrDefault(client, "hostname", "unknown")
	stats.OS = parseOS(runOrDefault(client, `grep PRETTY /etc/os-release`, "unknown"))
	stats.CPUTemp = parseCPUTemp(runOrDefault(client, "cat /sys/class/thermal/thermal_zone0/temp", "0"))
	stats.CPUUsage = measureCPUUsage(client)
	parseMemory(&stats, runOrDefault(client, `free -m | awk 'NR==2{print $2, $3, $4}'`, "0 0 0"))
	parseDisk(&stats, runOrDefault(client, `df -h / | awk 'NR==2{print $2, $3, $5}'`, "0 0 0%"))

	return stats
}

func measureCPUUsage(client *gossh.Client) float64 {
	cmd := "cat /proc/stat | head -1; sleep 1; cat /proc/stat | head -1"
	out, err := RunCommand(client, cmd)
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return 0
	}

	snap1 := lines[0]
	snap2 := lines[1]

	return calculateCPUPercent(snap1, snap2)
}

func calculateCPUPercent(snap1, snap2 string) float64 {
	parse := func(line string) (idle, total int64) {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return 0, 0
		}
		var vals [10]int64
		for i := 1; i < len(fields) && i <= 10; i++ {
			vals[i-1], _ = strconv.ParseInt(fields[i], 10, 64)
		}
		idle = vals[3]
		for _, v := range vals {
			total += v
		}
		return idle, total
	}

	idle1, total1 := parse(snap1)
	idle2, total2 := parse(snap2)

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)

	if totalDelta == 0 {
		return 0
	}

	return (1.0 - idleDelta/totalDelta) * 100.0
}

func parseOS(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "PRETTY_NAME=")
	return strings.Trim(raw, `"`)
}

func parseCPUTemp(raw string) float64 {
	raw = strings.TrimSpace(raw)
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return val / 1000.0
}

func parseMemory(stats *NodeStats, raw string) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) < 2 {
		return
	}
	stats.MemTotalMB, _ = strconv.Atoi(fields[0])
	stats.MemUsedMB, _ = strconv.Atoi(fields[1])
}

func parseDisk(stats *NodeStats, raw string) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) < 3 {
		return
	}
	stats.DiskTotal = fields[0]
	stats.DiskUsed = fields[1]
	stats.DiskPercent = fields[2]
}

func runOrDefault(client *gossh.Client, cmd, def string) string {
	out, err := RunCommand(client, cmd)
	if err != nil {
		return def
	}
	return strings.TrimSpace(out)
}

func PrintStatsTable(results []NodeStats) {
	fmt.Printf("\n%-15s %-16s %-12s %-28s %-10s %-10s %-16s %-10s\n",
		"NODE", "IP", "HOSTNAME", "OS", "CPU%", "CPU TEMP", "MEM USED/TOTAL", "DISK")
	fmt.Println(strings.Repeat("-", 128))

	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("%-15s %-16s ❌ ERROR: %v\n", r.Node.Name, r.Node.IP, r.Error)
			continue
		}
		mem := fmt.Sprintf("%dMB / %dMB", r.MemUsedMB, r.MemTotalMB)
		disk := fmt.Sprintf("%s/%s (%s)", r.DiskUsed, r.DiskTotal, r.DiskPercent)

		fmt.Printf("%-15s %-16s %-12s %-22s %-10s %-10s %-16s %-10s\n",
			r.Node.Name,
			r.Node.IP,
			r.Hostname,
			r.OS,
			fmt.Sprintf("%.1f%%", r.CPUUsage),
			fmt.Sprintf("%.1f°C", r.CPUTemp),
			mem,
			disk,
		)
	}
	fmt.Println()
}
