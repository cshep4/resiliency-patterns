package leaderelection

import (
	"context"
	"fmt"
	"log"
	"time"

	"k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// leaseDuration is how long a leadership lease is valid
	leaseDuration = 10 * time.Second
	// retryPeriod is how often to retry acquiring leadership
	retryPeriod = 2 * time.Second
	// lockName is the base name for the lock file
	lockName = "leader-election-demo"
)

type leaderElector struct {
	// identity is the unique identifier for this node
	identity string
	// lockNamespace is the namespace where the lock is created
	lockNamespace string
	// leaderElector is the Kubernetes leader election instance
	elector *leaderelection.LeaderElector

	// leadershipLost is a channel to signal when leadership is lost
	leadershipLost chan struct{}
	// leadershipGained is a channel to signal when leadership is gained
	leadershipGained chan struct{}
}

func NewLeaderElector(nodeID, lockNamespace string) (*leaderElector, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID is required")
	}

	// Get the active kubernetes context
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	// Create a new lock using Kubernetes Leases
	l, err := rl.NewFromKubeconfig(
		rl.LeasesResourceLock,
		lockNamespace,
		lockName,
		rl.ResourceLockConfig{
			Identity: nodeID,
		},
		cfg,
		time.Second*10,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource lock: %w", err)
	}

	leadershipLost := make(chan struct{})
	leadershipGained := make(chan struct{})

	// Create a new leader election configuration
	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          l,
		LeaseDuration: leaseDuration,
		RenewDeadline: leaseDuration / 2,
		RetryPeriod:   retryPeriod,
		Name:          lockName,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Printf(" [%s] BECAME LEADER - Starting leadership duties", nodeID)
				leadershipGained <- struct{}{}
			},
			OnStoppedLeading: func() {
				log.Printf("🚨 [%s] LEADERSHIP LOST - Stopping leadership duties", nodeID)
				leadershipLost <- struct{}{}
			},
			OnNewLeader: func(identity string) {
				log.Printf("👥 [%s] New leader elected: %s", nodeID, identity)
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create leader elector: %w", err)
	}

	return &leaderElector{
		identity:         nodeID,
		lockNamespace:    lockNamespace,
		elector:          elector,
		leadershipLost:   leadershipLost,
		leadershipGained: leadershipGained,
	}, nil
}

// AcquireLease attempts to acquire leadership
// It will block and keep retrying until successful or the context is cancelled
func (le *leaderElector) AcquireLease(ctx context.Context) error {
	log.Printf("[%s] Attempting to acquire leadership using Kubernetes leader election...", le.identity)

	// Start the leader election process
	// This will block until we become leader or context is cancelled
	go le.elector.Run(ctx)

	// Wait for leadership to be gained
	select {
	case <-le.leadershipGained:
		log.Printf("👑 [%s] Leadership gained", le.identity)
	case <-ctx.Done():
		log.Printf("[%s] Context cancelled, stopping leadership acquisition", le.identity)
		return ctx.Err()
	}

	return nil
}

// MonitorLease continuously monitors the leadership status and renews the lease
// Calls onShutdown if leadership is lost and cleans up the lock
func (le *leaderElector) MonitorLease(ctx context.Context, onShutdown func()) {
	log.Printf("[%s] Starting lease monitoring...", le.identity)

	// Monitor for leadership loss
	select {
	case <-le.leadershipLost:
		log.Printf("🛑 [%s] Leadership lost, calling shutdown callback", le.identity)
		onShutdown()
	case <-ctx.Done():
		log.Printf("[%s] Context cancelled, stopping lease monitoring", le.identity)
	}
}
