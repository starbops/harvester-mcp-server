package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployments retrieves a list of deployments from the Harvester cluster.
func ListDeployments(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List deployments in all namespaces
		deployments, err := client.Clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list deployments: %v", err)), nil
		}

		deploymentsJSON, err := json.MarshalIndent(deployments, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert deployments to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(deploymentsJSON)), nil
	}

	// List deployments in specific namespace
	deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list deployments in namespace %s: %v", namespace, err)), nil
	}

	deploymentsJSON, err := json.MarshalIndent(deployments, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert deployments to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(deploymentsJSON)), nil
}

// GetDeployment retrieves details for a specific deployment from the Harvester cluster.
func GetDeployment(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		return mcp.NewToolResultError("Namespace is required"), nil
	}

	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Deployment name is required"), nil
	}

	deployment, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get deployment %s in namespace %s: %v", name, namespace, err)), nil
	}

	deploymentJSON, err := json.MarshalIndent(deployment, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert deployment to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(deploymentJSON)), nil
}
