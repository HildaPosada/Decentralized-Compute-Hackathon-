package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/HildaPosada/distributeai/coordinator/internal/models"
	"github.com/HildaPosada/distributeai/coordinator/internal/repository"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	db *repository.Database
}

func NewHandler(db *repository.Database) *Handler {
	return &Handler{db: db}
}

// maxInt returns the larger of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// SubmitJob handles job submission
func (h *Handler) SubmitJob(c *gin.Context) {
	var req models.JobSubmitRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create job with defaults
	job := &models.Job{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		DockerImage:    req.DockerImage,
		Command:        req.Command,
		Environment:    req.Environment,
		InputData:      req.InputData,
		RequiredCPU:    maxInt(req.RequiredCPU, 1),
		RequiredMemory: maxInt(req.RequiredMemory, 1),
		RequiredGPU:    req.RequiredGPU,
		Redundancy:     3, // Default k-of-n: 3 nodes
		Consensus:      2, // Need 2 to agree
		Status:         models.JobStatusPending,
		SubmittedBy:    "user", // TODO: Add authentication
		SubmittedAt:    time.Now(),
		CreditsRequired: 1,
	}

	if err := h.db.CreateJob(job); err != nil {
		log.Errorf("Failed to create job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	log.Infof("Job %s submitted: %s", job.ID, job.Name)

	c.JSON(http.StatusCreated, job)
}

// GetJob retrieves a specific job by ID
func (h *Handler) GetJob(c *gin.Context) {
	jobID := c.Param("id")

	job, err := h.db.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs returns all jobs
func (h *Handler) ListJobs(c *gin.Context) {
	jobs, err := h.db.GetAllJobs()
	if err != nil {
		log.Errorf("Failed to get jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// GetJobExecutions returns all executions for a job
func (h *Handler) GetJobExecutions(c *gin.Context) {
	jobID := c.Param("id")

	executions, err := h.db.GetJobExecutions(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":     jobID,
		"executions": executions,
		"count":      len(executions),
	})
}

// RegisterNode handles worker node registration
func (h *Handler) RegisterNode(c *gin.Context) {
	var req models.NodeRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node := &models.Node{
		ID:              req.ID,
		Name:            req.Name,
		Region:          req.Region,
		CPUCores:        req.CPUCores,
		MemoryGB:        req.MemoryGB,
		GPUEnabled:      req.GPUEnabled,
		GPUModel:        req.GPUModel,
		Status:          models.NodeStatusOnline,
		ReputationScore: 100.0, // Start with perfect reputation
		LastHeartbeat:   time.Now(),
		RegisteredAt:    time.Now(),
	}

	if err := h.db.RegisterNode(node); err != nil {
		log.Errorf("Failed to register node: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register node"})
		return
	}

	log.Infof("Node registered: %s (%s)", node.ID, node.Name)

	c.JSON(http.StatusCreated, node)
}

// NodeHeartbeat handles heartbeat from worker nodes
func (h *Handler) NodeHeartbeat(c *gin.Context) {
	nodeID := c.Param("id")

	var heartbeat models.Heartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	heartbeat.NodeID = nodeID
	heartbeat.Timestamp = time.Now()

	if err := h.db.UpdateNodeHeartbeat(nodeID, &heartbeat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetNode retrieves a specific node by ID
func (h *Handler) GetNode(c *gin.Context) {
	nodeID := c.Param("id")

	node, err := h.db.GetNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

// ListNodes returns all nodes
func (h *Handler) ListNodes(c *gin.Context) {
	nodes, err := h.db.GetAllNodes()
	if err != nil {
		log.Errorf("Failed to get nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve nodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// GetPendingJobs returns jobs waiting for a worker (for workers to poll)
func (h *Handler) GetPendingJobs(c *gin.Context) {
	nodeID := c.Param("nodeId")

	// Get all scheduled executions for this node
	executions, err := h.db.GetJobExecutions("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get executions"})
		return
	}

	// Filter for this node and scheduled status
	var pendingJobs []map[string]interface{}

	for _, exec := range executions {
		if exec.NodeID == nodeID && exec.Status == models.JobStatusScheduled {
			job, err := h.db.GetJob(exec.JobID)
			if err != nil {
				continue
			}

			pendingJobs = append(pendingJobs, map[string]interface{}{
				"execution_id": exec.ID,
				"job":          job,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_jobs": pendingJobs,
		"count":        len(pendingJobs),
	})
}

// SubmitJobResult handles job result submission from workers
func (h *Handler) SubmitJobResult(c *gin.Context) {
	var result models.JobResultSubmission

	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the execution
	executions, err := h.db.GetJobExecutions(result.JobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution"})
		return
	}

	var execution *models.JobExecution
	for _, exec := range executions {
		if exec.ID == result.ExecutionID {
			execution = exec
			break
		}
	}

	if execution == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}

	// Update execution
	now := time.Now()
	execution.CompletedAt = &now
	execution.Result = result.Result
	execution.ResultHash = result.ResultHash
	execution.ErrorMessage = result.ErrorMessage
	execution.Logs = result.Logs

	if result.ErrorMessage != "" {
		execution.Status = models.JobStatusFailed
	} else {
		execution.Status = models.JobStatusCompleted
	}

	if err := h.db.UpdateJobExecution(execution); err != nil {
		log.Errorf("Failed to update execution: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update execution"})
		return
	}

	// Mark node as online again
	if err := h.db.UpdateNodeStatus(result.NodeID, models.NodeStatusOnline); err != nil {
		log.Warnf("Failed to update node status: %v", err)
	}

	log.Infof("Job result submitted: execution=%s, job=%s, node=%s",
		result.ExecutionID, result.JobID, result.NodeID)

	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

// GetStats returns system statistics
func (h *Handler) GetStats(c *gin.Context) {
	nodes, _ := h.db.GetAllNodes()
	jobs, _ := h.db.GetAllJobs()

	var onlineNodes, busyNodes int
	var totalCPU, totalMemory int

	for _, node := range nodes {
		if node.Status == models.NodeStatusOnline {
			onlineNodes++
		} else if node.Status == models.NodeStatusBusy {
			busyNodes++
		}
		totalCPU += node.CPUCores
		totalMemory += node.MemoryGB
	}

	var completedJobs, runningJobs, failedJobs int

	for _, job := range jobs {
		switch job.Status {
		case models.JobStatusCompleted:
			completedJobs++
		case models.JobStatusRunning, models.JobStatusScheduled:
			runningJobs++
		case models.JobStatusFailed:
			failedJobs++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": gin.H{
			"total":  len(nodes),
			"online": onlineNodes,
			"busy":   busyNodes,
		},
		"resources": gin.H{
			"total_cpu_cores": totalCPU,
			"total_memory_gb": totalMemory,
		},
		"jobs": gin.H{
			"total":     len(jobs),
			"completed": completedJobs,
			"running":   runningJobs,
			"failed":    failedJobs,
		},
	})
}