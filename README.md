# Kubernetes Save & Restore

```ascii
_________ .__                        _________               .__
\_   ___ \|  |__ _____    ____  _____\_   ___ \___.__.______ |  |__   ___________
/    \  \/|  |  \\__  \  /  _ \/  ___/    \  \<   |  |\____ \|  |  \_/ __ \_  __ \
\     \___|   Y  \/ __ \(  <_> )___ \\     \___\___  ||  |_> >   Y  \  ___/|  | \/
 \______  /___|  (____  /\____/____  >\______  / ____||   __/|___|  /\___  >__|
        \/     \/     \/           \/        \/\/     |__|        \/     \/
```

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
  - [Backup](#backup)
  - [Restore](#restore)
- [Configuration](#configuration)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

## Introduction

Kubernetes Backup & Restore is a robust and user-friendly tool designed to simplify the process of backing up and restoring Kubernetes resources. Whether you're managing a small cluster or a large-scale Kubernetes deployment, this tool ensures your configurations, deployments, services, config maps, and secrets are securely backed up and easily recoverable.

![K8s](https://kubernetes.io/images/kubernetes-horizontal-color.png)

## Features

- **Comprehensive Backup**: Capture deployments, services, config maps, secrets, and more across all namespaces.
- **Seamless Restore**: Restore your Kubernetes resources with ease, ensuring minimal downtime.
- **Dry Run Mode**: Validate backup and restore operations without making actual changes.
- **Concurrent Processing**: Utilize worker pools for efficient handling of multiple resources.
- **Customizable Logging**: Configure log levels and output destinations to suit your monitoring needs.
- **Configuration Flexibility**: Easily configure via flags or environment variables.
- **Automated Testing**: Comprehensive test suite ensuring reliability and stability.

## Prerequisites

- **Go**: Version 1.23.1 or higher.
- **Kubernetes Cluster**: Access to a Kubernetes cluster with appropriate permissions.
- **kubectl**: Configured with access to your target cluster.

## Installation

### Release Assets

You can download the latest release assets from the [GitHub Releases](https://github.com/chaoscypher/k8s-backup-restore/releases) page. Follow these steps to install the `k8s-backup-restore` binary:

1. **Select the Appropriate Asset**: Choose the release asset that matches your operating system and architecture.

2. **Download the Asset**: Click on the asset to download it to your local machine.

3. **Extract the Archive**: Unzip or untar the downloaded archive to extract the `k8s-backup-restore-<os>-<arch>` binary.

4. **Rename the Binary**: Rename the extracted binary to `k8s-backup-restore`.

   ```bash
   mv k8s-backup-restore-<os>-<arch> k8s-backup-restore
   ```

5. **Move the Binary to Your PATH**: Move the renamed binary to a directory that is included in your system's `PATH`, or execute it directly from its current location.

   ```bash
   sudo mv k8s-backup-restore /usr/local/bin/
   ```

### Building from Source

1. **Clone the Repository**

   ```bash
   git clone https://github.com/yourusername/k8s-backup-restore.git
   cd k8s-backup-restore
   ```

2. **Build the Binary**

   ```bash
   go build -o k8s-backup-restore cmd/main.go
   ```

3. **Move to a Directory in PATH**

   ```bash
   sudo mv k8s-backup-restore /usr/local/bin/
   ```

## Usage

Kubernetes Backup & Restore offers two primary modes: `backup` and `restore`. Each mode comes with its own set of flags to customize the operation.

### Backup

Perform a backup of your Kubernetes resources.

```bash
k8s-backup-restore --mode=backup [flags]
```

**Example:**

```bash
k8s-backup-restore --mode=backup --backup-dir=/path/to/backup --dry-run=false --log-level=info
```

**Flags:**

- `--kubeconfig`: Path to the kubeconfig file (defaults to `$HOME/.kube/config`).
- `--context`: Kubernetes context to use.
- `--backup-dir`: Directory where backups will be stored.
- `--dry-run`: Execute a dry run without making any changes.
- `--log-level`: Logging level (`debug`, `info`, `warn`, `error`).
- `--log-file`: Path to the log file.

### Restore

Restore your Kubernetes resources from a backup.

```bash
k8s-backup-restore --mode=restore [flags]
```

**Example:**

```bash
k8s-backup-restore --mode=restore --restore-dir=/path/to/backup --dry-run=true --log-level=debug
```

**Flags:**

- `--kubeconfig`: Path to the kubeconfig file (defaults to `$HOME/.kube/config`).
- `--context`: Kubernetes context to use.
- `--restore-dir`: Directory from where backups will be restored.
- `--dry-run`: Execute a dry run without making any changes.
- `--log-level`: Logging level (`debug`, `info`, `warn`, `error`).
- `--log-file`: Path to the log file.

## Configuration

You can configure Kubernetes Backup & Restore using command-line flags or environment variables. Environment variables take precedence over flags.

| Flag            | Environment Variable | Description                                      |
| --------------- | -------------------- | ------------------------------------------------ |
| `--kubeconfig`  | `KUBECONFIG`         | Path to the kubeconfig file.                     |
| `--context`     | `KUBE_CONTEXT`       | Kubernetes context to use.                       |
| `--backup-dir`  | `BACKUP_DIR`         | Directory where backups will be stored.          |
| `--restore-dir` | `RESTORE_DIR`        | Directory from where backups will be restored.   |
| `--mode`        | `MODE`               | Operation mode: `backup` or `restore`.           |
| `--dry-run`     | `DRY_RUN`            | Execute a dry run without making any changes.    |
| `--log-level`   | `LOG_LEVEL`          | Logging level: `debug`, `info`, `warn`, `error`. |
| `--log-file`    | `LOG_FILE`           | Path to the log file.                            |

**Example using Environment Variables:**

```bash
export MODE=backup
export BACKUP_DIR=/path/to/backup
export LOG_LEVEL=info
k8s-backup-restore
```

## Logging

Kubernetes Backup & Restore provides flexible logging options to help you monitor and debug operations.

- **Log Levels**:

  - `DEBUG`: Detailed information, typically of interest only when diagnosing problems.
  - `INFO`: Confirmation that things are working as expected.
  - `WARN`: An indication that something unexpected happened, or indicative of some problem in the near future.
  - `ERROR`: Due to a more serious problem, the software has not been able to perform some function.

- **Log Output**:
  - **Standard Output**: By default, logs are written to `stdout`.
  - **Log File**: You can specify a log file using the `--log-file` flag.

## Testing

The project includes comprehensive tests covering various components.

### Run Unit Tests

```bash
go test -v ./...
```

### Run Integration Tests

```bash
export TEST_KUBECONFIG=<>
go test -tags=integration -v ./...
```

### Generate Coverage Report

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the Repository**
2. **Create a Feature Branch**

   ```bash
   git checkout -b feature/YourFeature
   ```

3. **Commit Your Changes**

   ```bash
   git commit -m "Add your feature"
   ```

4. **Push to the Branch**

   ```bash
   git push origin feature/YourFeature
   ```

5. **Open a Pull Request**

Ensure that all tests pass and adhere to the project’s coding standards.

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

For any inquiries or support, please open an issue on the [GitHub repository](https://github.com/chaoscypher/k8s-backup-restore).
