package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Virtual Machine Resource GVR (Group Version Resource)
var vmGVR = schema.GroupVersionResource{
	Group:    "kubevirt.io",
	Version:  "v1",
	Resource: "virtualmachines",
}

// ListVirtualMachines retrieves a list of VMs from the Harvester cluster.
func ListVirtualMachines(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dynamic client: %v", err)), nil
	}

	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List VMs in all namespaces
		vms, err := dynamicClient.Resource(vmGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list virtual machines: %v", err)), nil
		}

		vmsJSON, err := json.MarshalIndent(vms, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert virtual machines to JSON: %v", err)), nil
		}

		return mcp.NewToolResultText(string(vmsJSON)), nil
	}

	// List VMs in specific namespace
	vms, err := dynamicClient.Resource(vmGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list virtual machines in namespace %s: %v", namespace, err)), nil
	}

	vmsJSON, err := json.MarshalIndent(vms, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert virtual machines to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(vmsJSON)), nil
}

// GetVirtualMachine retrieves details for a specific VM from the Harvester cluster.
func GetVirtualMachine(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dynamic client: %v", err)), nil
	}

	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		return mcp.NewToolResultError("Namespace is required"), nil
	}

	name, ok := req.Params.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("Virtual Machine name is required"), nil
	}

	vm, err := dynamicClient.Resource(vmGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get virtual machine %s in namespace %s: %v", name, namespace, err)), nil
	}

	// Format the VM for better readability
	formattedVM := formatVirtualMachine(vm)

	vmJSON, err := json.MarshalIndent(formattedVM, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert virtual machine to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(vmJSON)), nil
}

// formatVirtualMachine formats the VM unstructured object to a more readable format.
func formatVirtualMachine(vm *unstructured.Unstructured) map[string]interface{} {
	formattedVM := map[string]interface{}{
		"name":      vm.GetName(),
		"namespace": vm.GetNamespace(),
		"status":    getNestedMap(vm.Object, "status"),
		"spec":      getNestedMap(vm.Object, "spec"),
	}
	return formattedVM
}

// getNestedMap safely retrieves a nested map from an unstructured object.
func getNestedMap(obj map[string]interface{}, key string) map[string]interface{} {
	value, found, _ := unstructured.NestedMap(obj, key)
	if !found {
		return map[string]interface{}{}
	}
	return value
}
