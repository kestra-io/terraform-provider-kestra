package provider_v2

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &testResource{}
var _ resource.ResourceWithImportState = &testResource{}

func NewTestResource() resource.Resource {
	return &testResource{}
}

// testResource defines the resource implementation.
type testResource struct {
	providerData ProviderData
}

// testResourceModel describes the resource data model.
type testResourceModel struct {
	TestId    types.String `tfsdk:"test_id"`
	Namespace types.String `tfsdk:"namespace"`
	Content   types.String `tfsdk:"content"`
}

func (r *testResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test"
}

func (r *testResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Test resource",

		Attributes: map[string]schema.Attribute{
			"test_id": schema.StringAttribute{
				MarkdownDescription: "The Test id",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The Test namespace",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The actual Test YAML content",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *testResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ProviderData type, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerData = *providerData
}

func (r *testResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data testResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configuredNamespace := data.Namespace.ValueString()
	configuredTestId := data.TestId.ValueString()
	content := data.Content.ValueString()
	if !(contains(content, configuredNamespace) && contains(content, configuredTestId)) {
		resp.Diagnostics.AddError(
			"namespace and test_id should match the YAML Test",
			"The content field must contain both the namespace and test_id values.",
		)
	}
}

func (r *testResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan testResourceModel
	// Read Terraform plan plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, httpResponse, err := r.providerData.Client.TestSuitesAPI.CreateTestSuite(ctx, r.providerData.TenantId).Body(plan.Content.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create test, got error: %s, full httpResponse: %v", err, httpResponse))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created a test resource, res: %+v", created))
	plan.Namespace = types.StringValue(created.Namespace)
	plan.TestId = types.StringValue(created.Id)
	plan.Content = types.StringValue(*created.Source)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan testResourceModel

	// Read Terraform prior state plan into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	read, httpResponse, err := r.providerData.Client.TestSuitesAPI.GetTestSuite(ctx, plan.Namespace.ValueString(), plan.TestId.ValueString(), r.providerData.TenantId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Test, got error: %s, full httpResponse: %v", err, httpResponse))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a test resource, res: %v", read))
	plan.Namespace = types.StringValue(read.Namespace)
	plan.TestId = types.StringValue(read.Id)
	plan.Content = types.StringValue(*read.Source)

	// Save updated plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan testResourceModel

	// Read Terraform prior state plan into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updated, httpResponse, err := r.providerData.Client.TestSuitesAPI.UpdateTestSuite(ctx, plan.Namespace.ValueString(), plan.TestId.ValueString(), r.providerData.TenantId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Test, got error: %s, full httpResponse: %v", err, httpResponse))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated a test resource, res: %v", updated))
	plan.Namespace = types.StringValue(updated.Namespace)
	plan.TestId = types.StringValue(updated.Id)
	plan.Content = types.StringValue(*updated.Source)

	// Save updated plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan testResourceModel

	// Read Terraform prior state plan into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, httpResponse, err := r.providerData.Client.TestSuitesAPI.DeleteTestSuite(ctx, plan.Namespace.ValueString(), plan.TestId.ValueString(), r.providerData.TenantId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Test, got error: %s, full httpResponse: %v", err, httpResponse))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted a test resource: %s %s", plan.Namespace, plan.TestId))

	// Save updated plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// contains is a helper to check substring presence
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) > 0 && (stringIndex(s, substr) != -1)
}

// stringIndex is a helper for strings.Index
func stringIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
