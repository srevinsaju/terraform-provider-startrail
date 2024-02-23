// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	bindings "github.com/srevinsaju/startrail-go-sdk"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServiceResource{}
var _ resource.ResourceWithImportState = &ServiceResource{}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

// ServiceResource defines the resource implementation.
type ServiceResource struct {
	client *StartrailProviderClient
}

type ServiceResourceModelLogging struct {
	Labels types.Map    `tfsdk:"labels"`
	Source types.String `tfsdk:"source"`
}

type ServiceResourceM0delSource struct {
	Labels types.Map    `tfsdk:"labels"`
	Source types.String `tfsdk:"source"`
}

type ServiceResourceModelMetadata struct {
	Labels types.Map `tfsdk:"labels"`
}

type ServiceResourceModelAccess struct {
	Auth     types.Bool   `tfsdk:"auth"`
	Endpoint types.String `tfsdk:"endpoint"`
	Internal types.Bool   `tfsdk:"internal"`
}

func (r *ServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Service resource",
		Blocks: map[string]schema.Block{
			"access": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auth": schema.BoolAttribute{
							Description: "Set to true if this endpoint requires authentication to connect",
							Optional:    true,
						},
						"endpoint": schema.StringAttribute{
							Description: "The upstream endpoint to use for API requests.",
							Required:    true,
						},
						"internal": schema.BoolAttribute{
							Description: "Set to true if this endpoint is internal to the cluster",
							Optional:    true,
						},
					},
					Blocks:        nil,
					CustomType:    nil,
					Validators:    nil,
					PlanModifiers: nil,
				},
			},
			"logging": schema.ListNestedBlock{
				MarkdownDescription: "Logging configuration for the service",

				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"labels": schema.MapAttribute{
							Description: "Labels to apply to the service",
							Optional:    true,
							ElementType: types.StringType,
						},
						"source": schema.StringAttribute{
							Description: "The source to use for the service",
							Required:    true,
						},
					},
				},
			},
			"source": schema.ListNestedBlock{
				MarkdownDescription: "List of sources to use for the service, this is a map of source names to source configurations.",

				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"labels": schema.MapAttribute{
							Description: "Labels to apply to the service",
							Optional:    true,
							ElementType: types.StringType,
						},
						"source": schema.StringAttribute{
							Description: "The source to use for the service",
							Required:    true,
						},
					},
				},
			},
			"metadata": schema.SingleNestedBlock{
				MarkdownDescription: "Metadata to apply to the service",

				Attributes: map[string]schema.Attribute{
					"labels": schema.MapAttribute{
						Description: "Labels to apply to the service",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Service name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9-]+$`), "Name of the service must be lowercase alphanumeric with dashes"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Service description",
				Optional:            true,
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Service disabled",
				Computed:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Service environment",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9-]+$`), "Environment must be lowercase alphanumeric with dashes"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remarks": schema.StringAttribute{
				MarkdownDescription: "Service remarks",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*StartrailProviderClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, diags := r.post(ctx, data)
	resp.Diagnostics.Append(diags...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var data ServiceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment := data.Environment.ValueString()
	if environment == "" {
		environment = r.client.Environment
	}

	clientReq := r.client.Client.ServiceAPI.Get(ctx, r.client.Tenant, environment, data.Name.ValueString())
	startrailResponse, execute, err := clientReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", err))
		return
	}
	handleStartrailDiagnostics(startrailResponse.GetDiagnostics(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if execute.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", execute.Body))
		return
	}

	data, diags := parseServiceResponse(startrailResponse)
	resp.Diagnostics.Append(diags...)
	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, diags := r.post(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) post(ctx context.Context, data ServiceModel) (ServiceModel, diag.Diagnostics) {

	var diags diag.Diagnostics

	environment := data.Environment.ValueString()

	if environment == "" {
		environment = r.client.Environment
	}
	tenant := r.client.Tenant

	metadata := bindings.NullableMetadata{}
	logging := map[string]bindings.Logging{}
	sources := map[string]bindings.Source{}
	access := []bindings.Access{}

	for _, a := range data.Access {
		access = append(access, bindings.Access{
			Auth:     a.Auth.ValueBool(),
			Endpoint: a.Endpoint.ValueString(),
			Internal: a.Internal.ValueBool(),
		})
	}
	for _, l := range data.Logging {
		b := bindings.Logging{}
		l.Labels.ElementsAs(ctx, &b.Labels, true)
		logging[l.Source.ValueString()] = b
	}
	for _, s := range data.Sources {
		b := bindings.Source{}
		s.Labels.ElementsAs(ctx, &b.Labels, true)
		sources[s.Source.ValueString()] = b
	}

	service := bindings.Service{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Remarks:     data.Remarks.ValueString(),
		Environment: environment,
		Tenant:      tenant,

		Disabled: data.Disabled.ValueBoolPointer(),
		Access:   access,
		Metadata: metadata,
		Logging:  logging,
		Sources:  sources,
	}

	clientReq := r.client.Client.ServiceAPI.Create(ctx)
	clientReq = clientReq.Service(service)
	startrailResponse, execute, err := clientReq.Execute()

	// error handling
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", err))
		return ServiceModel{}, diags
	}
	handleStartrailDiagnostics(startrailResponse.GetDiagnostics(), &diags)
	if diags.HasError() {
		return ServiceModel{}, diags
	}
	if execute.StatusCode != 200 {
		diags.AddError("Client Error", fmt.Sprintf("Unable to update service, got errors: %s", execute.Body))
		return ServiceModel{}, diags
	}

	data, diags = parseServiceResponse(startrailResponse)
	return data, diags
}

func parseServiceResponse(startrailResponse *bindings.ServiceResponse) (data ServiceModel, diags diag.Diagnostics) {
	s := startrailResponse.GetResponse()

	var tfLogging []ServiceResourceModelLogging
	for k, v := range s.Logging {
		l := map[string]attr.Value{}
		for k1, v1 := range v.Labels {
			l[k1] = types.StringValue(v1)
		}
		labels, d := types.MapValue(types.StringType, l)
		if d.HasError() {
			diags.Append(d...)
		}
		tfLogging = append(tfLogging, ServiceResourceModelLogging{
			Labels: labels,
			Source: types.StringValue(k),
		})
	}
	var tfSources []ServiceResourceM0delSource
	for k, v := range s.Sources {
		l := map[string]attr.Value{}
		for k1, v1 := range v.Labels {
			l[k1] = types.StringValue(v1)
		}
		labels, d := types.MapValue(types.StringType, l)
		if d.HasError() {
			diags.Append(d...)
		}
		tfSources = append(tfSources, ServiceResourceM0delSource{
			Labels: labels,
			Source: types.StringValue(k),
		})
	}
	var tfMetadata *ServiceResourceModelMetadata
	if s.HasMetadata() {
		l := map[string]attr.Value{}
		m := s.GetMetadata()
		if m.GetLabels() != nil {
			for k1, v1 := range m.GetLabels() {
				l[k1] = types.StringValue(v1)
			}
			labels, d := types.MapValue(types.StringType, l)
			if d.HasError() {
				diags.Append(d...)
			}
			tfMetadata = &ServiceResourceModelMetadata{
				Labels: labels,
			}
		}

	}
	var tfAccess []ServiceResourceModelAccess
	for _, a := range s.Access {
		tfAccess = append(tfAccess, ServiceResourceModelAccess{
			Auth:     types.BoolValue(a.Auth),
			Endpoint: types.StringValue(a.Endpoint),
			Internal: types.BoolValue(a.Internal),
		})
	}

	data = ServiceModel{
		Id:          types.StringValue(fmt.Sprintf("%s/%s/%s", s.GetTenant(), s.GetEnvironment(), s.GetName())),
		Name:        types.StringValue(s.GetName()),
		Description: types.StringValue(s.GetDescription()),
		Remarks:     types.StringValue(s.GetRemarks()),
		Environment: types.StringValue(s.GetEnvironment()),
		Disabled:    types.BoolValue(s.GetDisabled()),
		Access:      tfAccess,
		Metadata:    tfMetadata,
		Logging:     tfLogging,
		Sources:     tfSources,
	}
	return data, diags
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment := data.Environment.ValueString()
	if environment == "" {
		environment = r.client.Environment
	}

	clientReq := r.client.Client.ServiceAPI.Delete(ctx, r.client.Tenant, environment, data.Name.ValueString())
	startrailResponse, execute, err := clientReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", err))
		return
	}
	handleStartrailDiagnostics(startrailResponse.GetDiagnostics(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if execute.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", execute.Body))
		return
	}
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
