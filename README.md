# Harvester MCP Server

Model Context Protocol (MCP) server for Harvester HCI that enables Claude, Cursor, and other AI assistants to interact with Harvester clusters through the MCP protocol.

## Overview

Harvester MCP Server is a Go implementation of the [Model Context Protocol (MCP)](https://github.com/model-context-protocol/mcp) specifically designed for [Harvester HCI](https://github.com/harvester/harvester). It allows AI assistants like Claude and Cursor to perform CRUD operations on Harvester clusters, which are essentially Kubernetes clusters with Harvester-specific CRDs.

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
go build -o harvester-mcp-server cmd/harvester-mcp-server/main.go

# Run
./harvester-mcp-server
```

### Using Go Install

```bash
go install github.com/starbops/harvester-mcp-server/cmd/harvester-mcp-server@latest
```

## Configuration

The server automatically uses the kubeconfig file located at `~/.kube/config` or uses in-cluster configuration if deployed inside a Kubernetes cluster.

## Usage with Claude Desktop

1. Install Claude Desktop
2. Open Claude Desktop configuration file (`~/.config/claude-desktop/claude_desktop_config.json` or similar)
3. Add the Harvester MCP server to the `mcpServers` section:

```json
{
  "mcpServers": {
    "harvester": {
      "command": "/path/to/harvester-mcp-server"
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
