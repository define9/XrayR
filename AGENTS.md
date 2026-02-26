# AGENTS.md

This file provides guidelines for AI agents working on the XrayR codebase.

## Project Overview

XrayR is a Xray backend framework written in Go (v1.25) that supports V2Ray, Trojan, and Shadowsocks protocols. It integrates with multiple panel systems.

## Build, Lint, and Test Commands

### Dependencies
```bash
go mod download
```

### Build
```bash
# Standard build
go build -v -o XrayR

# Release build (used in CI)
go build -v -o build_assets/XrayR -trimpath -ldflags "-s -w -buildid="
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests in specific package
go test ./api/sspanel/...

# Run single test file
go test -v ./api/sspanel/sspanel_test.go ./api/sspanel/sspanel.go

# Run single test function
go test -v -run TestGetV2rayNodeInfo ./api/sspanel/...

# Run tests with coverage
go test -cover ./...
```

### Format Code
```bash
# Format all Go files
gofmt -w .

# Check formatting (dry-run)
gofmt -d .
```

### Vet
```bash
go vet ./...
```

### Module Tidy
```bash
go mod tidy
```

## Code Style Guidelines

### Imports

Group imports in this order:
1. Standard library
2. Third-party packages
3. Internal/XrayR packages (github.com/XrayR-project/XrayR/...)

```go
import (
    "encoding/json"
    "fmt"
    "regexp"

    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/xtls/xray-core/core"

    "github.com/XrayR-project/XrayR/api"
    "github.com/XrayR-project/XrayR/panel"
)
```

### Naming Conventions

- **Packages**: lowercase, short, descriptive
- **Variables/Functions**: camelCase
- **Exported Types/Functions**: PascalCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE for error messages
- **Interfaces**: PascalCase, often ending with "er" (e.g., `API`, `Builder`)
- **Errors**: PascalCase constants for sentinel errors, fmt.Errorf for others

```go
// Constants
const (
    UserNotModified = "users not modified"
    NodeNotModified = "node not modified"
)

// Interfaces
type API interface {
    GetNodeInfo() (nodeInfo *NodeInfo, err error)
    GetUserList() (userList *[]UserInfo, err error)
}

// Variables
var rootCmd = &cobra.Command{...}

// Functions
func GetSystemInfo() (Cpu float64, Mem float64, Disk float64, Uptime uint64, err error) {
```

### Error Handling

- Return errors explicitly from functions
- Use `fmt.Errorf("message: %w", err)` for wrapped errors
- Use `log` or `logger` for logging errors in void functions
- Sentinel errors as package-level constants (PascalCase)

```go
// Return error
func (c *Controller) Start() error {
    newNodeInfo, err := c.apiClient.GetNodeInfo()
    if err != nil {
        return err
    }
    // ...
}

// Wrap error
if err != nil {
    return fmt.Errorf("Parse config file %v failed: %s", cfgFile, err)
}

// Log error in void function
if err := c.AddInboundLimiter(...); err != nil {
    c.logger.Print(err)
}
```

### Struct Tags

Use `mapstructure` tags for configuration structs:

```go
type Config struct {
    APIHost             string  `mapstructure:"ApiHost"`
    NodeID              int     `mapstructure:"NodeID"`
    Key                 string  `mapstructure:"ApiKey"`
    NodeType            string  `mapstructure:"NodeType"`
}
```

### Comments

- Document all exported types and functions
- Use single-line comments for package-level declarations
- Comments should be complete sentences

```go
// Config API config
type Config struct {...}

// GetSystemInfo get the system info of a given periodic
func GetSystemInfo() (Cpu float64, ...) {...}
```

### Logging

Use logrus with structured logging:

```go
log "github.com/sirupsen/logrus"

// In struct initialization
logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
    "Host": api.Describe().APIHost,
    "Type": api.Describe().NodeType,
})

// Usage
c.logger.Printf("Added %d new users", len(*userInfo))
c.logger.Print(err)
```

### Testing

- Test files use `_test.go` suffix
- Test functions use `Test` prefix
- Use standard `testing` package with testify for assertions
- Table-driven tests are acceptable

```go
func TestGetV2rayNodeInfo(t *testing.T) {
    client := CreateClient()
    nodeInfo, err := client.GetNodeInfo()
    if err != nil {
        t.Error(err)
    }
    t.Log(nodeInfo)
}
```

### Logging in Tests

Avoid `t.Log` for debug output; use it sparingly. Tests should either pass or fail clearly.

### Configuration

- Use Viper for configuration management
- Support YAML config files
- Watch config for hot-reload

### Project Structure

```
├── api/              # Panel API implementations (sspanel, v2board, etc.)
├── app/              # Core application features (dispatcher, etc.)
├── cmd/              # CLI commands (root, version, etc.)
├── common/           # Shared utilities (lego, limiter, serverstatus)
├── panel/            # Panel configuration and management
├── service/          # Core service logic
├── .github/workflows/# CI/CD pipelines
├── release/          # Release assets (config examples, geo files)
├── main.go           # Entry point
└── config.yml        # Default configuration
```

### Key Patterns

1. **API Interface Pattern**: Each panel implementation implements the `API` interface
2. **Controller Pattern**: Controller manages Xray instance, API client, and node state
3. **Builder Pattern**: Inbound/Outbound/User builders for configuration construction
4. **Periodic Tasks**: Use `xray-core/common/task` for periodic operations

### CI/CD

- GitHub Actions for multi-platform builds (Windows, Linux, macOS, FreeBSD, etc.)
- Builds are triggered on PRs and pushes to master
- Release artifacts include geoip/geosite databases and config templates

### Things to Avoid

- Don't suppress errors with `_`
- Don't use `panic` in production code except for init/recoverable errors
- Don't commit generated files (use .gitignore)
- Don't modify generated code manually (regenerate instead)
