# Leader Election Example

A practical implementation of leader election pattern with multiple backend options for distributed systems coordination.

## Overview

This example demonstrates how to implement leader election in a distributed system where multiple instances compete to become the leader. Only one instance can be the leader at any given time, and if the leader fails, another instance automatically takes over.

The implementation provides two different backend options:
- **[File-based](/high-availability/leader-election/internal/leaderelection/file/)** (default) - Uses the local file system locking for local development and testing
- **[Kubernetes-based](/high-availability/leader-election/internal/leaderelection/kubernetes/)** - Uses Kubernetes Leases for production deployments

## How It Works

### Leader Election Algorithm

1. **Lease-based Leadership**: Uses distributed locking with lease expiration
2. **Heartbeat Mechanism**: Leader periodically renews its lease
3. **Automatic Failover**: If leader fails to renew lease, others can acquire leadership
4. **Conflict Resolution**: Only one instance can hold the lock at a time

### Key Components

- **File-based Implementation**: Simple file locking for development
- **Kubernetes Implementation**: Production-ready using Kubernetes Leases
- **Lease Management**: Handles lock acquisition, renewal, and expiration
- **Callbacks**: Notifications when leadership status changes
- **Heartbeat Logging**: Visual indication of current leader status

## Usage

### Quick Demo (File-based)

```bash
# Start the demo (single instance)
make demo

# Or run directly
make run
```

### Multi-Instance Demo

**Option 1: Manual (3 separate terminals)**
```bash
# Terminal 1
make run

# Terminal 2  
make run

# Terminal 3
make run
```

**Option 2: Automatic with tmux**
```bash
# Starts all 3 instances in tmux session
make demo-parallel

# To stop all instances
make kill-demo
```

## Example Output

```
2024/01/10 14:30:25 Starting leader election demo for node: 12345
2024/01/10 14:30:25 ✅ [12345] Leader election started. Press Ctrl+C to stop
2024/01/10 14:30:27 [12345] Successfully acquired leadership
2024/01/10 14:30:27 🎉 [12345] BECAME LEADER - Starting leadership duties
2024/01/10 14:30:27 👑 [12345] Status: LEADER - Heartbeat at 14:30:27
2024/01/10 14:30:28 👑 [12345] Status: LEADER - Heartbeat at 14:30:28
```

When you start additional instances:
```
# Second instance output
2024/01/10 14:30:30  [67890] Status: FOLLOWER - Heartbeat at 14:30:30
2024/01/10 14:30:31  [67890] Status: FOLLOWER - Heartbeat at 14:30:31

# Third instance output  
2024/01/10 14:30:32  [11111] Status: FOLLOWER - Heartbeat at 14:30:32
2024/01/10 14:30:33  [11111] Status: FOLLOWER - Heartbeat at 14:30:33
```

When the leader stops:
```
# Leader stops, follower takes over
2024/01/10 14:30:45 [67890] Successfully acquired leadership  
2024/01/10 14:30:45 🎉 [67890] BECAME LEADER - Starting leadership duties
2024/01/10 14:30:45 👑 [67890] Status: LEADER - Heartbeat at 14:30:45
```

## Architecture

### File-based Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Node 1    │    │   Node 2    │    │   Node 3    │
│  (Leader)   │    │ (Follower)  │    │ (Follower)  │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌─────────────┐
                    │ Lock File   │
                    │ /tmp/       │
                    │ demo.lock   │
                    └─────────────┘
```

### Kubernetes Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Node 1    │    │   Node 2    │    │   Node 3    │
│  (Leader)   │    │ (Follower)  │    │ (Follower)  │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌─────────────┐
                    │ Kubernetes  │
                    │ Lease       │
                    │ Resource    │
                    └─────────────┘
```

## Real-World Applications

This pattern is useful for:

- **Database Migration Leaders**: Only one instance runs migrations
- **Scheduled Job Coordination**: Prevent duplicate cron job execution
- **Cache Warming**: Single instance handles cache refresh
- **Monitoring Coordination**: One instance sends alerts
- **Resource Cleanup**: Coordinated cleanup tasks
- **Kubernetes Operators**: Only one replica reconciles resources

## Implementation Details

### Interface Design

Both implementations follow the same `LeaderElector` interface:

```go
type LeaderElector interface {
    AcquireLease(ctx context.Context) error
    MonitorLease(ctx context.Context, onShutdown func())
}
```

### File-based Implementation

- Uses atomic file creation (`O_EXCL`) for lock acquisition
- Stores lease data as `identity:timestamp` in lock file
- Checks lease expiration by comparing timestamps
- Renews lease by updating the lock file timestamp

### Kubernetes Implementation

- Uses `k8s.io/client-go/tools/leaderelection` package
- Creates Kubernetes Lease resources for coordination
- Leverages Kubernetes API server for distributed locking
- Provides callback-based leadership notifications

## Limitations

### File-based Implementation

- **File System Dependency**: Requires shared file system
- **Not Suitable for High Frequency**: File I/O overhead
- **Local Development**: Best for development/testing scenarios
- **Single Node**: Limited to single machine deployments

### Kubernetes Implementation

- **Kubernetes Dependency**: Only works in Kubernetes environments
- **API Server Dependency**: Requires access to Kubernetes API server
- **Network Overhead**: Each lease operation requires API call
- **No Fencing Guarantees**: See note about fencing below

## Important Notes

### Fencing (Kubernetes Implementation)

The Kubernetes implementation does not provide fencing guarantees. In rare cases, you might have multiple leaders running simultaneously if:

1. Leader is paused/blocked and lease expires
2. New leader is elected
3. Original leader resumes and continues as leader

To mitigate this:
- Use short lease durations for faster failover
- Implement application-level fencing if needed
- Monitor for multiple active leaders

### Best Practices

1. **Use unique identities**: Ensure each replica has a unique node ID
2. **Handle callbacks properly**: Implement proper cleanup in shutdown callbacks
3. **Monitor lease status**: Watch for lease transitions and failures
4. **Set appropriate timeouts**: Balance between failover speed and stability
5. **Use namespaces**: Isolate leader election locks by namespace (Kubernetes)