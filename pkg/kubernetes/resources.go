package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/starbops/harvester-mcp-server/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

// ResourceHandler provides a unified interface for handling Kubernetes resources.
type ResourceHandler struct {
	client        *client.Client
	dynamicClient dynamic.Interface
	k8sClient     *kubernetes.Clientset
	mapper        *restmapper.DeferredDiscoveryRESTMapper
}

// NewResourceHandler creates a new ResourceHandler instance.
func NewResourceHandler(client *client.Client) (*ResourceHandler, error) {
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &ResourceHandler{
		client:        client,
		dynamicClient: dynamicClient,
		k8sClient:     client.Clientset,
	}, nil
}

// ListResources retrieves a list of resources of the specified type.
func (h *ResourceHandler) ListResources(ctx context.Context, gvr schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	if namespace == "" {
		return h.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	}
	return h.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// GetResource retrieves a specific resource by name.
func (h *ResourceHandler) GetResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	return h.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateResource creates a new resource.
func (h *ResourceHandler) CreateResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return h.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
}

// UpdateResource updates an existing resource.
func (h *ResourceHandler) UpdateResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return h.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
}

// DeleteResource deletes a resource by name.
func (h *ResourceHandler) DeleteResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	return h.dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// IsNamespaced determines if a resource type is namespaced or cluster-scoped.
func (h *ResourceHandler) IsNamespaced(gvr schema.GroupVersionResource) (bool, error) {
	apiResourceList, err := h.k8sClient.Discovery().ServerResourcesForGroupVersion(gvr.GroupVersion().String())
	if err != nil {
		return false, err
	}

	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Name == gvr.Resource {
			return apiResource.Namespaced, nil
		}
	}

	return false, fmt.Errorf("resource %s not found in group version %s", gvr.Resource, gvr.GroupVersion().String())
}

// FormatResourceList formats a list of resources into a human-readable string based on resource type.
func (h *ResourceHandler) FormatResourceList(list *unstructured.UnstructuredList, gvr schema.GroupVersionResource) string {
	switch {
	case gvr.Resource == "pods" && gvr.Group == "":
		return formatPodList(list)
	case gvr.Resource == "services" && gvr.Group == "":
		return formatServiceList(list)
	case gvr.Resource == "namespaces" && gvr.Group == "":
		return formatNamespaceList(list)
	case gvr.Resource == "nodes" && gvr.Group == "":
		return formatNodeList(list)
	case gvr.Resource == "deployments" && gvr.Group == "apps":
		return formatDeploymentList(list)
	case gvr.Resource == "virtualmachines" && gvr.Group == "kubevirt.io":
		return formatVirtualMachineList(list)
	case gvr.Resource == "networks" && gvr.Group == "network.harvesterhci.io":
		return formatNetworkList(list)
	case gvr.Resource == "volumes" && gvr.Group == "storage.harvesterhci.io":
		return formatVolumeList(list)
	case gvr.Resource == "virtualmachineimages" && gvr.Group == "harvesterhci.io":
		return formatImageList(list)
	case gvr.Resource == "customresourcedefinitions" && gvr.Group == "apiextensions.k8s.io":
		return formatCRDList(list)
	default:
		// Generic formatter for unsupported resource types
		return formatGenericResourceList(list, gvr)
	}
}

// FormatResource formats a single resource into a human-readable string based on resource type.
func (h *ResourceHandler) FormatResource(resource *unstructured.Unstructured, gvr schema.GroupVersionResource) string {
	switch {
	case gvr.Resource == "pods" && gvr.Group == "":
		return formatPod(resource)
	case gvr.Resource == "services" && gvr.Group == "":
		return formatService(resource)
	case gvr.Resource == "namespaces" && gvr.Group == "":
		return formatNamespace(resource)
	case gvr.Resource == "nodes" && gvr.Group == "":
		return formatNode(resource)
	case gvr.Resource == "deployments" && gvr.Group == "apps":
		return formatDeployment(resource)
	case gvr.Resource == "virtualmachines" && gvr.Group == "kubevirt.io":
		return formatVirtualMachine(resource)
	case gvr.Resource == "networks" && gvr.Group == "network.harvesterhci.io":
		return formatNetwork(resource)
	case gvr.Resource == "volumes" && gvr.Group == "storage.harvesterhci.io":
		return formatVolume(resource)
	case gvr.Resource == "virtualmachineimages" && gvr.Group == "harvesterhci.io":
		return formatImage(resource)
	case gvr.Resource == "customresourcedefinitions" && gvr.Group == "apiextensions.k8s.io":
		return formatCRD(resource)
	default:
		// Generic formatter for unsupported resource types
		return formatGenericResource(resource, gvr)
	}
}

// formatGenericResourceList creates a generic human-readable representation of resources
func formatGenericResourceList(list *unstructured.UnstructuredList, gvr schema.GroupVersionResource) string {
	if len(list.Items) == 0 {
		return fmt.Sprintf("No %s found in the specified namespace(s).", gvr.Resource)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d %s:\n\n", len(list.Items), gvr.Resource))

	// Group resources by namespace if they are namespaced
	resourcesByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		if namespace == "" {
			namespace = "cluster-scoped"
		}
		resourcesByNamespace[namespace] = append(resourcesByNamespace[namespace], item)
	}

	// Print resources grouped by namespace
	for namespace, items := range resourcesByNamespace {
		if namespace == "cluster-scoped" {
			sb.WriteString(fmt.Sprintf("Cluster-scoped resources (%d items)\n", len(items)))
		} else {
			sb.WriteString(fmt.Sprintf("Namespace: %s (%d items)\n", namespace, len(items)))
		}

		for _, item := range items {
			sb.WriteString(fmt.Sprintf("  • %s\n", item.GetName()))

			// Add creation time
			creationTime := item.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			// Add basic info from status if available
			status, found, _ := unstructured.NestedMap(item.Object, "status")
			if found {
				sb.WriteString("    Status:\n")
				for key, value := range status {
					sb.WriteString(fmt.Sprintf("      %s: %v\n", key, value))
				}
			}

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// formatGenericResource creates a generic human-readable representation of a single resource
func formatGenericResource(resource *unstructured.Unstructured, gvr schema.GroupVersionResource) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s: %s\n", strings.Title(gvr.Resource), resource.GetName()))

	if namespace := resource.GetNamespace(); namespace != "" {
		sb.WriteString(fmt.Sprintf("Namespace: %s\n", namespace))
	} else {
		sb.WriteString("Scope: Cluster-wide\n")
	}

	// Creation time
	creationTime := resource.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("Created: %s\n", creationTime))

	// Labels
	if labels := resource.GetLabels(); len(labels) > 0 {
		sb.WriteString("\nLabels:\n")
		for key, value := range labels {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Annotations
	if annotations := resource.GetAnnotations(); len(annotations) > 0 {
		sb.WriteString("\nAnnotations:\n")
		for key, value := range annotations {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Spec
	spec, found, _ := unstructured.NestedMap(resource.Object, "spec")
	if found {
		sb.WriteString("\nSpec:\n")
		for key, value := range spec {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Status
	status, found, _ := unstructured.NestedMap(resource.Object, "status")
	if found {
		sb.WriteString("\nStatus:\n")
		for key, value := range status {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	return sb.String()
}

// Helper function to safely get a nested string from an unstructured object
func getNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, _ := unstructured.NestedString(obj, fields...)
	if !found {
		return ""
	}
	return val
}

// Helper function to safely get a nested int64 from an unstructured object
func getNestedInt64(obj map[string]interface{}, fields ...string) int64 {
	val, found, _ := unstructured.NestedInt64(obj, fields...)
	if !found {
		return 0
	}
	return val
}

// Helper function to safely get a nested bool from an unstructured object
func getNestedBool(obj map[string]interface{}, fields ...string) bool {
	val, found, _ := unstructured.NestedBool(obj, fields...)
	if !found {
		return false
	}
	return val
}

// Helper function to safely get a nested string slice from an unstructured object
func getNestedStringSlice(obj map[string]interface{}, fields ...string) []string {
	val, found, _ := unstructured.NestedStringSlice(obj, fields...)
	if !found {
		return []string{}
	}
	return val
}

// Helper function to safely get a nested map from an unstructured object
func getNestedMap(obj map[string]interface{}, fields ...string) map[string]interface{} {
	val, found, _ := unstructured.NestedMap(obj, fields...)
	if !found {
		return map[string]interface{}{}
	}
	return val
}
