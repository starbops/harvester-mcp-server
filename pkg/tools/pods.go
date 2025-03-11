package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPods retrieves a list of pods from the Harvester cluster.
func ListPods(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List pods in all namespaces
		pods, err := client.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list pods: %v", err)), nil
		}

		podsJSON, err := json.MarshalIndent(pods, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert pods to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(podsJSON)), nil
	}

	// List pods in specific namespace
	pods, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list pods in namespace %s: %v", namespace, err)), nil
	}

	podsJSON, err := json.MarshalIndent(pods, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert pods to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(podsJSON)), nil
}

// GetPod retrieves details for a specific pod from the Harvester cluster.
func GetPod(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		return mcp.NewToolResultError("Namespace is required"), nil
	}

	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Pod name is required"), nil
	}

	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get pod %s in namespace %s: %v", name, namespace, err)), nil
	}

	podJSON, err := json.MarshalIndent(pod, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert pod to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(podJSON)), nil
}

// DeletePod deletes a specific pod from the Harvester cluster.
func DeletePod(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		return mcp.NewToolResultError("Namespace is required"), nil
	}

	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Pod name is required"), nil
	}

	err := client.Clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete pod %s in namespace %s: %v", name, namespace, err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted pod %s in namespace %s", name, namespace)), nil
}
