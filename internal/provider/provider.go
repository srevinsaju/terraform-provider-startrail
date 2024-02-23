// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bindings "github.com/srevinsaju/startrail-go-sdk"
	"net/http"
	"net/url"
)

// Ensure StartrailProvider satisfies various provider interfaces.
var _ provider.Provider = &StartrailProvider{}

// StartrailProvider defines the provider implementation.
type StartrailProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// StartrailProviderModel describes the provider data model.
type StartrailProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	ApiKey      types.String `tfsdk:"api_key"`
	Debug       types.Bool   `tfsdk:"debug"`
	Environment types.String `tfsdk:"environment"`
	Tenant      types.String `tfsdk:"tenant"`
}

type StartrailProviderClient struct {
	Client      *bindings.APIClient
	Tenant      string
	Environment string
}

func (p *StartrailProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "startrail"
	resp.Version = p.version
}

func (p *StartrailProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The upstream endpoint to use for API requests.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key to use for API requests.",
				Optional:            true,
				Sensitive:           true,
			},
			"tenant": schema.StringAttribute{
				MarkdownDescription: "The tenant to use for API requests.",
				Optional:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The environment to use for API requests.",
				Optional:            true,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Enable debug mode.",
				Optional:            true,
			},
		},
	}
}

func (p *StartrailProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data StartrailProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	u, err := url.Parse(data.Endpoint.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid endpoint", "The endpoint is not a valid URL")
		return
	}

	client := bindings.NewAPIClient(&bindings.Configuration{
		Host:   "",
		Scheme: "",
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ijl6U0ZFUnlBNl8wUk9CcV9Ddl9SdiJ9.eyJpc3MiOiJodHRwczovL3N0YXJ0cmFpbC1kZXYudXMuYXV0aDAuY29tLyIsInN1YiI6Imdvb2dsZS1vYXV0aDJ8MTAyNjc0NDg5NjUxMTA4MzUzMjA2IiwiYXVkIjpbImh0dHBzOi8vZGV2LnN0YXJ0cmFpbC5zcmV2LmluIiwiaHR0cHM6Ly9zdGFydHJhaWwtZGV2LnVzLmF1dGgwLmNvbS91c2VyaW5mbyJdLCJpYXQiOjE3MDg2ODQ0NTAsImV4cCI6MTcwODc3MDg1MCwiYXpwIjoiYVg0ckh0WFY4bTdSNW9UektPWlhkMXpVenFzeGlSN3ciLCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIn0.UXpEmR9tBpy878R42igkp3OVEUCTbSOTfG8uTFwNv1SqlWurx4cIVD6QIFdheJmAa2MI0ZWRweC3aEKDIDP07Vx1-7HJpmBz87T5TXVHFnPEcqYYwvJaKM5EvoumKYAjW2oIV6fyS_ig595AdM-gaxQAmZAuFNMnaXonQRc8AUC14eHyM54NDEp8Duzv1KZ8CieHwbHIb4m8ghoOyxQc32CZEhpH2NBRoGoPAcdqOt-pl_aVvfSLJWnH2l3_m4DqdWauyXyKiV3BVbFy86s9ZFE01lu-gQTgQQTPk_7hgevQb11P92XGKvC75B_U_ZBpb1IWzSTEqLUdYpcE0FxZeg"),
		},
		UserAgent: "startrail-terraform-provider/" + p.version,
		Debug:     data.Debug.ValueBool(),
		Servers: []bindings.ServerConfiguration{
			{
				URL: u.String(),
			},
		},
		OperationServers: nil,
		HTTPClient:       http.DefaultClient,
	})
	tenant := data.Tenant.ValueString()
	if tenant == "" {
		tenant = "default"
	}

	c := &StartrailProviderClient{
		Client:      client,
		Tenant:      tenant,
		Environment: data.Environment.ValueString(),
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *StartrailProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServiceResource,
	}
}

func (p *StartrailProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServiceDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StartrailProvider{
			version: version,
		}
	}
}
