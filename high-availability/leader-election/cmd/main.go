package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cshep4/resiliency-patterns/examples/high-availability/leader-election/internal/leaderelection"
)

const (
	LeaseDuration = 10 * time.Second
	RetryPeriod   = 2 * time.Second
	LockName      = "leader-election-demo"
	LockDir       = "/tmp"
)

func main() {
	nodeID := generateNodeID()

	log.Printf("Starting leader election demo for node: %s", nodeID)
	log.Printf("Lock name: %s, Lock directory: %s", LockName, LockDir)
	log.Printf("Lease duration: %v, Retry period: %v", LeaseDuration, RetryPeriod)

	config := leaderelection.LeaseConfig{
		LeaseDuration: LeaseDuration,
		RetryPeriod:   RetryPeriod,
		LockName:      LockName,
		Identity:      nodeID,
		LockDir:       LockDir,
	}

	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func() {
			log.Printf("🎉 [%s] BECAME LEADER - Starting leadership duties", nodeID)
		},
		OnStoppedLeading: func() {
			log.Printf("😞 [%s] LOST LEADERSHIP - Stepping down", nodeID)
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
					emoji, nodeID, status, time.Now().Format("15:04:05"))
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("✅ [%s] Leader election started. Press Ctrl+C to stop", nodeID)
	log.Printf("💡 Tip: Run multiple instances to see leader election in action")

	<-sigChan
	log.Printf("🛑 [%s] Received shutdown signal, stopping...", nodeID)

	elector.Stop()
	cancel()

	time.Sleep(100 * time.Millisecond)
	log.Printf("👋 [%s] Shutdown complete", nodeID)
}

func generateNodeID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	
	pid := os.Getpid()
	timestamp := time.Now().Unix()
	
	return fmt.Sprintf("node-%s-%d-%d", hostname, pid, timestamp)
}
