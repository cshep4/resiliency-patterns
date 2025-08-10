package leaderelection

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DefaultLeaseDuration = 10 * time.Second
	DefaultRenewDeadline = 8 * time.Second
	DefaultRetryPeriod   = 2 * time.Second
)

type LeaseConfig struct {
	LeaseDuration time.Duration
	RenewDeadline time.Duration
	RetryPeriod   time.Duration
	LockName      string
	Identity      string
	LockDir       string
}

type LeaderElector struct {
	config     LeaseConfig
	isLeader   bool
	mu         sync.RWMutex
	onStarted  func()
	onStopped  func()
	lockFile   string
	cancelFunc context.CancelFunc
}

type LeaderCallbacks struct {
	OnStartedLeading func()
	OnStoppedLeading func()
}

func NewLeaderElector(config LeaseConfig, callbacks LeaderCallbacks) *LeaderElector {
	if config.LeaseDuration == 0 {
		config.LeaseDuration = DefaultLeaseDuration
	}
	if config.RenewDeadline == 0 {
		config.RenewDeadline = DefaultRenewDeadline
	}
	if config.RetryPeriod == 0 {
		config.RetryPeriod = DefaultRetryPeriod
	}
	if config.LockDir == "" {
		config.LockDir = "/tmp"
	}

	lockFile := filepath.Join(config.LockDir, fmt.Sprintf("%s.lock", config.LockName))

	return &LeaderElector{
		config:    config,
		onStarted: callbacks.OnStartedLeading,
		onStopped: callbacks.OnStoppedLeading,
		lockFile:  lockFile,
	}
}

func (le *LeaderElector) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	le.cancelFunc = cancel

	ticker := time.NewTicker(le.config.RetryPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			le.releaseLease()
			return
		case <-ticker.C:
			le.tryAcquireOrRenewLease(ctx)
		}
	}
}

func (le *LeaderElector) IsLeader() bool {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.isLeader
}

func (le *LeaderElector) Stop() {
	if le.cancelFunc != nil {
		le.cancelFunc()
	}
}

func (le *LeaderElector) tryAcquireOrRenewLease(ctx context.Context) {
	if le.IsLeader() {
		if le.renewLease() {
			return
		}
		log.Printf("[%s] Failed to renew lease, stepping down as leader", le.config.Identity)
		le.stepDown()
	}

	if le.tryAcquireLease() {
		log.Printf("[%s] Successfully acquired leadership", le.config.Identity)
		le.becomeLeader()
	}
}

func (le *LeaderElector) tryAcquireLease() bool {
	if _, err := os.Stat(le.lockFile); err == nil {
		if !le.isLeaseExpired() {
			return false
		}
		log.Printf("[%s] Found expired lease, attempting to acquire", le.config.Identity)
	}

	file, err := os.OpenFile(le.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer file.Close()

	leaseData := fmt.Sprintf("%s:%d", le.config.Identity, time.Now().Unix())
	if _, err := file.WriteString(leaseData); err != nil {
		os.Remove(le.lockFile)
		return false
	}

	return true
}

func (le *LeaderElector) renewLease() bool {
	if _, err := os.Stat(le.lockFile); err != nil {
		return false
	}

	leaseData := fmt.Sprintf("%s:%d", le.config.Identity, time.Now().Unix())
	err := os.WriteFile(le.lockFile, []byte(leaseData), 0644)
	return err == nil
}

func (le *LeaderElector) isLeaseExpired() bool {
	data, err := os.ReadFile(le.lockFile)
	if err != nil {
		return true
	}

	var identity string
	var timestamp int64
	if _, err := fmt.Sscanf(string(data), "%s:%d", &identity, &timestamp); err != nil {
		return true
	}

	leaseTime := time.Unix(timestamp, 0)
	return time.Since(leaseTime) > le.config.LeaseDuration
}

func (le *LeaderElector) releaseLease() {
	if le.IsLeader() {
		le.stepDown()
	}
	os.Remove(le.lockFile)
}

func (le *LeaderElector) becomeLeader() {
	le.mu.Lock()
	le.isLeader = true
	le.mu.Unlock()

	if le.onStarted != nil {
		le.onStarted()
	}
}

func (le *LeaderElector) stepDown() {
	le.mu.Lock()
	wasLeader := le.isLeader
	le.isLeader = false
	le.mu.Unlock()

	if wasLeader && le.onStopped != nil {
		le.onStopped()
	}
}
