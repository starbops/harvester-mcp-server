package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListServices retrieves a list of services from the Harvester cluster.
func ListServices(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List services in all namespaces
		services, err := client.Clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list services: %v", err)), nil
		}

		servicesJSON, err := json.MarshalIndent(services, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert services to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(servicesJSON)), nil
	}

	// List services in specific namespace
	services, err := client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list services in namespace %s: %v", namespace, err)), nil
	}

	servicesJSON, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert services to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(servicesJSON)), nil
}

// GetService retrieves details for a specific service from the Harvester cluster.
func GetService(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		return mcp.NewToolResultError("Namespace is required"), nil
	}

	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Service name is required"), nil
	}

	service, err := client.Clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get service %s in namespace %s: %v", name, namespace, err)), nil
	}

	serviceJSON, err := json.MarshalIndent(service, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert service to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(serviceJSON)), nil
}
