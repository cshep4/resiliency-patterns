package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	lockFile      = "leader.lock"
	leaseDuration = 10 * time.Second
)

func main() {
	id := os.Getenv("NODE_ID") // Each replica should have a unique NODE_ID

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	for {
		// Try to acquire leadership lease (exclusive lock via file create)
		f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			if _, err = f.WriteString(id); err != nil {
				// Handle error & release lock...
			}

			// We acquired the lock, run as leader...
			runLeaderProcess(ctx)

			return
		}

		// If os.OpenFile errors, run in follower mode: wait for lease duration + retry
		select {
		case <-time.After(leaseDuration):
			// Retry acquiring the lock
		case <-ctx.Done():
			// Exit the application...
			return
		}
	}
}

func runLeaderProcess(ctx context.Context) {
	leaderTicker := time.NewTicker(1 * time.Second)

	run := true
	for run {
		select {
		case <-leaderTicker.C:
			log.Println("[", id, "]", "leader: doing work…")
		case <-ctx.Done():
			run = false
		}
	}

	leaderTicker.Stop()
	f.Close()
	_ = os.Remove(lockFile)
	log.Println("[", id, "]", "released leadership, exiting")
	return
}
