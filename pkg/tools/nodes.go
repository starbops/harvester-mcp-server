package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNodes retrieves a list of nodes from the Harvester cluster.
func ListNodes(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list nodes: %v", err)), nil
	}

	nodesJSON, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert nodes to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(nodesJSON)), nil
}

// GetNode retrieves details for a specific node from the Harvester cluster.
func GetNode(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Node name is required"), nil
	}

	node, err := client.Clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get node %s: %v", name, err)), nil
	}

	nodeJSON, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert node to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(nodeJSON)), nil
}
