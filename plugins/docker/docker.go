package docker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerPlugin represents the Docker container plugin.
type DockerPlugin struct{}

// LintAndValidate performs validation and linting on the Dockerfile/Containerfile.
func (p *DockerPlugin) LintAndValidate(containerFilePath string) (string, error) {
	var finalOutput bytes.Buffer

	// 1. Check if the file exists
	if _, err := os.Stat(containerFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("container file does not exist: %s", containerFilePath)
	}

	finalOutput.WriteString(fmt.Sprintf("Validating container file: %s\n", filepath.Base(containerFilePath)))

	// 2. Basic syntax validation with docker build --dry-run (if available in Docker 20.10+)
	// Note: Not all Docker versions support --dry-run, so we'll skip this for now
	// and rely on hadolint for validation

	// 3. Lint with hadolint (Dockerfile linter)
	cmdHadolint := exec.Command("hadolint", containerFilePath)
	hadolintOutput, err := cmdHadolint.CombinedOutput()
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// hadolint returns exit code 1 for warnings/errors
			finalOutput.Write(hadolintOutput)
			if exitError.ExitCode() > 0 {
				// Check if it's just warnings or actual errors
				outputStr := string(hadolintOutput)
				if strings.Contains(outputStr, "error:") {
					return finalOutput.String(), fmt.Errorf("hadolint found errors: %w", err)
				}
				// Just warnings, continue
				finalOutput.WriteString("Linting completed with warnings.\n")
			}
		} else {
			// hadolint not found or other error
			finalOutput.WriteString("Warning: hadolint not found, skipping linting (install hadolint for better validation)\n")
		}
	} else {
		finalOutput.Write(hadolintOutput)
		if len(hadolintOutput) == 0 {
			finalOutput.WriteString("Linting completed - no issues found.\n")
		} else {
			finalOutput.WriteString("Linting completed.\n")
		}
	}

	finalOutput.WriteString("Validation passed.\n")
	return finalOutput.String(), nil
}

// Build builds the Docker image from the Dockerfile/Containerfile.
func (p *DockerPlugin) Build(containerFilePath, imageName, imageTag string) (string, error) {
	// Get the directory containing the Dockerfile/Containerfile
	containerDir := filepath.Dir(containerFilePath)
	
	// Construct the image name with tag
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)
	
	// Build the Docker image
	cmdBuild := exec.Command("docker", "build", 
		"-f", containerFilePath,
		"-t", fullImageName,
		containerDir,
	)
	
	output, err := cmdBuild.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("docker build failed: %w", err)
	}
	
	return string(output), nil
}

// Run builds and optionally runs the Docker container.
// For safety, by default we only build the image. Running containers requires explicit configuration.
func (p *DockerPlugin) Run(containerFilePath, imageName, imageTag string, autoRun bool) (string, error) {
	var finalOutput bytes.Buffer
	
	// First, build the image
	buildOutput, err := p.Build(containerFilePath, imageName, imageTag)
	finalOutput.WriteString("=== BUILD OUTPUT ===\n")
	finalOutput.WriteString(buildOutput)
	finalOutput.WriteString("\n")
	
	if err != nil {
		return finalOutput.String(), err
	}
	
	finalOutput.WriteString(fmt.Sprintf("Successfully built image: %s:%s\n", imageName, imageTag))
	
	// Optionally run the container (disabled by default for safety)
	if autoRun {
		fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)
		containerName := fmt.Sprintf("%s-%s", imageName, imageTag)
		
		// Remove old container if it exists
		exec.Command("docker", "rm", "-f", containerName).Run()
		
		// Run the container
		cmdRun := exec.Command("docker", "run", 
			"--name", containerName,
			"-d", // Run in detached mode
			fullImageName,
		)
		
		runOutput, err := cmdRun.CombinedOutput()
		finalOutput.WriteString("=== RUN OUTPUT ===\n")
		finalOutput.WriteString(string(runOutput))
		
		if err != nil {
			return finalOutput.String(), fmt.Errorf("docker run failed: %w", err)
		}
		
		finalOutput.WriteString(fmt.Sprintf("Successfully started container: %s\n", containerName))
	} else {
		finalOutput.WriteString("Container not started (auto_run disabled). Image is ready to use.\n")
	}
	
	return finalOutput.String(), nil
}

// FindContainerFile looks for Dockerfile or Containerfile in the specified directory.
// Returns the path to the file if found, or an error if not found.
func FindContainerFile(dir string) (string, error) {
	// Check for Dockerfile first
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		return dockerfilePath, nil
	}
	
	// Check for Containerfile
	containerfilePath := filepath.Join(dir, "Containerfile")
	if _, err := os.Stat(containerfilePath); err == nil {
		return containerfilePath, nil
	}
	
	return "", fmt.Errorf("no Dockerfile or Containerfile found in %s", dir)
}
