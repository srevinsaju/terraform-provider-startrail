// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ServiceDataSource{}

func NewServiceDataSource() datasource.DataSource {
	return &ServiceDataSource{}
}

// ServiceDataSource defines the data source implementation.
type ServiceDataSource struct {
	client *StartrailProviderClient
}

func (d *ServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *ServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Service data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Service identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Service name",
				Required:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Service environment",
				Required:            true,
			},
		},
	}
}

func (d *ServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*StartrailProviderClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment := data.Environment.ValueString()
	if environment == "" {
		environment = d.client.Environment
	}

	clientReq := d.client.Client.ServiceAPI.Get(ctx, d.client.Tenant, environment, data.Name.ValueString())
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
