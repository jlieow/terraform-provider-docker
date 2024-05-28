package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &dockerimageDataSource{}
	_ datasource.DataSourceWithConfigure = &dockerimageDataSource{}
)

// NewdockerimageDataSource is a helper function to simplify the provider implementation.
func DataSourceDockerImage() datasource.DataSource {
	return &dockerimageDataSource{}
}

// dockerimageDataSource is the data source implementation.
type dockerimageDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *dockerimageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// dockerimageDataSourceModel maps the data source schema data.
type dockerimageDataSourceModel struct {
	Images []dockerimageModel `tfsdk:"images"`
}

// dockerimageModel maps image schema data.
type dockerimageModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Tag     types.String `tfsdk:"tag"`
	Created types.String `tfsdk:"created"`
	Size    types.Int64  `tfsdk:"size"`
}

// Schema defines the schema for the data source.
func (d *dockerimageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"tag": schema.StringAttribute{
							Computed: true,
						},
						"created": schema.StringAttribute{
							Computed: true,
						},
						"size": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dockerimageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dockerimageDataSourceModel

	images, err := d.client.ImageList(context.Background(), dockertypes.ImageListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Docker Images, please ensure that docker daemon is up and running.",
			err.Error(),
		)
		return
	}

	for _, image := range images {

		name := "<none>"
		tag := "<none>"

		if len(image.RepoTags) > 0 {
			splitted := strings.Split(image.RepoTags[0], ":")
			name = splitted[0]
			tag = splitted[1]
		}

		// Converts unix timestamp to time object
		t := time.Unix(image.Created, 0)

		imagesState := dockerimageModel{
			ID:      types.StringValue(image.ID),
			Name:    types.StringValue(name),
			Tag:     types.StringValue(tag),
			Created: types.StringValue(t.String()),
			Size:    types.Int64Value(int64(image.Size)),
		}

		// resp.Diagnostics.AddWarning(image.ID, "comment")

		state.Images = append(state.Images, imagesState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *dockerimageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}
