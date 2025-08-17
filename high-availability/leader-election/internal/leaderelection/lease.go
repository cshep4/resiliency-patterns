package leaderelection

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	LeaseDuration = 10 * time.Second
	RetryPeriod   = 2 * time.Second
	LockName      = "leader-election-demo"
	LockDir       = "/tmp"
)

type LeaderElector struct {
	identity string
	lockFile string
}

func NewLeaderElector(nodeID string) *LeaderElector {
	lockFile := filepath.Join(LockDir, fmt.Sprintf("%s.lock", LockName))

	return &LeaderElector{
		identity: nodeID,
		lockFile: lockFile,
	}
}

func (le *LeaderElector) AcquireLease() {
	log.Printf("[%s] Attempting to acquire leadership...", le.identity)

	// Try once immediately
	if le.tryAcquireLease() {
		log.Printf("🎉 [%s] Successfully acquired leadership!", le.identity)
		return
	}

	// If not successful, use ticker for retries
	ticker := time.NewTicker(RetryPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if le.tryAcquireLease() {
			log.Printf("🎉 [%s] Successfully acquired leadership!", le.identity)
			return
		}
	}
}

func (le *LeaderElector) tryAcquireLease() bool {
	if _, err := os.Stat(le.lockFile); err == nil {
		if !le.isLeaseExpired() {
			return false
		}
		log.Printf("[%s] Found expired lease, attempting to acquire", le.identity)
	}

	file, err := os.OpenFile(le.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer file.Close()

	leaseData := fmt.Sprintf("%s:%d", le.identity, time.Now().Unix())
	if _, err := file.WriteString(leaseData); err != nil {
		os.Remove(le.lockFile)
		return false
	}

	return true
}

func (le *LeaderElector) isLeaseExpired() bool {
	data, err := os.ReadFile(le.lockFile)
	if err != nil {
		return true
	}

	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		return true
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return true
	}

	leaseTime := time.Unix(timestamp, 0)
	return time.Since(leaseTime) > LeaseDuration
}

func (le *LeaderElector) MonitorLease(ctx context.Context, onShutdown func()) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Printf("[%s] Starting lease monitoring...", le.identity)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Lease monitoring stopped", le.identity)
			return
		case <-ticker.C:
			if !le.isCurrentLeader() {
				log.Printf("🚨 [%s] Lease lost! Shutting down...", le.identity)
				onShutdown()
				return
			}
		}
	}
}

func (le *LeaderElector) isCurrentLeader() bool {
	data, err := os.ReadFile(le.lockFile)
	if err != nil {
		return false
	}

	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		return false
	}

	// Check if we own the lease
	if parts[0] != le.identity {
		return false
	}

	// Check if lease is still valid
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}

	leaseTime := time.Unix(timestamp, 0)
	return time.Since(leaseTime) <= LeaseDuration
}
