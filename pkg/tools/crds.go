package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListCRDs retrieves a list of Custom Resource Definitions from the Harvester cluster.
func ListCRDs(ctx context.Context, client *client.Client, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Create a new API extensions client
	apiextensionsClient, err := apiextensionsv1.NewForConfig(client.Config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create API extensions client: %v", err)), nil
	}

	crds, err := apiextensionsClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list CRDs: %v", err)), nil
	}

	// Filter Harvester-specific CRDs
	harvesterCRDs := &v1.CustomResourceDefinitionList{
		Items: []v1.CustomResourceDefinition{},
	}

	for _, crd := range crds.Items {
		if crd.Spec.Group == "harvesterhci.io" ||
			crd.Spec.Group == "kubevirt.io" ||
			crd.Spec.Group == "cdi.kubevirt.io" {
			harvesterCRDs.Items = append(harvesterCRDs.Items, crd)
		}
	}

	crdsJSON, err := json.MarshalIndent(harvesterCRDs, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to convert CRDs to JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(crdsJSON)), nil
}
