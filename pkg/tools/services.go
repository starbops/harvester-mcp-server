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

// ListServices retrieves a list of services from the Harvester cluster.
func ListServices(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace, ok := req.Params.Arguments["namespace"].(string)
	if !ok || namespace == "" {
		// List services in all namespaces
		services, err := client.Clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list services: %v", err)), nil
		}

		// Create a summary of services instead of returning raw JSON
		summary := formatServiceListSummary(services)
		return mcp.NewToolResultText(summary), nil
	}

	// List services in specific namespace
	services, err := client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list services in namespace %s: %v", namespace, err)), nil
	}

	// Create a summary of services instead of returning raw JSON
	summary := formatServiceListSummary(services)
	return mcp.NewToolResultText(summary), nil
}

// formatServiceListSummary creates a human-readable summary of services
func formatServiceListSummary(services *corev1.ServiceList) string {
	if len(services.Items) == 0 {
		return "No services found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d service(s):\n\n", len(services.Items)))

	// Group services by namespace
	servicesByNamespace := make(map[string][]corev1.Service)
	for _, svc := range services.Items {
		servicesByNamespace[svc.Namespace] = append(servicesByNamespace[svc.Namespace], svc)
	}

	// Print services grouped by namespace
	for namespace, nsSvcs := range servicesByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d services)\n", namespace, len(nsSvcs)))

		for _, svc := range nsSvcs {
			// Basic service info
			sb.WriteString(fmt.Sprintf("  • %s\n", svc.Name))

			// Type
			sb.WriteString(fmt.Sprintf("    Type: %s\n", svc.Spec.Type))

			// Cluster IP
			if svc.Spec.ClusterIP != "" {
				sb.WriteString(fmt.Sprintf("    Cluster IP: %s\n", svc.Spec.ClusterIP))
			}

			// External IPs
			if len(svc.Spec.ExternalIPs) > 0 {
				sb.WriteString(fmt.Sprintf("    External IPs: %s\n", strings.Join(svc.Spec.ExternalIPs, ", ")))
			}

			// LoadBalancer IP
			if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && svc.Status.LoadBalancer.Ingress != nil {
				var ips []string
				for _, ing := range svc.Status.LoadBalancer.Ingress {
					if ing.IP != "" {
						ips = append(ips, ing.IP)
					}
					if ing.Hostname != "" {
						ips = append(ips, ing.Hostname)
					}
				}
				if len(ips) > 0 {
					sb.WriteString(fmt.Sprintf("    LoadBalancer IP(s): %s\n", strings.Join(ips, ", ")))
				}
			}

			// Ports
			if len(svc.Spec.Ports) > 0 {
				sb.WriteString("    Ports:\n")
				for _, port := range svc.Spec.Ports {
					protocol := string(port.Protocol)
					if port.Name != "" {
						sb.WriteString(fmt.Sprintf("      - %s: %d", port.Name, port.Port))
					} else {
						sb.WriteString(fmt.Sprintf("      - %d", port.Port))
					}

					if port.NodePort != 0 {
						sb.WriteString(fmt.Sprintf(" (NodePort: %d)", port.NodePort))
					}

					if port.TargetPort.String() != "0" {
						sb.WriteString(fmt.Sprintf(" → %s", port.TargetPort.String()))
					}

					sb.WriteString(fmt.Sprintf(" %s\n", protocol))
				}
			}

			// Selectors
			if len(svc.Spec.Selector) > 0 {
				sb.WriteString("    Selector:\n")
				for key, value := range svc.Spec.Selector {
					sb.WriteString(fmt.Sprintf("      %s: %s\n", key, value))
				}
			}

			// Creation time
			creationTime := svc.CreationTimestamp.Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
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

	// Format the service into a more readable format
	summary := formatServiceDetail(service)
	return mcp.NewToolResultText(summary), nil
}

// formatServiceDetail creates a human-readable summary of a single service
func formatServiceDetail(svc *corev1.Service) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Service: %s\n", svc.Name))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", svc.Namespace))
	sb.WriteString(fmt.Sprintf("Type: %s\n", svc.Spec.Type))

	// IP Addresses
	if svc.Spec.ClusterIP != "" {
		sb.WriteString(fmt.Sprintf("Cluster IP: %s\n", svc.Spec.ClusterIP))
	}

	if len(svc.Spec.ExternalIPs) > 0 {
		sb.WriteString(fmt.Sprintf("External IPs: %s\n", strings.Join(svc.Spec.ExternalIPs, ", ")))
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && svc.Status.LoadBalancer.Ingress != nil {
		var ips []string
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				ips = append(ips, ing.IP)
			}
			if ing.Hostname != "" {
				ips = append(ips, ing.Hostname)
			}
		}
		if len(ips) > 0 {
			sb.WriteString(fmt.Sprintf("LoadBalancer IP(s): %s\n", strings.Join(ips, ", ")))
		}
	}

	// Session Affinity
	sb.WriteString(fmt.Sprintf("Session Affinity: %s\n", svc.Spec.SessionAffinity))

	// Creation time
	creationTime := svc.CreationTimestamp.Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// Labels
	if len(svc.Labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range svc.Labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Annotations
	if len(svc.Annotations) > 0 {
		sb.WriteString("\nAnnotations:\n")
		for key, value := range svc.Annotations {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Ports
	if len(svc.Spec.Ports) > 0 {
		sb.WriteString("\nPorts:\n")
		for _, port := range svc.Spec.Ports {
			protocol := string(port.Protocol)
			if port.Name != "" {
				sb.WriteString(fmt.Sprintf("  - %s: %d", port.Name, port.Port))
			} else {
				sb.WriteString(fmt.Sprintf("  - %d", port.Port))
			}

			if port.NodePort != 0 {
				sb.WriteString(fmt.Sprintf(" (NodePort: %d)", port.NodePort))
			}

			if port.TargetPort.String() != "0" {
				sb.WriteString(fmt.Sprintf(" → %s", port.TargetPort.String()))
			}

			sb.WriteString(fmt.Sprintf(" %s\n", protocol))
		}
	}

	// Selector
	if len(svc.Spec.Selector) > 0 {
		sb.WriteString("\nSelector:\n")
		for key, value := range svc.Spec.Selector {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Endpoints would require another API call

	return sb.String()
}
