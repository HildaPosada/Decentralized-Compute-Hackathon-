package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CoordinatorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCoordinatorClient(baseURL string) *CoordinatorClient {
	return &CoordinatorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NodeRegisterRequest matches coordinator model
type NodeRegisterRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	CPUCores   int    `json:"cpu_cores"`
	MemoryGB   int    `json:"memory_gb"`
	GPUEnabled bool   `json:"gpu_enabled"`
	GPUModel   string `json:"gpu_model"`
}

// Heartbeat matches coordinator model
type Heartbeat struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	ActiveJobs  int     `json:"active_jobs"`
}

// Job matches coordinator model (simplified)
type Job struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	DockerImage    string            `json:"docker_image"`
	Command        []string          `json:"command"`
	Environment    map[string]string `json:"environment"`
	InputData      string            `json:"input_data"`
	RequiredCPU    int               `json:"required_cpu"`
	RequiredMemory int               `json:"required_memory"`
}

// PendingJob wraps job with execution ID
type PendingJob struct {
	ExecutionID string `json:"execution_id"`
	Job         Job    `json:"job"`
}

// JobResultSubmission matches coordinator model
type JobResultSubmission struct {
	ExecutionID  string `json:"execution_id"`
	JobID        string `json:"job_id"`
	NodeID       string `json:"node_id"`
	Result       string `json:"result"`
	ResultHash   string `json:"result_hash"`
	ErrorMessage string `json:"error_message"`
	Logs         string `json:"logs"`
}

// RegisterNode registers this worker with the coordinator
func (c *CoordinatorClient) RegisterNode(req *NodeRegisterRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/nodes/register",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// SendHeartbeat sends a heartbeat to the coordinator
func (c *CoordinatorClient) SendHeartbeat(nodeID string, heartbeat *Heartbeat) error {
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v1/nodes/%s/heartbeat", c.baseURL, nodeID),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed: %s", resp.Status)
	}

	return nil
}

// GetPendingJobs fetches jobs assigned to this node
func (c *CoordinatorClient) GetPendingJobs(nodeID string) ([]PendingJob, error) {
	resp, err := c.httpClient.Get(
		fmt.Sprintf("%s/api/v1/nodes/%s/pending-jobs", c.baseURL, nodeID),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get pending jobs: %s", resp.Status)
	}

	var result struct {
		PendingJobs []PendingJob `json:"pending_jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.PendingJobs, nil
}

// SubmitJobResult submits the result of a completed job
func (c *CoordinatorClient) SubmitJobResult(result *JobResultSubmission) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/worker/result",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("result submission failed: %s - %s", resp.Status, string(body))
	}

	return nil
}
