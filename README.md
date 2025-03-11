# Harvester MCP Server

Model Context Protocol (MCP) server for Harvester HCI that enables Claude Desktop, Cursor, and other AI assistants to interact with Harvester clusters through the MCP protocol.

## Overview

Harvester MCP Server is a Go implementation of the [Model Context Protocol (MCP)](https://spec.modelcontextprotocol.io/specification/2024-11-05/) specifically designed for [Harvester HCI](https://github.com/harvester/harvester). It allows AI assistants like Claude Desktop and Cursor to perform CRUD operations on Harvester clusters, which are essentially Kubernetes clusters with Harvester-specific CRDs.

## Features

- **Kubernetes Core Resources**:
  - Pods: List, Get, Delete
  - Deployments: List, Get
  - Services: List, Get
  - Namespaces: List, Get
  - Nodes: List, Get
  - Custom Resource Definitions (CRDs): List

- **Harvester-Specific Resources**:
  - Virtual Machines: List, Get
  - Images: List
  - Volumes: List
  - Networks: List

## Requirements

- Go 1.21+
- Access to a Harvester cluster with a valid kubeconfig

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/starbops/harvester-mcp-server.git
cd harvester-mcp-server

# Build
make build

# Run
./bin/harvester-mcp-server
```

### Using Go Install

```bash
go install github.com/starbops/harvester-mcp-server/cmd/harvester-mcp-server@latest
```

## Configuration

The server automatically looks for Kubernetes configuration in the following order:
1. In-cluster configuration (if running inside a Kubernetes cluster)
2. Path specified by the `--kubeconfig` flag
3. Path specified by the `KUBECONFIG` environment variable
4. Default location at `~/.kube/config`

### Command-Line Flags

```
Usage:
  harvester-mcp-server [flags]

Flags:
  -h, --help                help for harvester-mcp-server
      --kubeconfig string   Path to the kubeconfig file (default is $KUBECONFIG or $HOME/.kube/config)
```

### Examples

Using a specific kubeconfig file:
```bash
harvester-mcp-server --kubeconfig=/path/to/kubeconfig.yaml
```

Using the KUBECONFIG environment variable:
```bash
export KUBECONFIG=$HOME/config.yaml
harvester-mcp-server
```

## Usage with Claude Desktop

1. Install Claude Desktop
2. Open Claude Desktop configuration file (`~/.config/claude-desktop/claude_desktop_config.json` or similar)
3. Add the Harvester MCP server to the `mcpServers` section:

```json
{
  "mcpServers": {
    "harvester": {
      "command": "/path/to/harvester-mcp-server",
      "args": ["--kubeconfig=/path/to/kubeconfig.yaml"]
    }
  }
}
```

4. Restart Claude Desktop
5. The Harvester MCP tools should now be available to Claude

## Development

### Project Structure

- `cmd/harvester-mcp-server`: Main application entry point
- `pkg/client`: Kubernetes client implementation
- `pkg/cmd`: CLI commands implementation using Cobra
- `pkg/mcp`: MCP server implementation
- `pkg/tools`: Tool implementations for interacting with Harvester resources

### Adding New Tools

To add a new tool:

1. Create a new function in the appropriate file under `pkg/tools`
2. Register the tool in `pkg/mcp/server.go` in the `registerTools` method

## License

MIT License

## Acknowledgments

- [Harvester HCI](https://github.com/harvester/harvester) - The foundation for this project
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - The Go SDK for Model Context Protocol
- [manusa/kubernetes-mcp-server](https://github.com/manusa/kubernetes-mcp-server) - Reference implementation for Kubernetes MCP server
- [spf13/cobra](https://github.com/spf13/cobra) - CLI command framework
