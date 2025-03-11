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

// ListPods retrieves a list of pods from the Harvester cluster.
func ListPods(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List pods in all namespaces
		pods, err := client.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list pods: %v", err)), nil
		}

		// Create a summary of pods instead of returning raw JSON
		summary := formatPodListSummary(pods)
		return mcp.NewToolResultText(summary), nil
	}

	// List pods in specific namespace
	pods, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list pods in namespace %s: %v", namespace, err)), nil
	}

	// Create a summary of pods instead of returning raw JSON
	summary := formatPodListSummary(pods)
	return mcp.NewToolResultText(summary), nil
}

// formatPodListSummary creates a human-readable summary of pods
func formatPodListSummary(pods *corev1.PodList) string {
	if len(pods.Items) == 0 {
		return "No pods found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d pod(s):\n\n", len(pods.Items)))

	// Group pods by namespace
	podsByNamespace := make(map[string][]corev1.Pod)
	for _, pod := range pods.Items {
		podsByNamespace[pod.Namespace] = append(podsByNamespace[pod.Namespace], pod)
	}

	// Print pods grouped by namespace
	for namespace, nsPods := range podsByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d pods)\n", namespace, len(nsPods)))

		for _, pod := range nsPods {
			// Get pod status
			status := string(pod.Status.Phase)
			if pod.Status.Reason != "" {
				status = pod.Status.Reason
			}

			// Add ready container count
			ready := 0
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Ready {
					ready++
				}
			}

			// Basic pod info
			sb.WriteString(fmt.Sprintf("  • %s\n", pod.Name))
			sb.WriteString(fmt.Sprintf("    Status: %s\n", status))
			sb.WriteString(fmt.Sprintf("    Ready: %d/%d containers\n", ready, len(pod.Spec.Containers)))

			// Add node name if scheduled
			if pod.Spec.NodeName != "" {
				sb.WriteString(fmt.Sprintf("    Node: %s\n", pod.Spec.NodeName))
			}

			// Add IP if assigned
			if pod.Status.PodIP != "" {
				sb.WriteString(fmt.Sprintf("    IP: %s\n", pod.Status.PodIP))
			}

			// Add creation time
			creationTime := pod.CreationTimestamp.Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			// Add restart count
			totalRestarts := 0
			for _, containerStatus := range pod.Status.ContainerStatuses {
				totalRestarts += int(containerStatus.RestartCount)
			}
			sb.WriteString(fmt.Sprintf("    Restarts: %d\n", totalRestarts))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
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

	// Format the pod into a more readable format
	summary := formatPodDetail(pod)
	return mcp.NewToolResultText(summary), nil
}

// formatPodDetail creates a human-readable summary of a single pod
func formatPodDetail(pod *corev1.Pod) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pod: %s\n", pod.Name))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))

	// Status
	status := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		status = pod.Status.Reason
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Node
	if pod.Spec.NodeName != "" {
		sb.WriteString(fmt.Sprintf("Node: %s\n", pod.Spec.NodeName))
	}

	// IP addresses
	if pod.Status.PodIP != "" {
		sb.WriteString(fmt.Sprintf("Pod IP: %s\n", pod.Status.PodIP))
	}

	// Creation time
	creationTime := pod.CreationTimestamp.Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// QoS Class
	sb.WriteString(fmt.Sprintf("QoS Class: %s\n", pod.Status.QOSClass))

	// Labels
	if len(pod.Labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range pod.Labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Containers
	sb.WriteString("\nContainers:\n")
	for i, container := range pod.Spec.Containers {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, container.Name))
		sb.WriteString(fmt.Sprintf("     Image: %s\n", container.Image))

		// Container resources
		if container.Resources.Limits != nil || container.Resources.Requests != nil {
			sb.WriteString("     Resources:\n")
			if container.Resources.Limits != nil {
				if cpu := container.Resources.Limits.Cpu(); cpu != nil {
					sb.WriteString(fmt.Sprintf("       Limits CPU: %s\n", cpu.String()))
				}
				if memory := container.Resources.Limits.Memory(); memory != nil {
					sb.WriteString(fmt.Sprintf("       Limits Memory: %s\n", memory.String()))
				}
			}
			if container.Resources.Requests != nil {
				if cpu := container.Resources.Requests.Cpu(); cpu != nil {
					sb.WriteString(fmt.Sprintf("       Requests CPU: %s\n", cpu.String()))
				}
				if memory := container.Resources.Requests.Memory(); memory != nil {
					sb.WriteString(fmt.Sprintf("       Requests Memory: %s\n", memory.String()))
				}
			}
		}

		// Container status
		var containerStatus *corev1.ContainerStatus
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == container.Name {
				containerStatus = &cs
				break
			}
		}

		if containerStatus != nil {
			sb.WriteString(fmt.Sprintf("     Ready: %t\n", containerStatus.Ready))
			sb.WriteString(fmt.Sprintf("     Restarts: %d\n", containerStatus.RestartCount))

			// Container state
			if containerStatus.State.Running != nil {
				startTime := containerStatus.State.Running.StartedAt.Format(time.RFC3339)
				sb.WriteString(fmt.Sprintf("     State: Running (since %s)\n", startTime))
			} else if containerStatus.State.Waiting != nil {
				reason := containerStatus.State.Waiting.Reason
				message := containerStatus.State.Waiting.Message
				sb.WriteString(fmt.Sprintf("     State: Waiting (%s)\n", reason))
				if message != "" {
					sb.WriteString(fmt.Sprintf("     Message: %s\n", message))
				}
			} else if containerStatus.State.Terminated != nil {
				reason := containerStatus.State.Terminated.Reason
				exitCode := containerStatus.State.Terminated.ExitCode
				sb.WriteString(fmt.Sprintf("     State: Terminated (%s, exit code: %d)\n", reason, exitCode))
			}
		}

		sb.WriteString("\n")
	}

	// Events could be included here but would require a separate API call

	return sb.String()
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
