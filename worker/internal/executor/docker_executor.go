package executor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type DockerExecutor struct {
	client *client.Client
}

type ExecutionResult struct {
	Output     string
	OutputHash string
	Logs       string
	Error      string
	Success    bool
}

func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerExecutor{client: cli}, nil
}

// ExecuteJob runs a job in a Docker container
func (e *DockerExecutor) ExecuteJob(
	ctx context.Context,
	dockerImage string,
	command []string,
	environment map[string]string,
	inputData string,
) *ExecutionResult {
	result := &ExecutionResult{}

	// Pull the image
	log.Infof("Pulling Docker image: %s", dockerImage)
	reader, err := e.client.ImagePull(ctx, dockerImage, types.ImagePullOptions{})
	if err != nil {
		result.Error = fmt.Sprintf("Failed to pull image: %v", err)
		return result
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	// Prepare environment variables
	var envVars []string
	for key, value := range environment {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	// Create container config
	containerConfig := &container.Config{
		Image:        dockerImage,
		Cmd:          command,
		Env:          envVars,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			Memory:   512 * 1024 * 1024, // 512MB limit
			NanoCPUs: 1000000000,        // 1 CPU
		},
	}

	// Create container
	resp, err := e.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create container: %v", err)
		return result
	}

	containerID := resp.ID

	// Start container
	if err := e.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		result.Error = fmt.Sprintf("Failed to start container: %v", err)
		return result
	}

	log.Infof("Container started: %s", containerID[:12])

	// Wait for container to finish (with timeout)
	statusCh, errCh := e.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			result.Error = fmt.Sprintf("Container wait error: %v", err)
			return result
		}
	case status := <-statusCh:
		log.Infof("Container finished with status: %d", status.StatusCode)

		// Get container logs
		out, err := e.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
		if err != nil {
			result.Error = fmt.Sprintf("Failed to get logs: %v", err)
			return result
		}
		defer out.Close()

		// Read logs
		logBytes, err := io.ReadAll(out)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to read logs: %v", err)
			return result
		}

		result.Logs = string(logBytes)
		result.Output = cleanDockerLogs(string(logBytes))

		// Generate hash of output for verification
		hash := sha256.Sum256([]byte(result.Output))
		result.OutputHash = hex.EncodeToString(hash[:])

		if status.StatusCode == 0 {
			result.Success = true
		} else {
			result.Error = fmt.Sprintf("Container exited with code %d", status.StatusCode)
		}
	case <-time.After(5 * time.Minute):
		// Timeout - kill container
		e.client.ContainerKill(ctx, containerID, "SIGKILL")
		result.Error = "Execution timeout (5 minutes)"
		return result
	}

	return result
}

// cleanDockerLogs removes Docker's stream headers from logs
func cleanDockerLogs(logs string) string {
	// Docker adds 8-byte headers to each line, remove them
	lines := strings.Split(logs, "\n")
	var cleaned []string

	for _, line := range lines {
		if len(line) > 8 {
			cleaned = append(cleaned, line[8:])
		} else if len(line) > 0 {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
}

func (e *DockerExecutor) Close() error {
	return e.client.Close()
}
