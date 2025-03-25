package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PodFormatter handles formatting for Pod resources
type PodFormatter struct{}

func (f *PodFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pod: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Status
	status := getNestedString(res.Object, "status", "phase")
	reason := getNestedString(res.Object, "status", "reason")
	if reason != "" {
		status = reason
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Node
	nodeName := getNestedString(res.Object, "spec", "nodeName")
	if nodeName != "" {
		sb.WriteString(fmt.Sprintf("Node: %s\n", nodeName))
	}

	// IP addresses
	podIP := getNestedString(res.Object, "status", "podIP")
	if podIP != "" {
		sb.WriteString(fmt.Sprintf("Pod IP: %s\n", podIP))
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// QoS Class
	qosClass := getNestedString(res.Object, "status", "qosClass")
	sb.WriteString(fmt.Sprintf("QoS Class: %s\n", qosClass))

	// Labels
	if labels := res.GetLabels(); len(labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Containers
	containers, _, _ := unstructured.NestedSlice(res.Object, "spec", "containers")
	if len(containers) > 0 {
		sb.WriteString("\nContainers:\n")
		for i, containerObj := range containers {
			container, ok := containerObj.(map[string]interface{})
			if !ok {
				continue
			}

			name, _, _ := unstructured.NestedString(container, "name")
			image, _, _ := unstructured.NestedString(container, "image")

			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, name))
			sb.WriteString(fmt.Sprintf("     Image: %s\n", image))

			// Container resources
			resources, found, _ := unstructured.NestedMap(container, "resources")
			if found {
				sb.WriteString("     Resources:\n")
				limits, limitsFound, _ := unstructured.NestedMap(resources, "limits")
				if limitsFound {
					for resource, value := range limits {
						sb.WriteString(fmt.Sprintf("       Limits %s: %v\n", resource, value))
					}
				}

				requests, requestsFound, _ := unstructured.NestedMap(resources, "requests")
				if requestsFound {
					for resource, value := range requests {
						sb.WriteString(fmt.Sprintf("       Requests %s: %v\n", resource, value))
					}
				}
			}

			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (f *PodFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No pods found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d pod(s):\n\n", len(list.Items)))

	// Group pods by namespace
	podsByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		podsByNamespace[namespace] = append(podsByNamespace[namespace], item)
	}

	// Print pods grouped by namespace
	for namespace, pods := range podsByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d pods)\n", namespace, len(pods)))

		for _, pod := range pods {
			// Get pod status
			status := getNestedString(pod.Object, "status", "phase")
			reason := getNestedString(pod.Object, "status", "reason")
			if reason != "" {
				status = reason
			}

			// Add ready container count
			ready := 0
			containerStatuses, _, _ := unstructured.NestedSlice(pod.Object, "status", "containerStatuses")
			for _, csObj := range containerStatuses {
				cs, ok := csObj.(map[string]interface{})
				if ok {
					isReady, found, _ := unstructured.NestedBool(cs, "ready")
					if found && isReady {
						ready++
					}
				}
			}

			containers, _, _ := unstructured.NestedSlice(pod.Object, "spec", "containers")

			// Basic pod info
			sb.WriteString(fmt.Sprintf("  • %s\n", pod.GetName()))
			sb.WriteString(fmt.Sprintf("    Status: %s\n", status))
			sb.WriteString(fmt.Sprintf("    Ready: %d/%d containers\n", ready, len(containers)))

			// Add node name if scheduled
			nodeName := getNestedString(pod.Object, "spec", "nodeName")
			if nodeName != "" {
				sb.WriteString(fmt.Sprintf("    Node: %s\n", nodeName))
			}

			// Add IP if assigned
			podIP := getNestedString(pod.Object, "status", "podIP")
			if podIP != "" {
				sb.WriteString(fmt.Sprintf("    IP: %s\n", podIP))
			}

			// Add creation time
			creationTime := pod.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			// Add restart count
			totalRestarts := 0
			for _, csObj := range containerStatuses {
				cs, ok := csObj.(map[string]interface{})
				if ok {
					restarts, found, _ := unstructured.NestedInt64(cs, "restartCount")
					if found {
						totalRestarts += int(restarts)
					}
				}
			}
			sb.WriteString(fmt.Sprintf("    Restarts: %d\n", totalRestarts))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// ServiceFormatter handles formatting for Service resources
type ServiceFormatter struct{}

func (f *ServiceFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Service: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Type
	svcType := getNestedString(res.Object, "spec", "type")
	sb.WriteString(fmt.Sprintf("Type: %s\n", svcType))

	// ClusterIP
	clusterIP := getNestedString(res.Object, "spec", "clusterIP")
	sb.WriteString(fmt.Sprintf("Cluster IP: %s\n", clusterIP))

	// External IPs
	externalIPs, _, _ := unstructured.NestedStringSlice(res.Object, "spec", "externalIPs")
	if len(externalIPs) > 0 {
		sb.WriteString("External IPs:\n")
		for _, ip := range externalIPs {
			sb.WriteString(fmt.Sprintf("  %s\n", ip))
		}
	}

	// Selectors
	selector, selectorFound, _ := unstructured.NestedMap(res.Object, "spec", "selector")
	if selectorFound && len(selector) > 0 {
		sb.WriteString("\nSelector:\n")
		for key, value := range selector {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Ports
	ports, _, _ := unstructured.NestedSlice(res.Object, "spec", "ports")
	if len(ports) > 0 {
		sb.WriteString("\nPorts:\n")
		for _, portObj := range ports {
			port, ok := portObj.(map[string]interface{})
			if !ok {
				continue
			}

			portNumber, _, _ := unstructured.NestedInt64(port, "port")
			targetPort, _, _ := unstructured.NestedFieldNoCopy(port, "targetPort")
			protocol, _, _ := unstructured.NestedString(port, "protocol")
			name, _, _ := unstructured.NestedString(port, "name")

			if name != "" {
				sb.WriteString(fmt.Sprintf("  %s:\n", name))
				sb.WriteString(fmt.Sprintf("    Port: %d\n", portNumber))
				sb.WriteString(fmt.Sprintf("    Target Port: %v\n", targetPort))
				sb.WriteString(fmt.Sprintf("    Protocol: %s\n", protocol))
			} else {
				sb.WriteString(fmt.Sprintf("  Port: %d\n", portNumber))
				sb.WriteString(fmt.Sprintf("  Target Port: %v\n", targetPort))
				sb.WriteString(fmt.Sprintf("  Protocol: %s\n", protocol))
			}
			sb.WriteString("\n")
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	return sb.String()
}

func (f *ServiceFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No services found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d service(s):\n\n", len(list.Items)))

	// Group services by namespace
	servicesByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		servicesByNamespace[namespace] = append(servicesByNamespace[namespace], item)
	}

	// Print services grouped by namespace
	for namespace, services := range servicesByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d services)\n", namespace, len(services)))

		for _, svc := range services {
			// Type
			svcType := getNestedString(svc.Object, "spec", "type")

			// ClusterIP
			clusterIP := getNestedString(svc.Object, "spec", "clusterIP")

			// Ports
			ports, _, _ := unstructured.NestedSlice(svc.Object, "spec", "ports")

			// Basic service info
			sb.WriteString(fmt.Sprintf("  • %s\n", svc.GetName()))
			sb.WriteString(fmt.Sprintf("    Type: %s\n", svcType))
			sb.WriteString(fmt.Sprintf("    Cluster IP: %s\n", clusterIP))

			if len(ports) > 0 {
				sb.WriteString("    Ports:\n")
				for _, portObj := range ports {
					port, ok := portObj.(map[string]interface{})
					if !ok {
						continue
					}

					portNumber, _, _ := unstructured.NestedInt64(port, "port")
					targetPort, _, _ := unstructured.NestedFieldNoCopy(port, "targetPort")
					protocol, _, _ := unstructured.NestedString(port, "protocol")
					name, _, _ := unstructured.NestedString(port, "name")

					portInfo := fmt.Sprintf("%d", portNumber)
					if name != "" {
						portInfo = fmt.Sprintf("%s:%d", name, portNumber)
					}

					sb.WriteString(fmt.Sprintf("      %s → %v/%s\n", portInfo, targetPort, protocol))
				}
			}

			// External IP if available
			externalIPs, _, _ := unstructured.NestedStringSlice(svc.Object, "spec", "externalIPs")
			if len(externalIPs) > 0 {
				sb.WriteString("    External IPs:\n")
				for _, ip := range externalIPs {
					sb.WriteString(fmt.Sprintf("      %s\n", ip))
				}
			}

			// Creation time
			creationTime := svc.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// NamespaceFormatter handles formatting for Namespace resources
type NamespaceFormatter struct{}

func (f *NamespaceFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetName()))

	// Status
	status := getNestedString(res.Object, "status", "phase")
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// Labels
	if labels := res.GetLabels(); len(labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Annotations
	if annotations := res.GetAnnotations(); len(annotations) > 0 {
		sb.WriteString("\nAnnotations:\n")
		for key, value := range annotations {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	return sb.String()
}

func (f *NamespaceFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No namespaces found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d namespace(s):\n\n", len(list.Items)))

	for _, ns := range list.Items {
		// Get status
		status := getNestedString(ns.Object, "status", "phase")

		// Basic namespace info
		sb.WriteString(fmt.Sprintf("• %s\n", ns.GetName()))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", status))

		// Creation time
		creationTime := ns.GetCreationTimestamp().Format(time.RFC3339)
		sb.WriteString(fmt.Sprintf("  Created: %s\n", creationTime))

		sb.WriteString("\n")
	}

	return sb.String()
}

// NodeFormatter handles formatting for Node resources
type NodeFormatter struct{}

func (f *NodeFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Node: %s\n", res.GetName()))

	// Node status
	status := "Unknown"
	conditions, _, _ := unstructured.NestedSlice(res.Object, "status", "conditions")
	for _, condObj := range conditions {
		cond, ok := condObj.(map[string]interface{})
		if !ok {
			continue
		}

		typeName, typeFound, _ := unstructured.NestedString(cond, "type")
		statusVal, statusFound, _ := unstructured.NestedString(cond, "status")

		if typeFound && statusFound && typeName == "Ready" {
			if statusVal == "True" {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Detailed conditions
	sb.WriteString("\nConditions:\n")
	for _, condObj := range conditions {
		cond, ok := condObj.(map[string]interface{})
		if !ok {
			continue
		}

		typeName, _, _ := unstructured.NestedString(cond, "type")
		statusVal, _, _ := unstructured.NestedString(cond, "status")
		message, _, _ := unstructured.NestedString(cond, "message")

		sb.WriteString(fmt.Sprintf("  %s: %s\n", typeName, statusVal))
		if message != "" {
			sb.WriteString(fmt.Sprintf("    Message: %s\n", message))
		}
	}

	// Addresses
	addresses, _, _ := unstructured.NestedSlice(res.Object, "status", "addresses")
	if len(addresses) > 0 {
		sb.WriteString("\nAddresses:\n")
		for _, addrObj := range addresses {
			addr, ok := addrObj.(map[string]interface{})
			if !ok {
				continue
			}

			addrType, typeFound, _ := unstructured.NestedString(addr, "type")
			addrVal, valFound, _ := unstructured.NestedString(addr, "address")

			if typeFound && valFound {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", addrType, addrVal))
			}
		}
	}

	// Node info
	nodeInfo, _, _ := unstructured.NestedMap(res.Object, "status", "nodeInfo")
	if len(nodeInfo) > 0 {
		sb.WriteString("\nNode Info:\n")
		for key, value := range nodeInfo {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Resources
	allocatable, _, _ := unstructured.NestedMap(res.Object, "status", "allocatable")
	capacity, _, _ := unstructured.NestedMap(res.Object, "status", "capacity")

	if len(allocatable) > 0 {
		sb.WriteString("\nAllocatable Resources:\n")
		for key, value := range allocatable {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	if len(capacity) > 0 {
		sb.WriteString("\nCapacity:\n")
		for key, value := range capacity {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *NodeFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No nodes found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d node(s):\n\n", len(list.Items)))

	for _, node := range list.Items {
		// Get node status
		status := "Ready"
		conditions, _, _ := unstructured.NestedSlice(node.Object, "status", "conditions")
		for _, condObj := range conditions {
			cond, ok := condObj.(map[string]interface{})
			if !ok {
				continue
			}

			typeName, typeFound, _ := unstructured.NestedString(cond, "type")
			status, statusFound, _ := unstructured.NestedString(cond, "status")

			if typeFound && statusFound && typeName == "Ready" {
				if status == "True" {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// Get addresses
		var internalIP, externalIP, hostname string
		addresses, _, _ := unstructured.NestedSlice(node.Object, "status", "addresses")
		for _, addrObj := range addresses {
			addr, ok := addrObj.(map[string]interface{})
			if !ok {
				continue
			}

			addrType, typeFound, _ := unstructured.NestedString(addr, "type")
			addrVal, valFound, _ := unstructured.NestedString(addr, "address")

			if typeFound && valFound {
				switch addrType {
				case "InternalIP":
					internalIP = addrVal
				case "ExternalIP":
					externalIP = addrVal
				case "Hostname":
					hostname = addrVal
				}
			}
		}

		// Get kubelet version
		kubeletVersion := getNestedString(node.Object, "status", "nodeInfo", "kubeletVersion")

		// Get allocatable resources
		allocatable, _, _ := unstructured.NestedMap(node.Object, "status", "allocatable")
		cpu := ""
		memory := ""
		if allocatable != nil {
			if cpuVal, ok := allocatable["cpu"]; ok {
				cpu = fmt.Sprintf("%v", cpuVal)
			}
			if memVal, ok := allocatable["memory"]; ok {
				memory = fmt.Sprintf("%v", memVal)
			}
		}

		// Basic node info
		sb.WriteString(fmt.Sprintf("• %s\n", node.GetName()))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", status))

		if internalIP != "" {
			sb.WriteString(fmt.Sprintf("  Internal IP: %s\n", internalIP))
		}

		if externalIP != "" {
			sb.WriteString(fmt.Sprintf("  External IP: %s\n", externalIP))
		}

		if hostname != "" && hostname != node.GetName() {
			sb.WriteString(fmt.Sprintf("  Hostname: %s\n", hostname))
		}

		if kubeletVersion != "" {
			sb.WriteString(fmt.Sprintf("  Kubelet Version: %s\n", kubeletVersion))
		}

		if cpu != "" || memory != "" {
			sb.WriteString("  Allocatable:\n")
			if cpu != "" {
				sb.WriteString(fmt.Sprintf("    CPU: %s\n", cpu))
			}
			if memory != "" {
				sb.WriteString(fmt.Sprintf("    Memory: %s\n", memory))
			}
		}

		// Creation time
		creationTime := node.GetCreationTimestamp().Format(time.RFC3339)
		sb.WriteString(fmt.Sprintf("  Created: %s\n", creationTime))

		sb.WriteString("\n")
	}

	return sb.String()
}

// DeploymentFormatter handles formatting for Deployment resources
type DeploymentFormatter struct{}

func (f *DeploymentFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Deployment: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Replicas info
	replicas := getNestedInt64(res.Object, "spec", "replicas")
	available := getNestedInt64(res.Object, "status", "availableReplicas")
	ready := getNestedInt64(res.Object, "status", "readyReplicas")
	updated := getNestedInt64(res.Object, "status", "updatedReplicas")

	sb.WriteString(fmt.Sprintf("Replicas: %d desired | %d updated | %d total | %d available | %d ready\n",
		replicas, updated, replicas, available, ready))

	// Strategy
	strategy := getNestedString(res.Object, "spec", "strategy", "type")
	sb.WriteString(fmt.Sprintf("Strategy: %s\n", strategy))

	// Selector
	selector, selectorFound, _ := unstructured.NestedMap(res.Object, "spec", "selector", "matchLabels")
	if selectorFound && len(selector) > 0 {
		sb.WriteString("\nSelector:\n")
		for key, value := range selector {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Containers
	containers, _, _ := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "containers")
	if len(containers) > 0 {
		sb.WriteString("\nContainers:\n")
		for i, containerObj := range containers {
			container, ok := containerObj.(map[string]interface{})
			if !ok {
				continue
			}

			name, _, _ := unstructured.NestedString(container, "name")
			image, _, _ := unstructured.NestedString(container, "image")

			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, name))
			sb.WriteString(fmt.Sprintf("     Image: %s\n", image))

			// Container resources
			resources, found, _ := unstructured.NestedMap(container, "resources")
			if found {
				sb.WriteString("     Resources:\n")
				limits, limitsFound, _ := unstructured.NestedMap(resources, "limits")
				if limitsFound {
					for resource, value := range limits {
						sb.WriteString(fmt.Sprintf("       Limits %s: %v\n", resource, value))
					}
				}

				requests, requestsFound, _ := unstructured.NestedMap(resources, "requests")
				if requestsFound {
					for resource, value := range requests {
						sb.WriteString(fmt.Sprintf("       Requests %s: %v\n", resource, value))
					}
				}
			}

			sb.WriteString("\n")
		}
	}

	// Conditions
	conditions, _, _ := unstructured.NestedSlice(res.Object, "status", "conditions")
	if len(conditions) > 0 {
		sb.WriteString("\nConditions:\n")
		for _, condObj := range conditions {
			cond, ok := condObj.(map[string]interface{})
			if !ok {
				continue
			}

			typeName, _, _ := unstructured.NestedString(cond, "type")
			statusVal, _, _ := unstructured.NestedString(cond, "status")
			reason, _, _ := unstructured.NestedString(cond, "reason")
			message, _, _ := unstructured.NestedString(cond, "message")

			sb.WriteString(fmt.Sprintf("  %s: %s\n", typeName, statusVal))
			if reason != "" {
				sb.WriteString(fmt.Sprintf("    Reason: %s\n", reason))
			}
			if message != "" {
				sb.WriteString(fmt.Sprintf("    Message: %s\n", message))
			}
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *DeploymentFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No deployments found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d deployment(s):\n\n", len(list.Items)))

	// Group deployments by namespace
	deploymentsByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		deploymentsByNamespace[namespace] = append(deploymentsByNamespace[namespace], item)
	}

	// Print deployments grouped by namespace
	for namespace, deployments := range deploymentsByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d deployments)\n", namespace, len(deployments)))

		for _, deployment := range deployments {
			// Get deployment status
			available := getNestedInt64(deployment.Object, "status", "availableReplicas")
			replicas := getNestedInt64(deployment.Object, "status", "replicas")
			updatedReplicas := getNestedInt64(deployment.Object, "status", "updatedReplicas")

			// Basic deployment info
			sb.WriteString(fmt.Sprintf("  • %s\n", deployment.GetName()))
			sb.WriteString(fmt.Sprintf("    Replicas: %d available / %d total / %d updated\n", available, replicas, updatedReplicas))

			// Add image info if available
			containers, _, _ := unstructured.NestedSlice(deployment.Object, "spec", "template", "spec", "containers")
			if len(containers) > 0 {
				sb.WriteString("    Images:\n")
				for _, containerObj := range containers {
					container, ok := containerObj.(map[string]interface{})
					if !ok {
						continue
					}

					name, _, _ := unstructured.NestedString(container, "name")
					image, _, _ := unstructured.NestedString(container, "image")

					sb.WriteString(fmt.Sprintf("      %s: %s\n", name, image))
				}
			}

			// Status conditions
			conditions, _, _ := unstructured.NestedSlice(deployment.Object, "status", "conditions")
			if len(conditions) > 0 {
				sb.WriteString("    Conditions:\n")
				for _, condObj := range conditions {
					cond, ok := condObj.(map[string]interface{})
					if !ok {
						continue
					}

					typeName, _, _ := unstructured.NestedString(cond, "type")
					statusVal, _, _ := unstructured.NestedString(cond, "status")

					sb.WriteString(fmt.Sprintf("      %s: %s\n", typeName, statusVal))
				}
			}

			// Creation time
			creationTime := deployment.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
