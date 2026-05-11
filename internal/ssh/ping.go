package ssh

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kingnathanal/pictl/internal/config"
)

type PingResult struct {
	Node    config.Node
	Up      bool
	Latency time.Duration
}

func PingAll(nodes []config.Node) {
	results := make([]PingResult, len(nodes))
	var wg sync.WaitGroup

	for i, node := range nodes {
		wg.Add(1)
		go func(i int, node config.Node) {
			defer wg.Done()
			results[i] = pingNode(node)
		}(i, node)
	}

	wg.Wait()
	printResults(results)
}

func pingNode(node config.Node) PingResult {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", node.IP+":22", 3*time.Second)
	latency := time.Since(start)

	if err != nil {
		return PingResult{Node: node, Up: false}
	}

	conn.Close()
	return PingResult{Node: node, Up: true, Latency: latency}
}

func printResults(results []PingResult) {
	fmt.Printf("\n%-15s %-16s %-10s %s\n", "Node", "IP", "Status", "Latency")
	fmt.Println("-------------------------------------------------------------")
	for _, r := range results {
		status := "✅ UP"
		latency := fmt.Sprintf("%dms", r.Latency.Milliseconds())
		if !r.Up {
			status = "❌ DOWN"
			latency = "-"
		}
		fmt.Printf("%-15s %-16s %-10s %s\n", r.Node.Name, r.Node.IP, status, latency)
	}
	fmt.Println()
}
