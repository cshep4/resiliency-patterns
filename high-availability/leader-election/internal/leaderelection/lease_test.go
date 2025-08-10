package leaderelection

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLeaderElector(t *testing.T) {
	config := LeaseConfig{
		LockName: "test-lock",
		Identity: "test-node-1",
	}
	callbacks := LeaderCallbacks{}

	le := NewLeaderElector(config, callbacks)

	if le.config.LeaseDuration != DefaultLeaseDuration {
		t.Errorf("Expected default lease duration %v, got %v", DefaultLeaseDuration, le.config.LeaseDuration)
	}
	if le.config.RenewDeadline != DefaultRenewDeadline {
		t.Errorf("Expected default renew deadline %v, got %v", DefaultRenewDeadline, le.config.RenewDeadline)
	}
	if le.config.RetryPeriod != DefaultRetryPeriod {
		t.Errorf("Expected default retry period %v, got %v", DefaultRetryPeriod, le.config.RetryPeriod)
	}
}

func TestTryAcquireLease(t *testing.T) {
	tmpDir := t.TempDir()
	config := LeaseConfig{
		LockName: "test-lock",
		Identity: "test-node-1",
		LockDir:  tmpDir,
	}
	callbacks := LeaderCallbacks{}

	le := NewLeaderElector(config, callbacks)

	if !le.tryAcquireLease() {
		t.Error("Expected to acquire lease on first attempt")
	}

	if le.tryAcquireLease() {
		t.Error("Expected to fail acquiring lease when already held")
	}

	os.Remove(le.lockFile)
}

func TestIsLeaseExpired(t *testing.T) {
	tmpDir := t.TempDir()
	config := LeaseConfig{
		LockName:      "test-lock",
		Identity:      "test-node-1",
		LockDir:       tmpDir,
		LeaseDuration: 100 * time.Millisecond,
	}
	callbacks := LeaderCallbacks{}

	le := NewLeaderElector(config, callbacks)

	if le.isLeaseExpired() {
		t.Error("Expected lease to not be expired when file doesn't exist")
	}

	le.tryAcquireLease()

	if le.isLeaseExpired() {
		t.Error("Expected lease to not be expired immediately after acquisition")
	}

	time.Sleep(150 * time.Millisecond)

	if !le.isLeaseExpired() {
		t.Error("Expected lease to be expired after lease duration")
	}

	os.Remove(le.lockFile)
}

func TestLeaderElection(t *testing.T) {
	tmpDir := t.TempDir()
	
	var leader1Started, leader1Stopped bool
	var leader2Started, leader2Stopped bool

	config1 := LeaseConfig{
		LockName:      "test-lock",
		Identity:      "node-1",
		LockDir:       tmpDir,
		LeaseDuration: 200 * time.Millisecond,
		RetryPeriod:   50 * time.Millisecond,
	}
	callbacks1 := LeaderCallbacks{
		OnStartedLeading: func() { leader1Started = true },
		OnStoppedLeading: func() { leader1Stopped = true },
	}

	config2 := LeaseConfig{
		LockName:      "test-lock",
		Identity:      "node-2",
		LockDir:       tmpDir,
		LeaseDuration: 200 * time.Millisecond,
		RetryPeriod:   50 * time.Millisecond,
	}
	callbacks2 := LeaderCallbacks{
		OnStartedLeading: func() { leader2Started = true },
		OnStoppedLeading: func() { leader2Stopped = true },
	}

	le1 := NewLeaderElector(config1, callbacks1)
	le2 := NewLeaderElector(config2, callbacks2)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go le1.Run(ctx)
	go le2.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	leadersCount := 0
	if le1.IsLeader() {
		leadersCount++
	}
	if le2.IsLeader() {
		leadersCount++
	}

	if leadersCount != 1 {
		t.Errorf("Expected exactly 1 leader, got %d", leadersCount)
	}

	le1.Stop()
	le2.Stop()

	time.Sleep(100 * time.Millisecond)

	os.Remove(filepath.Join(tmpDir, "test-lock.lock"))
}
