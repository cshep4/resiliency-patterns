package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/cshep4/resiliency-patterns/high-availability/leader-election/internal/leaderelection"
)

func main() {
	nodeID := "node-" + uuid.New().String()

	log.Printf("Starting leader election demo for node: %s", nodeID)
	log.Printf("💡 Tip: Run multiple instances to see leader election in action")

	elector := leaderelection.NewLeaderElector(nodeID)

	// Block until we acquire leadership
	elector.AcquireLease()

	// Now we are the leader, start heartbeat loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("✅ [%s] Now running as LEADER. Press Ctrl+C to stop", nodeID)

	for {
		select {
		case <-sigChan:
			log.Printf("🛑 [%s] Received shutdown signal, stopping...", nodeID)
			log.Printf("👋 [%s] Shutdown complete", nodeID)
			return
		case <-ticker.C:
			log.Printf("👑 [%s] Status: LEADER - Heartbeat at %s", 
				nodeID, time.Now().Format("15:04:05"))
		}
	}
}
