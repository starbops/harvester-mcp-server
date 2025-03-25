package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ResourceFormatter defines the interface for formatting Kubernetes resources
type ResourceFormatter interface {
	FormatResource(res *unstructured.Unstructured) string
	FormatResourceList(list *unstructured.UnstructuredList) string
}

// FormatterRegistry maintains a mapping of resource kinds to their formatters
type FormatterRegistry struct {
	formatters map[string]ResourceFormatter
}

// NewFormatterRegistry creates a new registry with all registered formatters
func NewFormatterRegistry() *FormatterRegistry {
	registry := &FormatterRegistry{
		formatters: make(map[string]ResourceFormatter),
	}

	// Register core Kubernetes formatters
	registry.Register("Pod", &PodFormatter{})
	registry.Register("Service", &ServiceFormatter{})
	registry.Register("Namespace", &NamespaceFormatter{})
	registry.Register("Node", &NodeFormatter{})
	registry.Register("Deployment", &DeploymentFormatter{})

	// Register Harvester specific formatters
	registry.Register("VirtualMachine", &VirtualMachineFormatter{})
	registry.Register("Volume", &VolumeFormatter{})
	registry.Register("Network", &NetworkFormatter{})
	registry.Register("VirtualMachineImage", &VMImageFormatter{})
	registry.Register("CustomResourceDefinition", &CRDFormatter{})

	return registry
}

// defaultRegistry is a package-level registry instance for use by backward compatibility functions
var defaultRegistry = NewFormatterRegistry()

// Register adds a new formatter to the registry
func (r *FormatterRegistry) Register(kind string, formatter ResourceFormatter) {
	r.formatters[kind] = formatter
}

// GetFormatter returns the formatter for a specific resource kind
func (r *FormatterRegistry) GetFormatter(kind string) (ResourceFormatter, bool) {
	formatter, exists := r.formatters[kind]
	return formatter, exists
}

// FormatResource formats a single resource using the appropriate formatter
func (r *FormatterRegistry) FormatResource(res *unstructured.Unstructured) string {
	kind := res.GetKind()
	if formatter, exists := r.GetFormatter(kind); exists {
		return formatter.FormatResource(res)
	}
	return genericResourceFormatter(res)
}

// FormatResourceList formats a list of resources using the appropriate formatter
func (r *FormatterRegistry) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No resources found in the specified namespace(s)."
	}

	// Determine the kind from the first item
	if len(list.Items) > 0 {
		kind := list.Items[0].GetKind()
		if formatter, exists := r.GetFormatter(kind); exists {
			return formatter.FormatResourceList(list)
		}
	}

	return genericResourceListFormatter(list)
}

// genericResourceFormatter creates a human-readable representation of any resource
func genericResourceFormatter(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Kind: %s\n", res.GetKind()))
	sb.WriteString(fmt.Sprintf("Name: %s\n", res.GetName()))

	if ns := res.GetNamespace(); ns != "" {
		sb.WriteString(fmt.Sprintf("Namespace: %s\n", ns))
	}

	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// Print labels if any
	if labels := res.GetLabels(); len(labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	return sb.String()
}

// genericResourceListFormatter creates a human-readable list of any resources
func genericResourceListFormatter(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No resources found in the specified namespace(s)."
	}

	var sb strings.Builder
	kind := "resources"
	if len(list.Items) > 0 {
		kind = list.Items[0].GetKind() + "s"
	}

	sb.WriteString(fmt.Sprintf("Found %d %s:\n\n", len(list.Items), kind))

	// Group resources by namespace
	resourcesByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		resourcesByNamespace[namespace] = append(resourcesByNamespace[namespace], item)
	}

	// Print resources grouped by namespace
	for namespace, resources := range resourcesByNamespace {
		if namespace == "" {
			sb.WriteString(fmt.Sprintf("Cluster-scoped (%d %s)\n", len(resources), kind))
		} else {
			sb.WriteString(fmt.Sprintf("Namespace: %s (%d %s)\n", namespace, len(resources), kind))
		}

		for _, resource := range resources {
			sb.WriteString(fmt.Sprintf("  • %s\n", resource.GetName()))

			// Creation time
			creationTime := resource.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// The following functions maintain backward compatibility with any existing code that
// may call them directly. They now use the registry pattern internally.

// FormatPodList formats a list of Pod resources in a human-readable form
func FormatPodList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Pod")
	return formatter.FormatResourceList(list)
}

// FormatPod formats a Pod resource in a human-readable form
func FormatPod(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Pod")
	return formatter.FormatResource(res)
}

// FormatServiceList formats a list of Service resources in a human-readable form
func FormatServiceList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Service")
	return formatter.FormatResourceList(list)
}

// FormatService formats a Service resource in a human-readable form
func FormatService(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Service")
	return formatter.FormatResource(res)
}

// FormatNamespaceList formats a list of Namespace resources in a human-readable form
func FormatNamespaceList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Namespace")
	return formatter.FormatResourceList(list)
}

// FormatNamespace formats a Namespace resource in a human-readable form
func FormatNamespace(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Namespace")
	return formatter.FormatResource(res)
}

// FormatNodeList formats a list of Node resources in a human-readable form
func FormatNodeList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Node")
	return formatter.FormatResourceList(list)
}

// FormatNode formats a Node resource in a human-readable form
func FormatNode(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Node")
	return formatter.FormatResource(res)
}

// FormatDeploymentList formats a list of Deployment resources in a human-readable form
func FormatDeploymentList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Deployment")
	return formatter.FormatResourceList(list)
}

// FormatDeployment formats a Deployment resource in a human-readable form
func FormatDeployment(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Deployment")
	return formatter.FormatResource(res)
}

// FormatVirtualMachineList formats a list of VirtualMachine resources in a human-readable form
func FormatVirtualMachineList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("VirtualMachine")
	return formatter.FormatResourceList(list)
}

// FormatVirtualMachine formats a VirtualMachine resource in a human-readable form
func FormatVirtualMachine(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("VirtualMachine")
	return formatter.FormatResource(res)
}

// FormatVolumeList formats a list of Volume resources in a human-readable form
func FormatVolumeList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Volume")
	return formatter.FormatResourceList(list)
}

// FormatVolume formats a Volume resource in a human-readable form
func FormatVolume(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Volume")
	return formatter.FormatResource(res)
}

// FormatNetworkList formats a list of Network resources in a human-readable form
func FormatNetworkList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("Network")
	return formatter.FormatResourceList(list)
}

// FormatNetwork formats a Network resource in a human-readable form
func FormatNetwork(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("Network")
	return formatter.FormatResource(res)
}

// FormatImageList formats a list of VirtualMachineImage resources in a human-readable form
func FormatImageList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("VirtualMachineImage")
	return formatter.FormatResourceList(list)
}

// FormatImage formats a VirtualMachineImage resource in a human-readable form
func FormatImage(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("VirtualMachineImage")
	return formatter.FormatResource(res)
}

// FormatCRDList formats a list of CustomResourceDefinition resources in a human-readable form
func FormatCRDList(list *unstructured.UnstructuredList) string {
	formatter, _ := defaultRegistry.GetFormatter("CustomResourceDefinition")
	return formatter.FormatResourceList(list)
}

// FormatCRD formats a CustomResourceDefinition resource in a human-readable form
func FormatCRD(res *unstructured.Unstructured) string {
	formatter, _ := defaultRegistry.GetFormatter("CustomResourceDefinition")
	return formatter.FormatResource(res)
}

// For backward compatibility with any code that may be using these unexported functions
var (
	formatPodList            = FormatPodList
	formatPod                = FormatPod
	formatServiceList        = FormatServiceList
	formatService            = FormatService
	formatNamespaceList      = FormatNamespaceList
	formatNamespace          = FormatNamespace
	formatNodeList           = FormatNodeList
	formatNode               = FormatNode
	formatDeploymentList     = FormatDeploymentList
	formatDeployment         = FormatDeployment
	formatVirtualMachineList = FormatVirtualMachineList
	formatVirtualMachine     = FormatVirtualMachine
	formatVolumeList         = FormatVolumeList
	formatVolume             = FormatVolume
	formatNetworkList        = FormatNetworkList
	formatNetwork            = FormatNetwork
	formatImageList          = FormatImageList
	formatImage              = FormatImage
	formatCRDList            = FormatCRDList
	formatCRD                = FormatCRD
)
