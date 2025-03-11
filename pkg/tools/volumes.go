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

// Volume Resource GVR (Group Version Resource)
var volumeGVR = schema.GroupVersionResource{
	Group:    "harvesterhci.io",
	Version:  "v1beta1",
	Resource: "volumes",
}

// ListVolumes retrieves a list of Volumes from the Harvester cluster.
func ListVolumes(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dynamic client: %v", err)), nil
	}

	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List volumes in all namespaces
		volumes, err := dynamicClient.Resource(volumeGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list volumes: %v", err)), nil
		}

		volumesJSON, err := json.MarshalIndent(volumes, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert volumes to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(volumesJSON)), nil
	}

	// List volumes in specific namespace
	volumes, err := dynamicClient.Resource(volumeGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list volumes in namespace %s: %v", namespace, err)), nil
	}

	volumesJSON, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert volumes to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(volumesJSON)), nil
}
