package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cshep4/resiliency-patterns/examples/high-availability/leader-election/internal/leaderelection"
)

func main() {
	var (
		nodeID      = flag.String("node-id", "", "Unique identifier for this node (required)")
		lockName    = flag.String("lock-name", "leader-election-demo", "Name of the leadership lock")
		lockDir     = flag.String("lock-dir", "/tmp", "Directory to store lock files")
		leaseDuration = flag.Duration("lease-duration", 10*time.Second, "Duration of leadership lease")
		retryPeriod = flag.Duration("retry-period", 2*time.Second, "Period between leadership attempts")
	)
	flag.Parse()

	if *nodeID == "" {
		log.Fatal("node-id is required. Use -node-id flag to specify a unique identifier")
	}

	log.Printf("Starting leader election demo for node: %s", *nodeID)
	log.Printf("Lock name: %s, Lock directory: %s", *lockName, *lockDir)
	log.Printf("Lease duration: %v, Retry period: %v", *leaseDuration, *retryPeriod)

	config := leaderelection.LeaseConfig{
		LeaseDuration: *leaseDuration,
		RetryPeriod:   *retryPeriod,
		LockName:      *lockName,
		Identity:      *nodeID,
		LockDir:       *lockDir,
	}

	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func() {
			log.Printf("🎉 [%s] BECAME LEADER - Starting leadership duties", *nodeID)
		},
		OnStoppedLeading: func() {
			log.Printf("😞 [%s] LOST LEADERSHIP - Stepping down", *nodeID)
		},
	}

	elector := leaderelection.NewLeaderElector(config, callbacks)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go elector.Run(ctx)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				status := "FOLLOWER"
				emoji := "👥"
				if elector.IsLeader() {
					status = "LEADER"
					emoji = "👑"
				}
				log.Printf("%s [%s] Status: %s - Heartbeat at %s", 
					emoji, *nodeID, status, time.Now().Format("15:04:05"))
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("✅ [%s] Leader election started. Press Ctrl+C to stop", *nodeID)
	log.Printf("💡 Tip: Run multiple instances with different node-ids to see leader election in action")
	log.Printf("   Example: go run cmd/main.go -node-id=node-1")
	log.Printf("   Example: go run cmd/main.go -node-id=node-2")
	log.Printf("   Example: go run cmd/main.go -node-id=node-3")

	<-sigChan
	log.Printf("🛑 [%s] Received shutdown signal, stopping...", *nodeID)

	elector.Stop()
	cancel()

	time.Sleep(100 * time.Millisecond)
	log.Printf("👋 [%s] Shutdown complete", *nodeID)
}
