package ssh

import (
	"fmt"
	"strconv"
	"strings"
	"sort"

	"github.com/kingnathanal/pictl/internal/config"
	gossh "golang.org/x/crypto/ssh"
)

type CheckResult struct {
	Name string
	Status string 
	Message string
}

type NodeReport struct {
	Node config.Node
	Checks []CheckResult
}

func CheckAll(nodes []config.Node, keyPath string) []NodeReport {
	results := make([]NodeReport, len(nodes))
	for i, node := range nodes {
		results[i] = checkNode(node, keyPath)
	}
	return results
}

func checkNode(node config.Node, keyPath string) NodeReport { 
	client, err := NewClient(node.User, node.IP, keyPath)
	if err != nil {
		return NodeReport{Node: node, Checks: []CheckResult{
			{Name: "SSH Connection", Status: "FAIL", Message: err.Error()},
		}}
	}
	checks := []CheckResult{
		checkFailedUnits(client),
		checkDiskSpace(client),
		checkTemperature(client),
		checkThrottling(client),
		checkRebootRequired(client),
	}
	client.Close()
	return NodeReport{Node: node, Checks: checks}
}

func checkDiskSpace(client *gossh.Client) CheckResult { 
	out, err := RunCommand(client, "df / --output=pcent | tail -1")
	if err != nil {
		return CheckResult{Name: "Disk Space", Status: "FAIL", Message: err.Error()}
	}

	pctStr := strings.TrimSpace(strings.Trim(out, "%\n"))
	pct, err := strconv.Atoi(pctStr)
	if err != nil {
		return CheckResult{Name: "Disk Space", Status: "FAIL", Message: "could not parse disk usage"}
	}

	switch { 
	case pct >= 90:
		return CheckResult{Name: "Disk Space", Status: "FAIL", Message: fmt.Sprintf("%d%% used", pct)}
	case pct >= 75:
		return CheckResult{Name: "Disk Space", Status: "WARN", Message: fmt.Sprintf("%d%% used", pct)}
	default:
		return CheckResult{Name: "Disk Space", Status: "PASS", Message: fmt.Sprintf("%d%% used", pct)}
	}
}

func checkTemperature(client *gossh.Client) CheckResult { 
	out, err := RunCommand(client, "vcgencmd measure_temp")
	if err != nil {
		return CheckResult{Name: "Temperature", Status: "FAIL", Message: err.Error()}
	}

	tempStr := strings.TrimPrefix(out, "temp=")
	tempStr = strings.TrimSuffix(tempStr, "'C\n")
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return CheckResult{Name: "Temperature", Status: "FAIL", Message: "could not parse temperature"}
	}

	switch {
	case temp >= 80:
		return CheckResult{Name: "Temperature", Status: "FAIL", Message: fmt.Sprintf("%.2f°C", temp)}
	case temp >= 70:
		return CheckResult{Name: "Temperature", Status: "WARN", Message: fmt.Sprintf("%.2f°C", temp)}
	default:
		return CheckResult{Name: "Temperature", Status: "PASS", Message: fmt.Sprintf("%.2f°C", temp)}
	}
}

func checkThrottling(client *gossh.Client) CheckResult { 
	out, err := RunCommand(client, "vcgencmd get_throttled")
	if err != nil {
		return CheckResult{Name: "Throttling", Status: "FAIL", Message: err.Error()}
	}

	throttleStr := strings.TrimPrefix(out, "throttled=")
	throttleStr = strings.TrimSuffix(throttleStr, "\n")
	throttleVal, err := strconv.ParseInt(throttleStr, 0, 64)
	if err != nil {
		return CheckResult{Name: "Throttling", Status: "FAIL", Message: "could not parse throttling status"}
	}

	if throttleVal == 0 {
		return CheckResult{Name: "Throttling", Status: "PASS", Message: "No throttling detected"}
	}

	return CheckResult{Name: "Throttling", Status: "FAIL", Message: fmt.Sprintf("Throttling detected (0x%X)", throttleVal)}
}

func checkFailedUnits(client *gossh.Client) CheckResult { 
	out, err := RunCommand(client, "systemctl --failed --no-legend")
	if err != nil {
		return CheckResult{Name: "Failed Units", Status: "FAIL", Message: err.Error()}
	}

	if strings.TrimSpace(out) == "" { 
		return CheckResult{Name: "Failed Units", Status: "PASS", Message: "No failed units"}
	}

	return CheckResult{Name: "Failed Units", Status: "FAIL", Message: strings.TrimSpace(out)}
}

func checkRebootRequired(client *gossh.Client) CheckResult { 
	out, err := RunCommand(client, "test -f /var/run/reboot-required && echo 'yes' || echo 'no'")
	if err != nil {
		return CheckResult{Name: "Reboot Required", Status: "FAIL", Message: err.Error()}
	}

	if strings.TrimSpace(out) == "yes" { 
		return CheckResult{Name: "Reboot Required", Status: "WARN", Message: "Reboot is required"}
	}

	return CheckResult{Name: "Reboot Required", Status: "PASS", Message: "No reboot required"}
}

func PrintCheckResults(reports []NodeReport) {

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Node.Name < reports[j].Node.Name
	})

	for _, report := range reports {
		fmt.Printf("\nNode: %s (IP: %s)\n", report.Node.Name, report.Node.IP)
		fmt.Println("--------------------------------------------------")
		for _, check := range report.Checks {
			statusIcon := "✅"
			if check.Status == "FAIL" {
				statusIcon = "❌"
			} else if check.Status == "WARN" {
				statusIcon = "⚠️"
			}
			fmt.Printf("%s %s: %s\n", statusIcon, check.Name, check.Message)
		}
	}
}