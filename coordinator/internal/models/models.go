package models

import (
	"time"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusScheduled  JobStatus = "scheduled"
	JobStatusRunning    JobStatus = "running"
	JobStatusVerifying  JobStatus = "verifying"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job represents a compute job to be executed
type Job struct {
	ID              string                 `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	DockerImage     string                 `json:"docker_image" db:"docker_image"`
	Command         []string               `json:"command" db:"command"`
	Environment     map[string]string      `json:"environment" db:"environment"`
	InputData       string                 `json:"input_data" db:"input_data"`
	RequiredCPU     int                    `json:"required_cpu" db:"required_cpu"`
	RequiredMemory  int                    `json:"required_memory" db:"required_memory"`
	RequiredGPU     bool                   `json:"required_gpu" db:"required_gpu"`
	Redundancy      int                    `json:"redundancy" db:"redundancy"` // How many nodes to run on
	Consensus       int                    `json:"consensus" db:"consensus"`   // How many must agree
	Status          JobStatus              `json:"status" db:"status"`
	SubmittedBy     string                 `json:"submitted_by" db:"submitted_by"`
	SubmittedAt     time.Time              `json:"submitted_at" db:"submitted_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Result          string                 `json:"result,omitempty" db:"result"`
	ErrorMessage    string                 `json:"error_message,omitempty" db:"error_message"`
	CreditsRequired int                    `json:"credits_required" db:"credits_required"`
}

// NodeStatus represents the current state of a worker node
type NodeStatus string

const (
	NodeStatusOnline   NodeStatus = "online"
	NodeStatusOffline  NodeStatus = "offline"
	NodeStatusBusy     NodeStatus = "busy"
	NodeStatusFaulty   NodeStatus = "faulty"
)

// Node represents a worker node in the network
type Node struct {
	ID               string     `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Region           string     `json:"region" db:"region"`
	CPUCores         int        `json:"cpu_cores" db:"cpu_cores"`
	MemoryGB         int        `json:"memory_gb" db:"memory_gb"`
	GPUEnabled       bool       `json:"gpu_enabled" db:"gpu_enabled"`
	GPUModel         string     `json:"gpu_model,omitempty" db:"gpu_model"`
	Status           NodeStatus `json:"status" db:"status"`
	ReputationScore  float64    `json:"reputation_score" db:"reputation_score"`
	TotalJobsRun     int        `json:"total_jobs_run" db:"total_jobs_run"`
	SuccessfulJobs   int        `json:"successful_jobs_run" db:"successful_jobs_run"`
	FailedJobs       int        `json:"failed_jobs" db:"failed_jobs"`
	CreditsEarned    int        `json:"credits_earned" db:"credits_earned"`
	LastHeartbeat    time.Time  `json:"last_heartbeat" db:"last_heartbeat"`
	RegisteredAt     time.Time  `json:"registered_at" db:"registered_at"`
	CurrentJobID     string     `json:"current_job_id,omitempty" db:"current_job_id"`
}

// JobExecution represents an instance of a job running on a specific node
type JobExecution struct {
	ID            string         `json:"id" db:"id"`
	JobID         string         `json:"job_id" db:"job_id"`
	NodeID        string         `json:"node_id" db:"node_id"`
	Status        JobStatus      `json:"status" db:"status"`
	StartedAt     time.Time      `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	Result        string         `json:"result,omitempty" db:"result"`
	ResultHash    string         `json:"result_hash,omitempty" db:"result_hash"`
	ErrorMessage  string         `json:"error_message,omitempty" db:"error_message"`
	Logs          string         `json:"logs,omitempty" db:"logs"`
}

// VerificationResult represents the outcome of k-of-n verification
type VerificationResult struct {
	JobID           string            `json:"job_id"`
	TotalExecutions int               `json:"total_executions"`
	ResultCounts    map[string]int    `json:"result_counts"`
	ConsensusResult string            `json:"consensus_result"`
	ConsensusReached bool             `json:"consensus_reached"`
	AgreementNodes  []string          `json:"agreement_nodes"`
	DisagreementNodes []string        `json:"disagreement_nodes"`
}

// Heartbeat represents a health check from a worker node
type Heartbeat struct {
	NodeID        string    `json:"node_id"`
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	ActiveJobs    int       `json:"active_jobs"`
}

// JobSubmitRequest represents the API request to submit a new job
type JobSubmitRequest struct {
	Name            string            `json:"name" binding:"required"`
	Description     string            `json:"description"`
	DockerImage     string            `json:"docker_image" binding:"required"`
	Command         []string          `json:"command" binding:"required"`
	Environment     map[string]string `json:"environment"`
	InputData       string            `json:"input_data"`
	RequiredCPU     int               `json:"required_cpu"`
	RequiredMemory  int               `json:"required_memory"`
	RequiredGPU     bool              `json:"required_gpu"`
}

// NodeRegisterRequest represents the API request for a node to register
type NodeRegisterRequest struct {
	ID         string `json:"id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Region     string `json:"region"`
	CPUCores   int    `json:"cpu_cores" binding:"required"`
	MemoryGB   int    `json:"memory_gb" binding:"required"`
	GPUEnabled bool   `json:"gpu_enabled"`
	GPUModel   string `json:"gpu_model"`
}

// JobResultSubmission represents a worker submitting a job result
type JobResultSubmission struct {
	ExecutionID  string `json:"execution_id" binding:"required"`
	JobID        string `json:"job_id" binding:"required"`
	NodeID       string `json:"node_id" binding:"required"`
	Result       string `json:"result"`
	ResultHash   string `json:"result_hash"`
	ErrorMessage string `json:"error_message"`
	Logs         string `json:"logs"`
}
