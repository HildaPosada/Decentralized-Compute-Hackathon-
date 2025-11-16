package scheduler

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/HildaPosada/distributeai/coordinator/internal/models"
	"github.com/HildaPosada/distributeai/coordinator/internal/repository"
	"github.com/HildaPosada/distributeai/coordinator/internal/verification"
	log "github.com/sirupsen/logrus"
)

// Scheduler handles job scheduling and distribution to worker nodes
type Scheduler struct {
	db       *repository.Database
	verifier *verification.Verifier
	stopChan chan struct{}
}

func NewScheduler(db *repository.Database, verifier *verification.Verifier) *Scheduler {
	return &Scheduler{
		db:       db,
		verifier: verifier,
		stopChan: make(chan struct{}),
	}
}

// Start begins the scheduling loop
func (s *Scheduler) Start() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info("Scheduler started")

	for {
		select {
		case <-ticker.C:
			s.schedulePendingJobs()
			s.checkRunningJobs()
			s.detectStaleNodes()
		case <-s.stopChan:
			log.Info("Scheduler stopped")
			return
		}
	}
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// schedulePendingJobs assigns pending jobs to available nodes
func (s *Scheduler) schedulePendingJobs() {
	jobs, err := s.db.GetPendingJobs()
	if err != nil {
		log.Errorf("Failed to get pending jobs: %v", err)
		return
	}

	for _, job := range jobs {
		if err := s.scheduleJob(job); err != nil {
			log.Errorf("Failed to schedule job %s: %v", job.ID, err)
		}
	}
}

// scheduleJob assigns a specific job to worker nodes
func (s *Scheduler) scheduleJob(job *models.Job) error {
	// Get available nodes that meet the requirements
	nodes, err := s.db.GetAvailableNodes(job.RequiredCPU, job.RequiredMemory, job.RequiredGPU)
	if err != nil {
		return fmt.Errorf("failed to get available nodes: %w", err)
	}

	if len(nodes) < job.Redundancy {
		log.Warnf("Not enough nodes available for job %s (need %d, have %d)",
			job.ID, job.Redundancy, len(nodes))
		return nil // Don't return error, just wait for more nodes
	}

	// Select top nodes based on reputation
	selectedNodes := nodes[:job.Redundancy]

	log.Infof("Scheduling job %s to %d nodes", job.ID, len(selectedNodes))

	// Create job executions for each selected node
	for _, node := range selectedNodes {
		execution := &models.JobExecution{
			ID:        uuid.New().String(),
			JobID:     job.ID,
			NodeID:    node.ID,
			Status:    models.JobStatusScheduled,
			StartedAt: time.Now(),
		}

		if err := s.db.CreateJobExecution(execution); err != nil {
			log.Errorf("Failed to create execution for node %s: %v", node.ID, err)
			continue
		}

		// Mark node as busy
		if err := s.db.UpdateNodeStatus(node.ID, models.NodeStatusBusy); err != nil {
			log.Warnf("Failed to update node status: %v", err)
		}
	}

	// Update job status to scheduled
	if err := s.db.UpdateJobStatus(job.ID, models.JobStatusScheduled, "", ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// checkRunningJobs monitors running jobs and performs verification
func (s *Scheduler) checkRunningJobs() {
	jobs, err := s.db.GetAllJobs()
	if err != nil {
		log.Errorf("Failed to get jobs: %v", err)
		return
	}

	for _, job := range jobs {
		// Skip non-active jobs
		if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
			continue
		}

		// Get executions for this job
		executions, err := s.db.GetJobExecutions(job.ID)
		if err != nil {
			log.Errorf("Failed to get executions for job %s: %v", job.ID, err)
			continue
		}

		// Count completed executions
		completedCount := 0
		failedCount := 0

		for _, exec := range executions {
			if exec.Status == models.JobStatusCompleted {
				completedCount++
			} else if exec.Status == models.JobStatusFailed {
				failedCount++
			}
		}

		// Check if we have enough completions to verify
		if completedCount >= job.Consensus {
			if err := s.verifier.CheckAndFinalizeJob(job.ID); err != nil {
				log.Errorf("Failed to verify job %s: %v", job.ID, err)
			}
		}

		// Check for job failure (too many failed executions)
		if failedCount > (job.Redundancy - job.Consensus) {
			log.Warnf("Job %s failed: too many execution failures", job.ID)
			s.db.UpdateJobStatus(job.ID, models.JobStatusFailed, "", "Too many execution failures")
		}

		// Handle stale jobs (scheduled but not progressing)
		if job.Status == models.JobStatusScheduled && job.SubmittedAt.Before(time.Now().Add(-10*time.Minute)) {
			// Check if any executions are actually running
			hasRunning := false
			for _, exec := range executions {
				if exec.Status == models.JobStatusRunning {
					hasRunning = true
					break
				}
			}

			if !hasRunning && completedCount == 0 {
				log.Warnf("Job %s appears stale, rescheduling", job.ID)
				s.db.UpdateJobStatus(job.ID, models.JobStatusPending, "", "")
			}
		}
	}
}

// detectStaleNodes marks nodes as offline if they haven't sent heartbeat
func (s *Scheduler) detectStaleNodes() {
	nodes, err := s.db.GetAllNodes()
	if err != nil {
		log.Errorf("Failed to get nodes: %v", err)
		return
	}

	staleThreshold := time.Now().Add(-2 * time.Minute)

	for _, node := range nodes {
		if node.Status == models.NodeStatusOnline && node.LastHeartbeat.Before(staleThreshold) {
			log.Warnf("Node %s is stale, marking offline", node.ID)
			if err := s.db.UpdateNodeStatus(node.ID, models.NodeStatusOffline); err != nil {
				log.Errorf("Failed to mark node offline: %v", err)
			}

			// Penalize reputation for going offline
			s.db.UpdateNodeReputation(node.ID, -20.0)
		}
	}
}
