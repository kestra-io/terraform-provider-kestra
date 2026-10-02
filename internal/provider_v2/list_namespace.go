package provider_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ list.ListResource              = &namespaceListResource{}
	_ list.ListResourceWithConfigure = &namespaceListResource{}
)

type namespaceListResource struct {
	providerData *ProviderData
}

func NewNamespaceListResource() list.ListResource {
	return &namespaceListResource{}
}

func (r *namespaceListResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "kestra_namespace"
}

func (r *namespaceListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{}
}

func (r *namespaceListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok || providerData == nil {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("got %T", req.ProviderData),
		)
		return
	}

	r.providerData = providerData
}

func (r *namespaceListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.providerData == nil || r.providerData.KestraClient == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"List resource is not configured",
				"The Kestra provider client was not configured before listing namespaces.",
			),
		})
		return
	}

	pageSize := kestraListPageSize
	if req.Limit > 0 && req.Limit < int64(pageSize) {
		pageSize = int(req.Limit)
	}

	sort := []string{"id:asc"}
	results := make([]list.ListResult, 0)

	for page := 1; ; page++ {
		paged, err := r.providerData.KestraClient.Namespaces().SearchNamespaces(
			ctx,
			r.providerData.TenantId,
			nil,
			intPointer(page),
			intPointer(pageSize),
			sort,
			nil,
			nil,
		)
		if err != nil {
			results = append(results, listResultError(req, ctx, "List namespaces failed", err))
			break
		}

		namespaces := paged.GetResults()
		stop := false
		for _, namespace := range namespaces {
			result := req.NewListResult(ctx)
			namespaceID := namespace.GetId()
			result.DisplayName = namespaceID

			result.Diagnostics.Append(result.Identity.Set(ctx, namespaceIdentityModel{
				NamespaceId: types.StringValue(namespaceID),
			})...)

			if req.IncludeResource {
				fullNamespace, err := r.providerData.KestraClient.Namespaces().Namespace(
					ctx,
					namespaceID,
					r.providerData.TenantId,
				)
				if err != nil {
					result.Diagnostics.AddError(
						"Read namespace for list result failed",
						fmt.Sprintf("failed to load %s: %s", namespaceID, err),
					)
					results = append(results, result)
					stop = true
					break
				}

				raw, err := json.Marshal(fullNamespace)
				if err != nil {
					result.Diagnostics.AddError(
						"Encode namespace list result failed",
						fmt.Sprintf("failed to encode %s: %s", namespaceID, err),
					)
					results = append(results, result)
					stop = true
					break
				}

				body := map[string]interface{}{}
				if err := json.Unmarshal(raw, &body); err != nil {
					result.Diagnostics.AddError(
						"Decode namespace list result failed",
						fmt.Sprintf("failed to decode %s: %s", namespaceID, err),
					)
					results = append(results, result)
					stop = true
					break
				}

				model := namespaceModel{
					Id:          types.StringValue(namespaceID),
					TenantId:    types.StringValue(r.providerData.TenantId),
					NamespaceId: types.StringValue(namespaceID),
				}
				if fullNamespace.StorageIsolation != nil {
					model.StorageIsolation = []isolation{{}}
				}
				if fullNamespace.SecretIsolation != nil {
					model.SecretIsolation = []isolation{{}}
				}

				result.Diagnostics.Append(bodyToNamespaceModel(
					ctx,
					body,
					r.providerData.TenantId,
					&model,
				)...)
				result.Diagnostics.Append(result.Resource.Set(ctx, &model)...)
			}

			results = append(results, result)

			if req.Limit > 0 && int64(len(results)) >= req.Limit {
				stop = true
				break
			}
		}

		if stop || len(namespaces) == 0 || int64(len(namespaces)) < int64(pageSize) {
			break
		}

		if req.Limit > 0 && int64(len(results)) >= req.Limit {
			break
		}

		if paged.GetTotal() > 0 && int64(page*pageSize) >= paged.GetTotal() {
			break
		}
	}

	resp.Results = slices.Values(results)
}

func intPointer(value int) *int {
	return &value
}
