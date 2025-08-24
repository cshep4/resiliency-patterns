package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"

	leaderelection "github.com/cshep4/resiliency-patterns/high-availability/leader-election/internal/leaderelection/file"
)

// LeaderElector is an interface that defines the methods for leader election
// implementations of the leader election pattern can be found in the
// internal/leaderelection package.
type LeaderElector interface {
	AcquireLease(ctx context.Context)
	MonitorLease(ctx context.Context, onShutdown func())
}

func main() {
	nodeID := fmt.Sprintf("%d", os.Getpid())

	log.Printf("Starting leader election demo for node: %s", nodeID)
	log.Printf("💡 Tip: Run multiple instances to see leader election in action")

	elector, err := leaderelection.NewLeaderElector(nodeID)
	if err != nil {
		log.Fatalf("Failed to create leader elector: %v", err)
	}

	log.Printf("🔍 [%s] Attempting to acquire leadership...", nodeID)

	// Create context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Block until we acquire leadership
	if err := elector.AcquireLease(ctx); err != nil {
		log.Fatalf("Failed to acquire leadership: %v", err)
	}

	log.Printf("👑 [%s] Acquired leadership", nodeID)

	// Create errgroup for managing goroutines
	g, ctx := errgroup.WithContext(ctx)

	// Start lease monitoring
	g.Go(func() error {
		elector.MonitorLease(ctx, func() {
			log.Printf("🛑 [%s] Lease lost, initiating shutdown...", nodeID)
			cancel()
		})
		return nil
	})

	log.Printf("🏁 [%s] Starting work loop...", nodeID)

	// Start worker process
	g.Go(func() error {
		return workerProcess(ctx, nodeID)
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil && err != context.Canceled {
		log.Printf("❌ [%s] Error: %v", nodeID, err)
	}

	log.Printf("👋 [%s] Shutdown complete", nodeID)
}

// workerProcess is an example of a service implementation that must be run in an active/passive manner.
func workerProcess(ctx context.Context, nodeID string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Printf("👷 [%s] Doing leader work...", nodeID)
		}
	}
}
