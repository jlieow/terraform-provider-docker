package provider

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &imageResource{}
	_ resource.ResourceWithConfigure = &imageResource{}
)

// NewimageResource is a helper function to simplify the provider implementation.
func NewImageResource() resource.Resource {
	return &imageResource{}
}

// imageResource is the resource implementation.
type imageResource struct {
	client *client.Client
}

// Metadata returns the resource type name.
func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// Schema defines the schema for the resource.
func (r *imageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "SHA256 ID of the image.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListNestedAttribute{
				Description: "List of image tags.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"repository": schema.StringAttribute{
							Description: "Image name.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"tag": schema.StringAttribute{
							Description: "Image tag.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"dir": schema.StringAttribute{
				Description: "Path to the directory that contains the Dockerfile. Defaults to '\".\".",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the image was first built. Adding new tags does not update this value.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dockerfile_name": schema.StringAttribute{
				Description: "Name of the Dockerfile if a unique name is used.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"platform": schema.StringAttribute{
				Description: "Set platform of the build output.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nocache": schema.BoolAttribute{
				Description: "Specify whether to use cache when building the image.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"pullparent": schema.BoolAttribute{
				Description: "Specify whether to pull parent images when building the image.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

type imageResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Tags           []tagModel   `tfsdk:"tags"`
	Dir            types.String `tfsdk:"dir"`
	Created        types.String `tfsdk:"created"`
	DockerFileName types.String `tfsdk:"dockerfile_name"`
	Platform       types.String `tfsdk:"platform"`
	NoCache        types.Bool   `tfsdk:"nocache"`
	PullParent     types.Bool   `tfsdk:"pullparent"`
	// Size    types.Int64  `tfsdk:"size"`
}

type tagModel struct {
	Repository types.String `tfsdk:"repository"`
	Tag        types.String `tfsdk:"tag"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Defaults if not declared in terraform plan
	dir := "."
	if plan.Dir.ValueString() != "" {
		dir = plan.Dir.ValueString()
	}

	dockerFile := "Dockerfile"
	if plan.DockerFileName.ValueString() != "" {
		dockerFile = plan.DockerFileName.ValueString()
	}

	platform := "linux/arm64"
	if plan.Platform.ValueString() != "" {
		platform = plan.Platform.ValueString()
	}

	// Builds Image
	buildResponse, err := imageBuild(r, ctx, dir, dockerFile, plan.Tags, platform)

	if err != nil {
		tflog.Debug(ctx, "Unable to build docker image")
		tflog.Debug(ctx, err.Error())
	}
	defer buildResponse.Body.Close()

	// Check if build response can be parsed
	result, parseErr := parseDockerDaemonJsonMessages(buildResponse.Body)
	if parseErr != nil {
		tflog.Debug(ctx, "Unable to read image build response")
		fmt.Println(parseErr.Error())
	} else {
		tflog.Debug(ctx, "Successfully read image build response")
		fmt.Printf("%+v\n", "Build Response is: ")
		fmt.Printf("%+v\n", result)

		// Map response body to schema and populate Computed attribute values
		imageInspect, _, err := r.client.ImageInspectWithRaw(ctx, types.StringValue(result.ID).ValueString())
		if err != nil {
			// resp.Diagnostics.AddError(
			// 	"Error Reading Image",
			// 	"Could not read Image ID "+state.ID.ValueString()+": "+err.Error(),
			// )

			resp.State.RemoveResource(ctx)
			return
		}

		plan.ID = types.StringValue(imageInspect.ID)
		plan.Created = types.StringValue(imageInspect.Created)

		// Gets each tag, puts it into tagModel{} and appends to state.Tags
		plan.Tags = []tagModel{}
		for _, item := range imageInspect.RepoTags {
			repotagSplit := strings.Split(item, ":")

			plan.Tags = append(plan.Tags, tagModel{
				Repository: types.StringValue(repotagSplit[0]),
				Tag:        types.StringValue(repotagSplit[1]),
			})
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Returns the image information and its raw representation.
	imageInspect, _, err := r.client.ImageInspectWithRaw(ctx, state.ID.ValueString())
	if err != nil {
		// resp.Diagnostics.AddError(
		// 	"Error Reading Image",
		// 	"Could not read Image ID "+state.ID.ValueString()+": "+err.Error(),
		// )

		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(imageInspect.ID)
	state.Created = types.StringValue(imageInspect.Created)

	// Gets each tag, puts it into tagModel{} and appends to state.Tags
	state.Tags = []tagModel{}
	for _, item := range imageInspect.RepoTags {
		repotagSplit := strings.Split(item, ":")

		state.Tags = append(state.Tags, tagModel{
			Repository: types.StringValue(repotagSplit[0]),
			Tag:        types.StringValue(repotagSplit[1]),
		})
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// // Get current image
	// // Identifies tags that do not currently exist in the plan but have been provisioned
	// // Check for differences between provisioned image and image specified in plan
	// // If there is a difference in tags
	// // Removes tags if there are more tags
	// // Add tags if there are less tags

	// // Retrieve values from plan
	// var plan imageResourceModel
	// diags := req.Plan.Get(ctx, &plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// imageInspect, _, err := r.client.ImageInspectWithRaw(ctx, plan.ID.ValueString())
	// if err != nil {
	// 	// resp.Diagnostics.AddError(
	// 	// 	"Error Reading Image",
	// 	// 	"Could not read Image ID "+state.ID.ValueString()+": "+err.Error(),
	// 	// )
	// 	return
	// }

	// provisionedTags := []tagModel{}
	// for _, item := range imageInspect.RepoTags {
	// 	repotagSplit := strings.Split(item, ":")

	// 	provisionedTags = append(provisionedTags, tagModel{
	// 		Repository: types.StringValue(repotagSplit[0]),
	// 		Tag:        types.StringValue(repotagSplit[1]),
	// 	})
	// }

	// // Identifies tags that do not currently exist in the plan but have been provisioned
	// uniqueTags := []tagModel{}
	// for _, currentTag := range provisionedTags {
	// 	exists := false
	// 	for _, planTag := range plan.Tags {
	// 		if currentTag == planTag {
	// 			exists = true
	// 		}
	// 	}

	// 	if !exists {
	// 		uniqueTags = append(uniqueTags, currentTag)
	// 	}
	// }

	// // // Prints unique Tags
	// // for _, uniqueTag := range uniqueTags {
	// // 	fmt.Println("uniqueTag")
	// // 	fmt.Println(uniqueTag)
	// // }

	// if len(provisionedTags) > len(plan.Tags) {
	// 	fmt.Println("Time to remove tags!")

	// 	// Uses exec as the API does not support tag removal and requires removal of the entire image
	// 	for _, uniqueTag := range uniqueTags {

	// 		repotag := uniqueTag.Repository.ValueString() + ":" + uniqueTag.Tag.ValueString()

	// 		fmt.Println("Removing tag: " + repotag)

	// 		cmd := exec.Command("docker", "rmi", repotag)
	// 		stdout, err := cmd.Output()

	// 		if err != nil {
	// 			fmt.Println(err.Error())
	// 			return
	// 		}

	// 		// Print the output
	// 		fmt.Println(string(stdout))
	// 	}
	// }

	// if len(provisionedTags) < len(plan.Tags) {
	// 	fmt.Println("Time to add tags!")

	// 	buildResponse, err := imageBuild(r, ctx, plan.Dir.ValueString(), plan.DockerFileName.ValueString(), plan.Tags)

	// 	if err != nil {
	// 		tflog.Debug(ctx, "Unable to build docker image")
	// 		tflog.Debug(ctx, err.Error())
	// 	}
	// 	defer buildResponse.Body.Close()
	// }

	// // If there are same number of tags, but the tags are different
	// // Remove and rebuild image with correct tags
	// if len(provisionedTags) == len(plan.Tags) && len(uniqueTags) > 0 {
	// 	fmt.Println("Rebuild image with correct tags!")

	// 	_, err = r.client.ImageRemove(ctx, plan.ID.ValueString(), image.RemoveOptions{Force: true, PruneChildren: true})
	// 	if err != nil {
	// 		tflog.Debug(ctx, "Unable to remove docker image")
	// 		tflog.Debug(ctx, err.Error())
	// 	}

	// 	buildResponse, err := imageBuild(r, ctx, plan.Dir.ValueString(), plan.DockerFileName.ValueString(), uniqueTags)

	// 	if err != nil {
	// 		tflog.Debug(ctx, "Unable to build docker image")
	// 		tflog.Debug(ctx, err.Error())
	// 	}
	// 	defer buildResponse.Body.Close()

	// 	// Uses exec as the API does not support tag removal and requires removal of the entire image
	// 	for _, tag := range plan.Tags {

	// 		repotag := tag.Repository.ValueString() + ":" + tag.Tag.ValueString()

	// 		fmt.Println("Removing tag: " + repotag)

	// 		cmd := exec.Command("docker", "rmi", repotag)
	// 		stdout, err := cmd.Output()

	// 		if err != nil {
	// 			fmt.Println(err.Error())
	// 			return
	// 		}

	// 		// Print the output
	// 		fmt.Println(string(stdout))
	// 	}
	// }

	// // Map response body to schema and populate Computed attribute values
	// imageInspect, _, err = r.client.ImageInspectWithRaw(ctx, plan.ID.ValueString())
	// if err != nil {
	// 	// resp.Diagnostics.AddError(
	// 	// 	"Error Reading Image",
	// 	// 	"Could not read Image ID "+state.ID.ValueString()+": "+err.Error(),
	// 	// )
	// 	return
	// }

	// fmt.Println("imageInspect.RepoTags")
	// fmt.Println(imageInspect.RepoTags)

	// plan.Tags = []tagModel{}
	// for _, item := range imageInspect.RepoTags {
	// 	repotagSplit := strings.Split(item, ":")

	// 	plan.Tags = append(plan.Tags, tagModel{
	// 		Repository: types.StringValue(repotagSplit[0]),
	// 		Tag:        types.StringValue(repotagSplit[1]),
	// 	})
	// }

	// diags = resp.State.Set(ctx, plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Docker Image
	_, err := r.client.ImageRemove(ctx, state.ID.ValueString(), image.RemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		tflog.Debug(ctx, "Unable to remove docker image")
		tflog.Debug(ctx, err.Error())

		resp.Diagnostics.AddError(
			"Unable to remove docker image",
			"Could not remove docker image, unexpected error: "+err.Error(),
		)
	}
}

func (r *imageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// If there are no errors, Terraform will automatically call the Read method to import the rest of the Terraform state.
}

// Configure adds the provider configured client to the data source.
func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// func createTarFromDir(dir string, ctx context.Context) *bytes.Reader {

// 	buf := new(bytes.Buffer)
// 	tw := tar.NewWriter(buf)
// 	defer tw.Close()

// 	items, _ := os.ReadDir(dir)
// 	for _, item := range items {
// 		if item.IsDir() {
// 			subitems, _ := os.ReadDir(item.Name())
// 			for _, subitem := range subitems {
// 				if !subitem.IsDir() {
// 					// handle file there
// 					fmt.Println("****dirfile")
// 					fmt.Println(item.Name() + "/" + subitem.Name())
// 				}
// 			}
// 		} else {
// 			// handle file there
// 			fmt.Println("****file")
// 			fmt.Println(item.Name())

// 			addFileToTar(ctx, tw, dir, item.Name())
// 		}
// 	}

// 	buildContext := bytes.NewReader(buf.Bytes())

// 	return buildContext
// }

// Move inside each directory and write info to tar
// dirPath : folder which you want to tar it.
// tw      : its tarFile writer to your tar file.
func traverseDirectoryAddFileToTar(ctx context.Context, tw *tar.Writer, dirPath string) int {

	fileCount := 0

	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	defer dir.Close()
	// read all the files/dir in it
	fis, err := dir.Readdir(0)

	if err != nil {
		log.Fatal(err)
	}

	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()

		addFileToTar(ctx, tw, dirPath, fi.Name())
		if fi.IsDir() {
			fileCount += traverseDirectoryAddFileToTar(ctx, tw, curPath)
		}

		fmt.Println(curPath)

		fileCount += 1
	}

	return fileCount
}

func addFileToTar(ctx context.Context, tw *tar.Writer, dir string, fileName string) {

	fileDir := dir

	// Checks and ensures that the dir can be joined with the filename to create a proper path
	lastCharOfString := string(dir[len(dir)-1])
	if lastCharOfString != "/" {
		fileDir = dir + string("/")
	}

	filePath := fileDir + fileName

	fileReader, err := os.Open(filePath)

	if err != nil {
		tflog.Debug(ctx, " :****unable to open Dockerfile")
	}
	readFile, err := io.ReadAll(fileReader)
	if err != nil {
		tflog.Debug(ctx, " :****unable to read dockerfile")
	}

	tarHeader := &tar.Header{
		Name: fileName,
		Size: int64(len(readFile)),
	}
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		tflog.Debug(ctx, " :****unable to write tar header")
	}
	_, err = tw.Write(readFile)
	if err != nil {
		tflog.Debug(ctx, " :****unable to write tar body")
	}
}

func parseDockerDaemonJsonMessages(r io.Reader) (dockertypes.BuildResult, error) {
	var result dockertypes.BuildResult
	decoder := json.NewDecoder(r)
	for {
		var jsonMessage jsonmessage.JSONMessage
		if err := decoder.Decode(&jsonMessage); err != nil {
			if err == io.EOF {
				break
			}
			return result, err
		}
		if err := jsonMessage.Error; err != nil {
			return result, err
		}
		if jsonMessage.Aux != nil {
			var r dockertypes.BuildResult
			if err := json.Unmarshal(*jsonMessage.Aux, &r); err != nil {
				// logrus.Warnf("Failed to unmarshal aux message. Cause: %s", err)
			} else {
				result.ID = r.ID
			}
		}
	}
	return result, nil
}

func imageBuild(r *imageResource, ctx context.Context, planDir string, dockerFileName string, planTags []tagModel, planPlatform string) (dockertypes.ImageBuildResponse, error) {

	// Defaults if not declared in terraform plan
	dir := "."
	if planDir != "" {
		dir = planDir
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	traverseDirectoryAddFileToTar(ctx, tw, dir)

	buildContext := bytes.NewReader(buf.Bytes())

	// buildContext := createTarFromDir(dir, ctx)

	dockerFile := "Dockerfile"
	if dockerFileName != "" {
		dockerFile = dockerFileName
	}

	platform := ""
	if planPlatform != "" {
		platform = planPlatform
	}

	// Assign tags
	tags := []string{}
	for _, item := range planTags {
		imageTagName := item.Repository.ValueString() + string(":") + item.Tag.ValueString()
		tags = append(tags, imageTagName)
	}

	tflog.Debug(ctx, "Starting Image Build")

	buildResponse, err := r.client.ImageBuild(
		ctx,
		buildContext,
		dockertypes.ImageBuildOptions{
			Context:    buildContext,
			Dockerfile: dockerFile,
			Tags:       tags,
			Remove:     true,
			Platform:   platform,
			NoCache:    true,
			PullParent: true,
		})

	return buildResponse, err
}
