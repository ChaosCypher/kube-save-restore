# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build
```bash
# Build binary
go build -o kube-save-restore

# Cross-compile for specific platform
GOOS=linux GOARCH=amd64 go build -o kube-save-restore-linux-amd64
```

### Test
```bash
# Run unit tests with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# Run integration tests (requires Kubernetes cluster)
go test -tags=integration -v ./internal/integration

# Run specific test
go test -v -run TestBackupManager ./internal/backup
```

### Lint and Format
```bash
# Format code
go fmt ./...
gofmt -s -l .

# Run static analysis
go vet ./...

# Run golangci-lint (as used in CI)
golangci-lint run --timeout=5m
```

## Architecture Overview

kube-save-restore is a Kubernetes backup/restore tool with concurrent processing capabilities. The codebase follows a modular architecture:

### Core Flow
1. **Entry Point** (`main.go`): Parses config, creates K8s client, routes to backup/restore
2. **Configuration** (`internal/config/`): Handles flags/env vars, validates settings
3. **Operations**:
   - **Backup** (`internal/backup/`): Concurrent resource collection and JSON serialization
   - **Restore** (`internal/restore/`): Worker pool-based parallel restoration
4. **Kubernetes API** (`internal/kubernetes/`): Abstracted client for all resource operations

### Key Design Patterns
- **Worker Pool**: Used in restore operations for concurrent processing (`internal/workerpool/`)
- **Error Groups**: Used in backup operations for concurrent resource collection
- **Interface-based Design**: Kubernetes client operations are abstracted behind interfaces
- **Structured Logging**: Thread-safe logger with configurable levels (`internal/logger/`)

### Resource Organization
Backups follow the structure: `/{backup-dir}/{namespace}/{resource-type}/{resource-name}.json`

Supported resources: Deployments, Services, ConfigMaps, Secrets, StatefulSets, HPAs, CronJobs, Jobs, PVCs

### Testing Strategy
- Unit tests alongside implementation files (*_test.go)
- Integration tests in `internal/integration/` (build tag: integration)
- CI runs both test suites with race detection

### Important Context
- No Makefile; commands are derived from CI workflow and Go conventions
- Configuration is CLI/env-based only (no config files)
- Dry-run mode available for both backup and restore operations
- Uses official Kubernetes Go client libraries (k8s.io/client-go)