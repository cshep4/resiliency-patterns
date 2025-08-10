package leaderelection

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLeaderElector(t *testing.T) {
	nodeID := "test-node-1"
	le := NewLeaderElector(nodeID)

	if le.identity != nodeID {
		t.Errorf("Expected identity %s, got %s", nodeID, le.identity)
	}

	expectedLockFile := filepath.Join(LockDir, LockName+".lock")
	if le.lockFile != expectedLockFile {
		t.Errorf("Expected lock file %s, got %s", expectedLockFile, le.lockFile)
	}
}

func TestTryAcquireLease(t *testing.T) {
	tmpDir := t.TempDir()
	nodeID := "test-node-1"
	
	// Create a custom elector with temp directory
	lockFile := filepath.Join(tmpDir, "test-lock.lock")
	le := &LeaderElector{
		identity: nodeID,
		lockFile: lockFile,
	}

	if !le.tryAcquireLease() {
		t.Error("Expected to acquire lease on first attempt")
	}

	// Second attempt should fail since lease is held
	le2 := &LeaderElector{
		identity: "test-node-2",
		lockFile: lockFile,
	}
	
	if le2.tryAcquireLease() {
		t.Error("Expected to fail acquiring lease when already held")
	}

	os.Remove(lockFile)
}

func TestIsLeaseExpired(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "test-lock.lock")
	
	le := &LeaderElector{
		identity: "test-node-1",
		lockFile: lockFile,
	}

	// Non-existent file should be considered expired
	if !le.isLeaseExpired() {
		t.Error("Expected lease to be expired when file doesn't exist")
	}

	// Acquire lease
	if !le.tryAcquireLease() {
		t.Fatal("Failed to acquire lease")
	}

	// Should not be expired immediately after acquisition
	if le.isLeaseExpired() {
		t.Error("Expected lease to not be expired immediately after acquisition")
	}

	// Manually create an expired lease
	expiredTime := time.Now().Add(-LeaseDuration - time.Second).Unix()
	leaseData := fmt.Sprintf("test-node-1:%d", expiredTime)
	os.WriteFile(lockFile, []byte(leaseData), 0644)

	if !le.isLeaseExpired() {
		t.Error("Expected lease to be expired after lease duration")
	}

	os.Remove(lockFile)
}

func TestAcquireLeaseBlocking(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "test-lock.lock")
	
	// First elector acquires lease
	le1 := &LeaderElector{
		identity: "test-node-1",
		lockFile: lockFile,
	}
	
	// Manually acquire lease to simulate blocking scenario
	le1.tryAcquireLease()
	
	// Verify lease is held (file exists and not expired)
	if _, err := os.Stat(lockFile); err != nil {
		t.Error("Expected lease file to exist")
	}
	
	// Second elector should not be able to acquire immediately
	le2 := &LeaderElector{
		identity: "test-node-2", 
		lockFile: lockFile,
	}
	
	if le2.tryAcquireLease() {
		t.Error("Expected second elector to fail acquiring active lease")
	}

	os.Remove(lockFile)
}
