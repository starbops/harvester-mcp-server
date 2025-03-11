package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNodes retrieves a list of nodes from the Harvester cluster.
func ListNodes(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list nodes: %v", err)), nil
	}

	// Create a summary of nodes instead of returning raw JSON
	summary := formatNodeListSummary(nodes)
	return mcp.NewToolResultText(summary), nil
}

// formatNodeListSummary creates a human-readable summary of nodes
func formatNodeListSummary(nodes *corev1.NodeList) string {
	if len(nodes.Items) == 0 {
		return "No nodes found in the Harvester cluster."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d node(s) in the Harvester cluster:\n\n", len(nodes.Items)))

	for i, node := range nodes.Items {
		sb.WriteString(fmt.Sprintf("%d. Name: %s\n", i+1, node.Name))

		// Add node IP addresses
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				sb.WriteString(fmt.Sprintf("   IP: %s\n", addr.Address))
			}
		}

		// Add status
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				isReady = true
			}
		}

		readyStatus := "Ready"
		if !isReady {
			readyStatus = "Not Ready"
		}
		sb.WriteString(fmt.Sprintf("   Status: %s\n", readyStatus))

		// Add roles
		var roles []string
		for label := range node.Labels {
			if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
				roles = append(roles, role)
			}
		}

		if len(roles) > 0 {
			sb.WriteString(fmt.Sprintf("   Roles: %s\n", strings.Join(roles, ", ")))
		} else {
			sb.WriteString("   Roles: <none>\n")
		}

		// Add cpu/memory resources
		cpu := node.Status.Capacity.Cpu().String()
		memory := node.Status.Capacity.Memory().String()
		sb.WriteString(fmt.Sprintf("   CPU: %s\n", cpu))
		sb.WriteString(fmt.Sprintf("   Memory: %s\n", memory))

		// Add creation timestamp
		creationTime := node.CreationTimestamp.Format(time.RFC3339)
		sb.WriteString(fmt.Sprintf("   Created: %s\n", creationTime))

		if i < len(nodes.Items)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
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

	// Format the node into a more readable format
	summary := formatNodeDetail(node)
	return mcp.NewToolResultText(summary), nil
}

// formatNodeDetail creates a human-readable summary of a single node
func formatNodeDetail(node *corev1.Node) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Node: %s\n\n", node.Name))

	// Add node IP addresses
	sb.WriteString("Addresses:\n")
	for _, addr := range node.Status.Addresses {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", addr.Type, addr.Address))
	}

	// Add status conditions
	sb.WriteString("\nConditions:\n")
	for _, condition := range node.Status.Conditions {
		status := "True"
		if condition.Status != corev1.ConditionTrue {
			status = "False"
		}
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", condition.Type, status))
	}

	// Add roles
	var roles []string
	for label := range node.Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			roles = append(roles, role)
		}
	}

	sb.WriteString("\nRoles: ")
	if len(roles) > 0 {
		sb.WriteString(strings.Join(roles, ", "))
	} else {
		sb.WriteString("<none>")
	}
	sb.WriteString("\n")

	// Add resources
	sb.WriteString("\nCapacity:\n")
	sb.WriteString(fmt.Sprintf("  CPU: %s\n", node.Status.Capacity.Cpu().String()))
	sb.WriteString(fmt.Sprintf("  Memory: %s\n", node.Status.Capacity.Memory().String()))
	sb.WriteString(fmt.Sprintf("  Pods: %s\n", node.Status.Capacity.Pods().String()))

	// Add system info
	sb.WriteString("\nSystem Info:\n")
	sb.WriteString(fmt.Sprintf("  Architecture: %s\n", node.Status.NodeInfo.Architecture))
	sb.WriteString(fmt.Sprintf("  OS: %s\n", node.Status.NodeInfo.OperatingSystem))
	sb.WriteString(fmt.Sprintf("  Kernel: %s\n", node.Status.NodeInfo.KernelVersion))
	sb.WriteString(fmt.Sprintf("  Container Runtime: %s\n", node.Status.NodeInfo.ContainerRuntimeVersion))
	sb.WriteString(fmt.Sprintf("  Kubelet: %s\n", node.Status.NodeInfo.KubeletVersion))

	return sb.String()
}
