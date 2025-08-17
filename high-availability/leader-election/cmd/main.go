package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cshep4/resiliency-patterns/high-availability/leader-election/internal/leaderelection"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func main() {
	nodeID := "node-" + uuid.New().String()

	log.Printf("Starting leader election demo for node: %s", nodeID)
	log.Printf("💡 Tip: Run multiple instances to see leader election in action")

	elector := leaderelection.NewLeaderElector(nodeID)

	// Block until we acquire leadership
	elector.AcquireLease()

	// Create context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

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

	// Start heartbeat loop
	g.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				log.Printf("👑 [%s] Status: LEADER - Heartbeat at %s",
					nodeID, time.Now().Format("15:04:05"))
			}
		}
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		log.Printf("❌ [%s] Error: %v", nodeID, err)
	}

	log.Printf("👋 [%s] Shutdown complete", nodeID)
}
