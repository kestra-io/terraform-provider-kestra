package provider_v2

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kestra-io/client-sdk/go-sdk/v2/kestra_api_client"
)

const kestraListPageSize = 1000

var (
	_ list.ListResource                 = &flowListResource{}
	_ list.ListResourceWithConfigure    = &flowListResource{}
	_ list.ListResourceWithRawV5Schemas = &flowListResource{}
)

type flowListResource struct {
	providerData        *ProviderData
	flowResourceFactory func() *sdkv2schema.Resource
}

type flowListIdentityModel struct {
	Namespace types.String `tfsdk:"namespace"`
	FlowID    types.String `tfsdk:"flow_id"`
}

type flowListResourceModel struct {
	ID        types.String `tfsdk:"id"`
	TenantID  types.String `tfsdk:"tenant_id"`
	Namespace types.String `tfsdk:"namespace"`
	FlowID    types.String `tfsdk:"flow_id"`
	Revision  types.Int64  `tfsdk:"revision"`
	Content   types.String `tfsdk:"content"`
}

func NewFlowListResource(flowResourceFactory func() *sdkv2schema.Resource) list.ListResource {
	return &flowListResource{
		flowResourceFactory: flowResourceFactory,
	}
}

func (r *flowListResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "kestra_flow"
}

func (r *flowListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{}
}

func (r *flowListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *flowListResource) RawV5Schemas(ctx context.Context, _ list.RawV5SchemaRequest, resp *list.RawV5SchemaResponse) {
	flowResource := r.flowResourceFactory()
	resp.ProtoV5Schema = flowResource.ProtoSchema(ctx)()
	resp.ProtoV5IdentitySchema = flowResource.ProtoIdentitySchema(ctx)()
}

func (r *flowListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.providerData == nil || r.providerData.KestraClient == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"List resource is not configured",
				"The Kestra provider client was not configured before listing flows.",
			),
		})
		return
	}

	pageSize := kestraListPageSize
	if req.Limit > 0 && req.Limit < int64(pageSize) {
		pageSize = int(req.Limit)
	}

	sort := []string{"namespace:asc", "id:asc"}
	results := make([]list.ListResult, 0)

	for page := 1; ; page++ {
		paged, err := r.providerData.KestraClient.Flows().SearchFlows(
			ctx,
			r.providerData.TenantId,
			kestra_api_client.PtrInt(page),
			kestra_api_client.PtrInt(pageSize),
			sort,
			nil,
		)
		if err != nil {
			results = append(results, listResultError(req, ctx, "List flows failed", err))
			break
		}

		flows := paged.GetResults()
		stop := false
		for _, flow := range flows {
			result := req.NewListResult(ctx)
			result.DisplayName = fmt.Sprintf("%s/%s", flow.GetNamespace(), flow.GetId())

			result.Diagnostics.Append(result.Identity.Set(ctx, flowListIdentityModel{
				Namespace: types.StringValue(flow.GetNamespace()),
				FlowID:    types.StringValue(flow.GetId()),
			})...)

			if req.IncludeResource {
				source := true
				fullFlow, err := r.providerData.KestraClient.Flows().Flow(
					ctx,
					flow.GetNamespace(),
					flow.GetId(),
					r.providerData.TenantId,
					&source,
					nil,
					nil,
				)
				if err != nil {
					result.Diagnostics.AddError(
						"Read flow for list result failed",
						fmt.Sprintf("failed to load %s: %s", result.DisplayName, err),
					)
					results = append(results, result)
					stop = true
					break
				}

				if fullFlow.Source == nil {
					result.Diagnostics.AddError(
						"Flow source is unavailable",
						fmt.Sprintf("Kestra returned no source for flow %s.", result.DisplayName),
					)
					results = append(results, result)
					stop = true
					break
				}

				revision := types.Int64Null()
				if fullFlow.Revision != nil {
					revision = types.Int64Value(int64(*fullFlow.Revision))
				}

				result.Diagnostics.Append(result.Resource.Set(ctx, flowListResourceModel{
					ID:        types.StringValue(fmt.Sprintf("%s/%s", fullFlow.GetNamespace(), fullFlow.GetId())),
					TenantID:  types.StringValue(r.providerData.TenantId),
					Namespace: types.StringValue(fullFlow.GetNamespace()),
					FlowID:    types.StringValue(fullFlow.GetId()),
					Revision:  revision,
					Content:   types.StringValue(*fullFlow.Source),
				})...)
			}

			results = append(results, result)

			if req.Limit > 0 && int64(len(results)) >= req.Limit {
				stop = true
				break
			}
		}

		if stop || len(flows) == 0 || int64(len(flows)) < int64(pageSize) {
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

func listResultError(req list.ListRequest, ctx context.Context, summary string, err error) list.ListResult {
	result := req.NewListResult(ctx)
	result.Diagnostics.AddError(summary, err.Error())
	return result
}
