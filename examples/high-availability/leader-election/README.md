# Leader Election Example

A practical implementation of leader election pattern using file-based locking for distributed systems coordination.

## Overview

This example demonstrates how to implement leader election in a distributed system where multiple instances compete to become the leader. Only one instance can be the leader at any given time, and if the leader fails, another instance automatically takes over.

## How It Works

### Leader Election Algorithm

1. **Lease-based Leadership**: Uses file-based locking with lease expiration
2. **Heartbeat Mechanism**: Leader periodically renews its lease
3. **Automatic Failover**: If leader fails to renew lease, others can acquire leadership
4. **Conflict Resolution**: Only one instance can hold the lock file at a time

### Key Components

- **LeaderElector**: Main coordination logic
- **Lease Management**: Handles lock acquisition, renewal, and expiration
- **Callbacks**: Notifications when leadership status changes
- **Heartbeat Logging**: Visual indication of current leader status

## Usage

### Quick Demo

```bash
# Start the demo (single instance)
make demo

# Or run a specific node
make run NODE_ID=node-1
```

### Multi-Instance Demo

**Option 1: Manual (3 separate terminals)**
```bash
# Terminal 1
make run-node1

# Terminal 2  
make run-node2

# Terminal 3
make run-node3
```

**Option 2: Automatic with tmux**
```bash
# Starts all 3 instances in tmux session
make demo-parallel

# To stop all instances
make kill-demo
```

### Custom Configuration

```bash
go run cmd/main.go \
  -node-id=my-node \
  -lock-name=my-service \
  -lease-duration=15s \
  -retry-period=3s \
  -lock-dir=/tmp
```

## Example Output

```
2024/01/10 14:30:25 Starting leader election demo for node: node-1
2024/01/10 14:30:25 ✅ [node-1] Leader election started. Press Ctrl+C to stop
2024/01/10 14:30:27 [node-1] Successfully acquired leadership
2024/01/10 14:30:27 🎉 [node-1] BECAME LEADER - Starting leadership duties
2024/01/10 14:30:27 👑 [node-1] Status: LEADER - Heartbeat at 14:30:27
2024/01/10 14:30:28 👑 [node-1] Status: LEADER - Heartbeat at 14:30:28
2024/01/10 14:30:29 👑 [node-1] Status: LEADER - Heartbeat at 14:30:29
```

When you start additional instances:
```
# node-2 output
2024/01/10 14:30:30 👥 [node-2] Status: FOLLOWER - Heartbeat at 14:30:30
2024/01/10 14:30:31 👥 [node-2] Status: FOLLOWER - Heartbeat at 14:30:31

# node-3 output  
2024/01/10 14:30:32 👥 [node-3] Status: FOLLOWER - Heartbeat at 14:30:32
2024/01/10 14:30:33 👥 [node-3] Status: FOLLOWER - Heartbeat at 14:30:33
```

When the leader stops:
```
# node-1 stops, node-2 takes over
2024/01/10 14:30:45 [node-2] Successfully acquired leadership  
2024/01/10 14:30:45 🎉 [node-2] BECAME LEADER - Starting leadership duties
2024/01/10 14:30:45 👑 [node-2] Status: LEADER - Heartbeat at 14:30:45
```

## Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `-node-id` | *required* | Unique identifier for this instance |
| `-lock-name` | `leader-election-demo` | Name of the leadership lock |
| `-lock-dir` | `/tmp` | Directory to store lock files |
| `-lease-duration` | `10s` | How long leadership lease lasts |
| `-retry-period` | `2s` | How often to attempt leadership |

## Architecture

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

## Key Features

### Lease-Based Leadership
- **Automatic Expiration**: Leases expire if not renewed
- **Graceful Handover**: Clean leadership transitions
- **Fault Tolerance**: Handles node failures automatically

### Visual Feedback
- **Real-time Status**: Every second heartbeat with leader/follower status
- **Leadership Events**: Clear notifications when leadership changes
- **Emoji Indicators**: 👑 for leader, 👥 for followers

### Robust Implementation
- **Race Condition Handling**: Atomic file operations
- **Cleanup on Exit**: Proper resource cleanup
- **Configurable Timing**: Adjustable lease and retry periods

## Testing

```bash
# Run unit tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race
```

## Real-World Applications

This pattern is useful for:

- **Database Migration Leaders**: Only one instance runs migrations
- **Scheduled Job Coordination**: Prevent duplicate cron job execution
- **Cache Warming**: Single instance handles cache refresh
- **Monitoring Coordination**: One instance sends alerts
- **Resource Cleanup**: Coordinated cleanup tasks

## Limitations

- **File System Dependency**: Requires shared file system
- **Not Suitable for High Frequency**: File I/O overhead
- **Local Development**: Best for development/testing scenarios

For production systems, consider:
- **etcd** with client-go leaderelection
- **Consul** leader election
- **Kubernetes** leader election
- **Database-based** coordination

## Next Steps

Try modifying the example to:
1. Add actual work simulation when leader
2. Implement different storage backends (Redis, etcd)
3. Add metrics collection for leadership duration
4. Implement graceful leadership handover
