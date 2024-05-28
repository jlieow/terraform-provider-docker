package provider

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &imagePushResource{}
	_ resource.ResourceWithConfigure = &imagePushResource{}
)

// NewimagePushResource is a helper function to simplify the provider implementation.
func NewImagePushResource() resource.Resource {
	return &imagePushResource{}
}

// imagePushResource is the resource implementation.
type imagePushResource struct {
	client *client.Client
}

// Metadata returns the resource type name.
func (r *imagePushResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_push"
}

type imagePushResourceModel struct {
	Image         types.String `tfsdk:"image"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	ServerAddress types.String `tfsdk:"serveraddress"`
	IdentityToken types.String `tfsdk:"identitytoken"`
	RegistryToken types.String `tfsdk:"registrytoken"`
}

// Schema defines the schema for the resource.
func (r *imagePushResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"image": schema.StringAttribute{
				Description: "Repository and tag of the image in the format repository:tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username of AuthConfig struct as specified in https://pkg.go.dev/github.com/docker/docker/api/types/registry#AuthConfig",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password of AuthConfig struct as specified in https://pkg.go.dev/github.com/docker/docker/api/types/registry#AuthConfig",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"serveraddress": schema.StringAttribute{
				Description: "ServerAddress of AuthConfig struct as specified in https://pkg.go.dev/github.com/docker/docker/api/types/registry#AuthConfig",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identitytoken": schema.StringAttribute{
				Description: "IdentityToken is used to authenticate the user and get an access token for the registry as specified in https://pkg.go.dev/github.com/docker/docker/api/types/registry#AuthConfig",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registrytoken": schema.StringAttribute{
				Description: "RegistryToken is a bearer token to be sent to a registry as specified in https://pkg.go.dev/github.com/docker/docker/api/types/registry#AuthConfig",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *imagePushResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imagePushResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	authConfig := registry.AuthConfig{
		Username:      plan.Username.ValueString(),
		Password:      plan.Password.ValueString(),
		ServerAddress: plan.ServerAddress.ValueString(),
		IdentityToken: plan.IdentityToken.ValueString(),
		RegistryToken: plan.RegistryToken.ValueString(),
	}

	authConfigEncoded, _ := registry.EncodeAuthConfig(authConfig)

	reader, err := r.client.ImagePush(
		ctx,
		plan.Image.ValueString(),
		image.PushOptions{
			RegistryAuth: authConfigEncoded,
		})

	if err != nil {
		tflog.Debug(ctx, "Unable to push docker image")
		tflog.Debug(ctx, err.Error())
	}

	tflog.Debug(ctx, "Docker image pushed!")

	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	// check errors
	fmt.Println("buf.String()")
	fmt.Println(buf.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *imagePushResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *imagePushResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *imagePushResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// Configure adds the provider configured client to the data source.
func (r *imagePushResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
