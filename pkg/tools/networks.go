package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Network Resource GVR (Group Version Resource)
var networkGVR = schema.GroupVersionResource{
	Group:    "network.harvesterhci.io",
	Version:  "v1beta1",
	Resource: "clusternetworks",
}

// ListNetworks retrieves a list of Networks from the Harvester cluster.
func ListNetworks(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dynamic client: %v", err)), nil
	}

	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List networks in all namespaces
		networks, err := dynamicClient.Resource(networkGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list networks: %v", err)), nil
		}

		networksJSON, err := json.MarshalIndent(networks, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert networks to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(networksJSON)), nil
	}

	// List networks in specific namespace
	networks, err := dynamicClient.Resource(networkGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list networks in namespace %s: %v", namespace, err)), nil
	}

	networksJSON, err := json.MarshalIndent(networks, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert networks to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(networksJSON)), nil
}
