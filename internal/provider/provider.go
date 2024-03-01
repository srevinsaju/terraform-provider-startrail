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
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"os"
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

func newClient(u *url.URL, version string, authorization string, debug bool) *bindings.APIClient {
	client := bindings.NewAPIClient(&bindings.Configuration{
		Host:   "",
		Scheme: "",
		DefaultHeader: map[string]string{
			"Authorization": authorization,
		},
		UserAgent: "startrail-terraform-provider/" + version,
		Debug:     debug,
		Servers: []bindings.ServerConfiguration{
			{
				URL: u.String(),
			},
		},
		OperationServers: nil,
		HTTPClient:       http.DefaultClient,
	})
	return client
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
		resp.Diagnostics.AddError("Invalid endpoint", "The endpoint is not a valid URL, got error: "+err.Error())
		return
	}
	var token string

	if os.Getenv("STARTRAIL_TOKEN") == "" && os.Getenv("STARTRAIL_API_KEY") == "" || data.ApiKey.IsNull() {
		client := newClient(u, p.version, "", data.Debug.ValueBool())
		auth, exec, err := client.HelloAPI.WellKnownAuth(ctx).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to authenticate, got error: %s", err))
			return
		}
		if exec.StatusCode != 200 {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to authenticate, got status code: %d", exec.StatusCode))
			return
		}
		if !auth.Device.Enabled {
			resp.Diagnostics.AddError("Client Error", "Device flow is not enabled for the tenant. Please pass an 'api_key' instead")
			return
		}

		config := oauth2.Config{
			ClientID: auth.Device.GetClientId(),
			Endpoint: oauth2.Endpoint{
				AuthURL:       auth.Device.GetAuthorizationUrl(),
				DeviceAuthURL: auth.Device.GetDeviceCodeUrl(),
				TokenURL:      auth.Device.GetTokenUrl(),
				AuthStyle:     0,
			},
			RedirectURL: "",
			Scopes:      auth.Device.GetScopes(),
		}
		refreshToken, err := keyring.Get("startrail", "refresh_token")
		if err != nil {
			resp.Diagnostics.AddError("Client Error", "Unable to get refresh token from keyring, got error: "+err.Error())
			return
		}
		tokenSource := config.TokenSource(ctx, &oauth2.Token{
			RefreshToken: refreshToken,
		})
		t, err := tokenSource.Token()
		// write the new refresh token to the keyring
		if err != nil {
			resp.Diagnostics.AddError("Client Error", "Unable to get token from token source, got error: "+err.Error())
			return
		}
		if t.RefreshToken != "" {
			_ = keyring.Set("startrail", "refresh_token", t.RefreshToken)
		}
		token = fmt.Sprintf("Bearer %s", t.AccessToken)
	} else if os.Getenv("STARTRAIL_TOKEN") != "" {
		token = fmt.Sprintf("Bearer %s", os.Getenv("STARTRAIL_TOKEN"))
	} else if os.Getenv("STARTRAIL_API_KEY") != "" {
		token = fmt.Sprintf("apiKey %s", os.Getenv("STARTRAIL_API_KEY"))
	} else {
		token = fmt.Sprintf("apiKey %s", data.ApiKey.ValueString())
	}

	tenant := data.Tenant.ValueString()
	if tenant == "" {
		tenant = "default"
	}

	client := newClient(u, p.version, token, data.Debug.ValueBool())
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
