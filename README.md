# Kubernetes Backup & Restore

```ascii
 __        ___.                                                                                  __                        
|  | ____ _\_ |__   ____             ___________ ___  __ ____           _______   ____   _______/  |_  ___________   ____  
|  |/ /  |  \ __ \_/ __ \   ______  /  ___/\__  \\  \/ // __ \   ______ \_  __ \_/ __ \ /  ___/\   __\/  _ \_  __ \_/ __ \ 
|    <|  |  / \_\ \  ___/  /_____/  \___ \  / __ \\   /\  ___/  /_____/  |  | \/\  ___/ \___ \  |  | (  <_> )  | \/\  ___/ 
|__|_ \____/|___  /\___  >         /____  >(____  /\_/  \___  >          |__|    \___  >____  > |__|  \____/|__|    \___  >
     \/         \/     \/               \/      \/          \/                       \/     \/                          \/ 
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

- **Go**: Version 1.22 or higher.
- **Kubernetes Cluster**: Access to a Kubernetes cluster with appropriate permissions.
- **kubectl**: Configured with access to your target cluster.

## Installation

### Using Go

```bash
go install github.com/chaoscypher/k8s-backup-restore@latest
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

| Flag               | Environment Variable | Description                                      |
|--------------------|-----------------------|--------------------------------------------------|
| `--kubeconfig`     | `KUBECONFIG`          | Path to the kubeconfig file.                     |
| `--context`        | `KUBE_CONTEXT`        | Kubernetes context to use.                       |
| `--backup-dir`     | `BACKUP_DIR`          | Directory where backups will be stored.          |
| `--restore-dir`    | `RESTORE_DIR`         | Directory from where backups will be restored.    |
| `--mode`           | `MODE`                | Operation mode: `backup` or `restore`.           |
| `--dry-run`        | `DRY_RUN`             | Execute a dry run without making any changes.     |
| `--log-level`      | `LOG_LEVEL`           | Logging level: `debug`, `info`, `warn`, `error`. |
| `--log-file`       | `LOG_FILE`            | Path to the log file.                             |

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

### Run Tests

```bash
go test -v ./...
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
