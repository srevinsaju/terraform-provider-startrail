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
	Endpoint types.String `tfsdk:"endpoint"`
	ApiKey   types.String `tfsdk:"api_key"`
	Debug    types.Bool   `tfsdk:"debug"`
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
	client := bindings.NewAPIClient(&bindings.Configuration{
		Host:   data.Endpoint.ValueString(),
		Scheme: "",
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("apiKey %s", data.ApiKey.ValueString()),
		},
		UserAgent:        "startrail-terraform-provider/" + p.version,
		Debug:            data.Debug.ValueBool(),
		Servers:          nil,
		OperationServers: nil,
		HTTPClient:       http.DefaultClient,
	})

	resp.DataSourceData = client
	resp.ResourceData = client
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
