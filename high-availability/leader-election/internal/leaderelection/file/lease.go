// Package leaderelection provides a file-based leader election mechanism
// that allows multiple nodes to compete for leadership in a distributed system.
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
	// leaseDuration is how long a leadership lease is valid
	leaseDuration = 10 * time.Second
	// retryPeriod is how often to retry acquiring leadership
	retryPeriod = 2 * time.Second
	// lockName is the base name for the lock file
	lockName = "leader-election-demo"
	// lockDir is the directory where lock files are stored
	lockDir = "/tmp"
)

// leaderElector manages leader election using file-based locking
type leaderElector struct {
	// identity is the unique identifier for this node
	identity string
	// lockFile is the full path to the lock file used for leader election
	lockFile string
}

// NewLeaderElector creates a new leaderElector instance with the given node ID
func NewLeaderElector(nodeID string) (*leaderElector, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID is required")
	}

	// Construct the full path to the lock file
	lockFile := filepath.Join(lockDir, fmt.Sprintf("%s.lock", lockName))

	return &leaderElector{
		identity: nodeID,
		lockFile: lockFile,
	}, nil
}

// AcquireLease attempts to acquire leadership by creating a lock file
// It will block and keep retrying until successful or the context is cancelled
func (le *leaderElector) AcquireLease(ctx context.Context) error {
	log.Printf("[%s] Attempting to acquire leadership...", le.identity)

	// Try once immediately to avoid unnecessary delay
	if le.tryAcquireLease() {
		log.Printf("🎉 [%s] Successfully acquired leadership!", le.identity)
		return nil
	}

	// If not successful, use ticker for periodic retries
	ticker := time.NewTicker(retryPeriod)
	defer ticker.Stop()

	// Keep trying until we acquire leadership or context is cancelled
	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop trying
			return ctx.Err()
		case <-ticker.C:
			// Time for another attempt
			if le.tryAcquireLease() {
				log.Printf("🎉 [%s] Successfully acquired leadership!", le.identity)
				return nil
			}
		}
	}
}

// tryAcquireLease attempts to acquire the leadership lease
// Returns true if successful, false otherwise
func (le *leaderElector) tryAcquireLease() bool {
	// Check if lock file already exists
	if _, err := os.Stat(le.lockFile); err == nil {
		// Lock file exists, check if it's expired
		if !le.isLeaseExpired() {
			// Lease is still valid, cannot acquire
			return false
		}
		log.Printf("[%s] Found expired lease, attempting to acquire", le.identity)
	}

	// Try to create the lock file atomically using O_EXCL
	// This ensures only one process can create the file
	file, err := os.OpenFile(le.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		// Failed to create file (likely already exists)
		return false
	}
	defer file.Close()

	// Write our identity and timestamp to the lock file
	leaseData := fmt.Sprintf("%s:%d", le.identity, time.Now().Unix())
	if _, err := file.WriteString(leaseData); err != nil {
		// Failed to write data, clean up the file
		os.Remove(le.lockFile)
		return false
	}

	// Successfully acquired the lease
	return true
}

// isLeaseExpired checks if the current lease has expired
// Returns true if expired or if there's any error reading the lease
func (le *leaderElector) isLeaseExpired() bool {
	// Try to read the lock file
	data, err := os.ReadFile(le.lockFile)
	if err != nil {
		// Cannot read file, consider it expired
		return true
	}

	// Parse the lease data format: "identity:timestamp"
	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		// Invalid format, consider it expired
		return true
	}

	// Parse the timestamp
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		// Invalid timestamp, consider it expired
		return true
	}

	// Check if the lease duration has passed
	leaseTime := time.Unix(timestamp, 0)
	return time.Since(leaseTime) > leaseDuration
}

// MonitorLease continuously monitors the leadership status and renews the lease
// Calls onShutdown if leadership is lost and cleans up the lock file
func (le *leaderElector) MonitorLease(ctx context.Context, onShutdown func()) {
	// Check lease status every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Printf("[%s] Starting lease monitoring...", le.identity)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop monitoring and clean up
			log.Printf("[%s] Lease monitoring stopped", le.identity)
			err := os.Remove(le.lockFile)
			if err != nil {
				log.Printf("[%s] Error removing lock file: %v", le.identity, err)
			}
			return
		case <-ticker.C:
			// Regular lease check
			if !le.isCurrentLeader() {
				// We're no longer the leader, shut down gracefully
				log.Printf("🚨 [%s] Lease lost! Shutting down...", le.identity)
				onShutdown()

				// Clean up the lock file
				err := os.Remove(le.lockFile)
				if err != nil {
					log.Printf("[%s] Error removing lock file: %v", le.identity, err)
				}
				return
			}

			// Renew the lease if it's time to do so
			if le.shouldRenewLease() {
				if err := le.renewLease(); err != nil {
					log.Printf("[%s] Failed to renew lease: %v", le.identity, err)
				}
			}
		}
	}
}

// isCurrentLeader checks if this node is currently the leader
// Returns true if we own the lease and it's still valid
func (le *leaderElector) isCurrentLeader() bool {
	// Read the current lock file
	data, err := os.ReadFile(le.lockFile)
	if err != nil {
		// Cannot read file, we're not the leader
		return false
	}

	// Parse the lease data format: "identity:timestamp"
	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		// Invalid format, we're not the leader
		return false
	}

	// Check if we own the lease
	if parts[0] != le.identity {
		// Someone else owns the lease
		return false
	}

	// Check if our lease is still valid (not expired)
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		// Invalid timestamp, we're not the leader
		return false
	}

	leaseTime := time.Unix(timestamp, 0)
	return time.Since(leaseTime) <= leaseDuration
}

// shouldRenewLease determines if it's time to renew the leadership lease
// Returns true if we should renew (when halfway through lease duration)
func (le *leaderElector) shouldRenewLease() bool {
	// Read the current lock file to get the last renewal time
	data, err := os.ReadFile(le.lockFile)
	if err != nil {
		// Cannot read file, cannot renew
		return false
	}

	// Parse the lease data format: "identity:timestamp"
	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		// Invalid format, cannot renew
		return false
	}

	// Parse the last renewal timestamp
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		// Invalid timestamp, cannot renew
		return false
	}

	// Calculate time since last renewal
	leaseTime := time.Unix(timestamp, 0)
	timeSinceRenewal := time.Since(leaseTime)

	// Renew when we're halfway through the lease duration
	// This provides a safety margin before the lease expires
	return timeSinceRenewal > leaseDuration/2
}

// renewLease updates the lease timestamp to extend our leadership
// Returns an error if the renewal fails
func (le *leaderElector) renewLease() error {
	// Create new lease data with current timestamp
	leaseData := fmt.Sprintf("%s:%d", le.identity, time.Now().Unix())
	// Atomically update the lock file with new timestamp
	return os.WriteFile(le.lockFile, []byte(leaseData), 0644)
}
