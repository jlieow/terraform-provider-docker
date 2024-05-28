package provider

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestDirTraversalUnnested(t *testing.T) {

	ctx := context.Background()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	expectedDirFileCount := 3
	discoveredDirFileCount := traverseDirectoryAddFileToTar(ctx, tw, "../../tests/docker_image_resource_test/unnested")

	fmt.Println(discoveredDirFileCount)

	if expectedDirFileCount != discoveredDirFileCount {
		t.Fatalf("Directory/File count is incorrect! Expected number of directory/files is " + strconv.Itoa(expectedDirFileCount) + " but found " + strconv.Itoa(discoveredDirFileCount) + " directory/files.")
	}
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestDirTraversalNested(t *testing.T) {

	ctx := context.Background()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	expectedDirFileCount := 23
	discoveredDirFileCount := traverseDirectoryAddFileToTar(ctx, tw, "../../tests/docker_image_resource_test/nested")

	fmt.Println(discoveredDirFileCount)

	if expectedDirFileCount != discoveredDirFileCount {
		t.Fatalf("Directory/File count is incorrect! Expected number of directory/files is " + strconv.Itoa(expectedDirFileCount) + " but found " + strconv.Itoa(discoveredDirFileCount) + " directory/files.")
	}
}
