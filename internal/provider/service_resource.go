// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	client *bindings.APIClient
}

// {
//  "access": [
//    {
//      "auth": true,
//      "endpoint": "string",
//      "internal": true
//    }
//  ],
//  "description": "This is a hello world service.",
//  "disabled": true,
//  "environment": "production",
//  "logging": {
//    "additionalProp1": {
//      "labels": {
//        "additionalProp1": "string",
//        "additionalProp2": "string",
//        "additionalProp3": "string"
//      }
//    },
//    "additionalProp2": {
//      "labels": {
//        "additionalProp1": "string",
//        "additionalProp2": "string",
//        "additionalProp3": "string"
//      }
//    },
//    "additionalProp3": {
//      "labels": {
//        "additionalProp1": "string",
//        "additionalProp2": "string",
//        "additionalProp3": "string"
//      }
//    }
//  },
//  "metadata": {
//    "labels": {
//      "additionalProp1": "string",
//      "additionalProp2": "string",
//      "additionalProp3": "string"
//    }
//  },
//  "name": "hello-world",
//  "remarks": "Make sure this service prints hello world on /",
//  "sources": {
//    "additionalProp1": {
//      "labels": {
//        "additionalProp1": "string",
//        "additionalProp2": "string",
//        "additionalProp3": "string"
//      }
//    },
//    "additionalProp2": {
//      "labels": {
//        "additionalProp1": "string",
//        "additionalProp2": "string",
//        "additionalProp3": "string"
//      }
//    },
//    "additionalProp3": {
//      "labels": {
//        "additionalProp1": "string",
//        "additionalProp2": "string",
//        "additionalProp3": "string"
//      }
//    }
//  },
//  "tenant": "startrail",
//  "updated_at": "2021-01-01T00:00:00.000000",
//  "updated_by": "string",
//  "updated_date": "string"
//}

// ServiceResourceModel describes the resource data model.
type ServiceResourceModel struct {
	Id          types.String   `tfsdk:"id"`
	Access      types.ListType `tfsdk:"access"`
	Description types.String   `tfsdk:"description"`
	Disabled    types.Bool     `tfsdk:"disabled"`
	Environment types.String   `tfsdk:"environment"`
	Logging     types.MapType  `tfsdk:"logging"`
	Metadata    types.MapType  `tfsdk:"metadata"`
	Name        types.String   `tfsdk:"name"`
	Remarks     types.String   `tfsdk:"remarks"`
	Sources     types.MapType  `tfsdk:"sources"`
	Tenant      types.String   `tfsdk:"tenant"`
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
			"source": schema.ListNestedBlock{
				MarkdownDescription: "List of sources to use for the service, this is a map of source names to source configurations.",

				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"labels": schema.MapAttribute{
							Description: "Labels to apply to the service",
							Optional:    true,
						},
						"source": schema.StringAttribute{
							Description: "The source to use for the service",
							Required:    true,
						},
					},
				},
			},
			"metadata": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"labels": schema.MapAttribute{
						Description: "Labels to apply to the service",
						Optional:    true,
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
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Service description",
				Optional:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Service disabled",
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Service environment",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9-]+$`), "Environment must be lowercase alphanumeric with dashes"),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_by": schema.StringAttribute{
				Computed: true,
			},
			"updated_date": schema.StringAttribute{
				Computed: true,
			},
			"remarks": schema.StringAttribute{
				MarkdownDescription: "Service remarks",
				Optional:            true,
			},
			"tenant": schema.StringAttribute{
				MarkdownDescription: "Service tenant",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9-]+$`), "Tenant must be lowercase alphanumeric with dashes"),
				},
			},
		},
	}
}

func (r *ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bindings.APIClient)

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
	var data ServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", data.Tenant.String(), data.Environment.String(), data.Name.String()))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

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
	var data ServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clientReq := r.client.ServiceAPI.Create(ctx, data.Tenant.String())
	var access []bindings.Access
	req.Plan.GetAttribute(ctx, path.Root("access"), &access)
	var metadata bindings.Metadata
	req.Plan.GetAttribute(ctx, path.Root("metadata"), &metadata)
	var logging map[string]bindings.Logging
	req.Plan.GetAttribute(ctx, path.Root("logging"), &logging)
	var sources map[string]bindings.Source
	req.Plan.GetAttribute(ctx, path.Root("sources"), &sources)

	service := bindings.Service{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Remarks:     data.Remarks.ValueString(),
		Environment: data.Environment.ValueString(),
		Disabled:    data.Disabled.ValueBoolPointer(),
		Access:      access,
		Metadata:    metadata,
		Logging:     logging,
		Sources:     sources,
	}
	req.Plan.GetAttribute(ctx, path.Root("access"), &service.Access)
	clientReq.Service(service)
	execute, err := clientReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", err))
		return
	}
	if execute.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", execute.Body))
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

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clientReq := r.client.ServiceAPI.Delete(ctx, data.Tenant.String(), data.Environment.String(), data.Name.String())
	execute, err := clientReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", err))
		return
	}
	if execute.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", execute.Body))
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
