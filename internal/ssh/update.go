package ssh

import (
	"fmt"
	"strings"

	"github.com/kingnathanal/pictl/internal/config"
)

type UpdateResult struct {
	Node   config.Node
	Output string
	Error  error
}

func UpdateAll(nodes []config.Node, keyPath string) []UpdateResult {
	resultCh := make(chan UpdateResult, len(nodes))

	for _, node := range nodes {
		go func(n config.Node) {
			resultCh <- runUpdate(n, keyPath)
		}(node)
	}

	results := make([]UpdateResult, 0, len(nodes))
	for range nodes {
		results = append(results, <-resultCh)
	}
	
	return results
}

func runUpdate(node config.Node, keyPath string) UpdateResult {
	client, err := NewClient(node.User, node.IP, keyPath)
	if err != nil {
		return UpdateResult{Node: node, Error: err}
	}
	defer client.Close()

	out, err := RunCommand(client, "sudo apt update && sudo apt upgrade -y")
	return UpdateResult{Node: node, Output: out, Error: err}
}

func PrintUpdateResults(results []UpdateResult) {
	fmt.Printf("\n%-15s %-16s %-10s %s\n", "Node", "IP", "Status", "Details")
	fmt.Println(strings.Repeat("-", 70))

	for _, res := range results {
		if res.Error != nil {
			summary := parseSummary(res.Output)
			fmt.Printf("%-15s %-16s ❌ Failed		%s\n", res.Node.Name, res.Node.IP, summary)
			continue
		}

		summary := parseSummary(res.Output)
		fmt.Printf("%-15s %-16s ✅ OK		%s\n", res.Node.Name, res.Node.IP, summary)
	}
	fmt.Println()
}

func parseSummary(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "upgraded,") {
			return strings.TrimSpace(line)
		}
	}
	return "completed"
}