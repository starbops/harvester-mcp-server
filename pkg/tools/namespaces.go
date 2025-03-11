package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNamespaces retrieves a list of namespaces from the Harvester cluster.
func ListNamespaces(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespaces, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list namespaces: %v", err)), nil
	}

	namespacesJSON, err := json.MarshalIndent(namespaces, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert namespaces to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(namespacesJSON)), nil
}

// GetNamespace retrieves details for a specific namespace from the Harvester cluster.
func GetNamespace(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Namespace name is required"), nil
	}

	namespace, err := client.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get namespace %s: %v", name, err)), nil
	}

	namespaceJSON, err := json.MarshalIndent(namespace, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert namespace to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(namespaceJSON)), nil
}
