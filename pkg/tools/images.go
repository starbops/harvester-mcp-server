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

// Image Resource GVR (Group Version Resource)
var imageGVR = schema.GroupVersionResource{
	Group:    "harvesterhci.io",
	Version:  "v1beta1",
	Resource: "virtualmachineimages",
}

// ListImages retrieves a list of Images from the Harvester cluster.
func ListImages(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dynamic client: %v", err)), nil
	}

	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List images in all namespaces
		images, err := dynamicClient.Resource(imageGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list images: %v", err)), nil
		}

		imagesJSON, err := json.MarshalIndent(images, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert images to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(imagesJSON)), nil
	}

	// List images in specific namespace
	images, err := dynamicClient.Resource(imageGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list images in namespace %s: %v", namespace, err)), nil
	}

	imagesJSON, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert images to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(imagesJSON)), nil
}
