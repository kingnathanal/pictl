package ssh

import (
	"fmt"

	"github.com/kingnathanal/pictl/internal/config"
)

type ExecResult struct {
	Node     config.Node
	Stdout   string
	Stderr   string
	Err      error
}

func ExecAll(nodes []config.Node, keyPath string, command string) []ExecResult {
	results := make([]ExecResult, len(nodes))
	for i, node := range nodes {
		results[i] = execNode(node, keyPath, command)
	}
	return results
}

func execNode(node config.Node, keyPath string, command string) ExecResult {
	client, err := NewClient(node.User, node.IP, keyPath)
	if err != nil {
		return ExecResult{Node: node, Err: err}
	}
	out, err := RunCommand(client, command)
	if err != nil {
		return ExecResult{Node: node, Err: err}
	}
	return ExecResult{Node: node, Stdout: out}
}

func PrintExecResults(results []ExecResult) {
	for _, res := range results {
		fmt.Printf("\n%-5s %-15s %-2s %-s\n", "Node:", res.Node.Name, "IP:", res.Node.IP)
		fmt.Println("------------------------------------------------------")
		if res.Err != nil {
			fmt.Printf("Error: %v\n", res.Err)
			if res.Stderr != "" {
				fmt.Printf("Stderr: %s\n", res.Stderr)
			}
		} else {
			fmt.Printf("%s\n", res.Stdout)
		}
	}
}