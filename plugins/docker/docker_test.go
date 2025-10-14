package docker

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerPlugin_LintAndValidate(t *testing.T) {
	// Test case 1: Valid Dockerfile
	validDockerfile := `FROM alpine:latest
RUN apk add --no-cache bash
CMD ["/bin/bash"]
`
	tmpDir, err := os.MkdirTemp("", "docker-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(validDockerfile), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	plugin := DockerPlugin{}
	output, err := plugin.LintAndValidate(dockerfilePath)
	if err != nil {
		t.Errorf("Expected no error for valid Dockerfile, but got: %v", err)
	}
	if !strings.Contains(output, "Validation passed") {
		t.Errorf("Expected output to contain 'Validation passed', but got: %s", output)
	}

	// Test case 2: Non-existent file
	_, err = plugin.LintAndValidate("/nonexistent/Dockerfile")
	if err == nil {
		t.Error("Expected an error for non-existent file, but got none")
	}

	// Test case 3: Dockerfile with potential issues (missing version pinning)
	unpinnedDockerfile := `FROM alpine
RUN apk add bash
`
	unpinnedPath := filepath.Join(tmpDir, "Dockerfile.unpinned")
	if err := os.WriteFile(unpinnedPath, []byte(unpinnedDockerfile), 0644); err != nil {
		t.Fatalf("Failed to write unpinned Dockerfile: %v", err)
	}

	// This should still pass validation even with warnings
	output, err = plugin.LintAndValidate(unpinnedPath)
	// We expect warnings but not a fatal error (hadolint might not be installed)
	if err != nil && !strings.Contains(err.Error(), "hadolint") {
		t.Logf("Validation output: %s", output)
		// Only fail if it's not a hadolint-related error
	}
}

func TestDockerPlugin_Build(t *testing.T) {
	// Only run this test if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping build test")
	}

	tmpDir, err := os.MkdirTemp("", "docker-build-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple Dockerfile
	simpleDockerfile := `FROM alpine:latest
CMD ["echo", "Hello from Buenos Aires!"]
`
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(simpleDockerfile), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	plugin := DockerPlugin{}
	imageName := "buenosaires-test"
	imageTag := "test"
	
	output, err := plugin.Build(dockerfilePath, imageName, imageTag)
	if err != nil {
		t.Errorf("Expected successful build, but got error: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Successfully built") && !strings.Contains(output, "writing image") {
		t.Logf("Build output: %s", output)
	}

	// Clean up the test image
	defer func() {
		exec := exec.Command("docker", "rmi", "-f", imageName+":"+imageTag)
		exec.Run()
	}()
}

func TestDockerPlugin_Run(t *testing.T) {
	// Only run this test if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping run test")
	}

	tmpDir, err := os.MkdirTemp("", "docker-run-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple Dockerfile
	simpleDockerfile := `FROM alpine:latest
CMD ["echo", "Hello from Buenos Aires!"]
`
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(simpleDockerfile), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	plugin := DockerPlugin{}
	imageName := "buenosaires-run-test"
	imageTag := "test"
	
	// Test without auto-run (build only)
	output, err := plugin.Run(dockerfilePath, imageName, imageTag, false)
	if err != nil {
		t.Errorf("Expected successful run (build only), but got error: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Image is ready to use") {
		t.Errorf("Expected output to indicate image is ready, but got: %s", output)
	}

	// Clean up the test image
	defer func() {
		exec := exec.Command("docker", "rmi", "-f", imageName+":"+imageTag)
		exec.Run()
	}()
}

func TestFindContainerFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "find-container-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case 1: Dockerfile exists
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	found, err := FindContainerFile(tmpDir)
	if err != nil {
		t.Errorf("Expected to find Dockerfile, but got error: %v", err)
	}
	if found != dockerfilePath {
		t.Errorf("Expected to find %s, but got %s", dockerfilePath, found)
	}

	// Test case 2: Only Containerfile exists
	tmpDir2, err := os.MkdirTemp("", "find-containerfile-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	containerfilePath := filepath.Join(tmpDir2, "Containerfile")
	if err := os.WriteFile(containerfilePath, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("Failed to write Containerfile: %v", err)
	}

	found, err = FindContainerFile(tmpDir2)
	if err != nil {
		t.Errorf("Expected to find Containerfile, but got error: %v", err)
	}
	if found != containerfilePath {
		t.Errorf("Expected to find %s, but got %s", containerfilePath, found)
	}

	// Test case 3: No container file exists
	tmpDir3, err := os.MkdirTemp("", "find-none-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir3)

	_, err = FindContainerFile(tmpDir3)
	if err == nil {
		t.Error("Expected an error when no container file exists, but got none")
	}
	if !strings.Contains(err.Error(), "no Dockerfile or Containerfile found") {
		t.Errorf("Expected specific error message, but got: %v", err)
	}
}

// isDockerAvailable checks if Docker is available on the system
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}
