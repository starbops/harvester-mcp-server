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

// ListNamespaces retrieves a list of namespaces from the Harvester cluster.
func ListNamespaces(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespaces, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list namespaces: %v", err)), nil
	}

	// Create a summary of namespaces instead of returning raw JSON
	summary := formatNamespaceListSummary(namespaces)
	return mcp.NewToolResultText(summary), nil
}

// formatNamespaceListSummary creates a human-readable summary of namespaces
func formatNamespaceListSummary(namespaces *corev1.NamespaceList) string {
	if len(namespaces.Items) == 0 {
		return "No namespaces found in the cluster."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d namespace(s):\n\n", len(namespaces.Items)))

	// Sort namespaces by status (Active first, then Terminating)
	activeNamespaces := make([]corev1.Namespace, 0)
	terminatingNamespaces := make([]corev1.Namespace, 0)
	otherNamespaces := make([]corev1.Namespace, 0)

	for _, ns := range namespaces.Items {
		if ns.Status.Phase == corev1.NamespaceActive {
			activeNamespaces = append(activeNamespaces, ns)
		} else if ns.Status.Phase == corev1.NamespaceTerminating {
			terminatingNamespaces = append(terminatingNamespaces, ns)
		} else {
			otherNamespaces = append(otherNamespaces, ns)
		}
	}

	// Process active namespaces
	if len(activeNamespaces) > 0 {
		sb.WriteString("Active Namespaces:\n")
		for _, ns := range activeNamespaces {
			formatNamespaceEntry(&sb, ns)
		}
		sb.WriteString("\n")
	}

	// Process terminating namespaces
	if len(terminatingNamespaces) > 0 {
		sb.WriteString("Terminating Namespaces:\n")
		for _, ns := range terminatingNamespaces {
			formatNamespaceEntry(&sb, ns)
		}
		sb.WriteString("\n")
	}

	// Process other namespaces
	if len(otherNamespaces) > 0 {
		sb.WriteString("Other Namespaces:\n")
		for _, ns := range otherNamespaces {
			formatNamespaceEntry(&sb, ns)
		}
	}

	return sb.String()
}

// formatNamespaceEntry formats a single namespace entry in the list
func formatNamespaceEntry(sb *strings.Builder, ns corev1.Namespace) {
	// Basic namespace info
	sb.WriteString(fmt.Sprintf("  • %s\n", ns.Name))
	sb.WriteString(fmt.Sprintf("    Status: %s\n", ns.Status.Phase))

	// Creation time
	creationTime := ns.CreationTimestamp.Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

	// Labels (select important ones only to avoid clutter)
	if len(ns.Labels) > 0 {
		sb.WriteString("    Labels: ")
		labelPairs := []string{}
		for key, value := range ns.Labels {
			labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", key, value))
		}
		sb.WriteString(strings.Join(labelPairs, ", "))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
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

	// Format the namespace into a more readable format
	summary := formatNamespaceDetail(namespace)
	return mcp.NewToolResultText(summary), nil
}

// formatNamespaceDetail creates a human-readable summary of a single namespace
func formatNamespaceDetail(ns *corev1.Namespace) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", ns.Name))
	sb.WriteString(fmt.Sprintf("Status: %s\n", ns.Status.Phase))

	// Creation time
	creationTime := ns.CreationTimestamp.Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// Resource quotas and limits would require additional API calls

	// Labels
	if len(ns.Labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range ns.Labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Annotations
	if len(ns.Annotations) > 0 {
		sb.WriteString("\nAnnotations:\n")
		for key, value := range ns.Annotations {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Finalizers
	if len(ns.Spec.Finalizers) > 0 {
		sb.WriteString("\nFinalizers:\n")
		for _, finalizer := range ns.Spec.Finalizers {
			sb.WriteString(fmt.Sprintf("  - %s\n", finalizer))
		}
	}

	return sb.String()
}
