package provider

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &dockerProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dockerProvider{
			version: version,
		}
	}
}

// dockerProvider is the provider implementation.
type dockerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *dockerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "docker"
	resp.Version = p.version
}

func (p *dockerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Attributes: map[string]schema.Attribute{
		// 	"region": schema.StringAttribute{
		// 		Optional: true,
		// 	},
		// 	"access_key": schema.StringAttribute{
		// 		Optional: true,
		// 	},
		// 	"secret_key": schema.StringAttribute{
		// 		Optional:  true,
		// 		Sensitive: true,
		// 	},
		// },
	}
}

// dockerProviderModel maps provider schema data to a Go type.
// type dockerProviderModel struct {
// 	// Region    types.String `tfsdk:"region"`
// 	// AccessKey types.String `tfsdk:"access_key"`
// 	// SecretKey types.String `tfsdk:"secret_key"`
// }

func (p *dockerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	// // Retrieve provider data from configuration
	// var config dockerProviderModel
	// diags := req.Config.Get(ctx, &config)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // if config.Region.IsUnknown() {
	// // 	resp.Diagnostics.AddAttributeError(
	// // 		path.Root("region"),
	// // 		"Unknown Region",
	// // 		"The provider cannot create the Custom S3 client as there is an unknown configuration value for the AWS Region. ",
	// // 	)
	// // }
	// // if config.AccessKey.IsUnknown() {
	// // 	resp.Diagnostics.AddAttributeError(
	// // 		path.Root("access_key"),
	// // 		"Unknown Access Key value",
	// // 		"The provider cannot create the Custom S3 client as there is an unknown configuration value for the AWS Access Key. ",
	// // 	)
	// // }
	// // if config.SecretKey.IsUnknown() {
	// // 	resp.Diagnostics.AddAttributeError(
	// // 		path.Root("secret_key"),
	// // 		"Unknown Secret Key value",
	// // 		"The provider cannot create the Custom S3 client as there is an unknown configuration value for the AWS Secret Key. ",
	// // 	)
	// // }
	// // if resp.Diagnostics.HasError() {
	// // 	return
	// // }

	// // region := os.Getenv("AWS_REGION")
	// // access_key := os.Getenv("AWS_ACCESS_KEY_ID")
	// // secret_key := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// // if !config.Region.IsNull() {
	// // 	region = config.Region.ValueString()
	// // }

	// // if !config.AccessKey.IsNull() {
	// // 	access_key = config.AccessKey.ValueString()
	// // }

	// // if !config.SecretKey.IsNull() {
	// // 	secret_key = config.SecretKey.ValueString()
	// // }

	// // if region == "" {
	// // 	resp.Diagnostics.AddAttributeError(
	// // 		path.Root("region"),
	// // 		"Missing Region",
	// // 		"The provider cannot create the AWS client as there is a missing or empty value for the Region. ",
	// // 	)
	// // }

	// // if access_key == "" {
	// // 	resp.Diagnostics.AddAttributeError(
	// // 		path.Root("access_key"),
	// // 		"Missing Access Key",
	// // 		"The provider cannot create the AWS client as there is a missing or empty value for the Access Key. ",
	// // 	)
	// // }

	// // if secret_key == "" {
	// // 	resp.Diagnostics.AddAttributeError(
	// // 		path.Root("secret_key"),
	// // 		"Missing Secret Key",
	// // 		"The provider cannot create the AWS client as there is a missing or empty value for the Secret Key. ",
	// // 	)
	// // }
	// // if resp.Diagnostics.HasError() {
	// // 	return
	// // }

	// Create Docker client
	apiClient, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Make the Docker client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

// DataSources defines the data sources implemented in the provider.
func (p *dockerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		DataSourceDockerImage,
	}
}

// Resources defines the resources implemented in the provider.
func (p *dockerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewImageResource,
		NewImagePushResource,
	}
}
