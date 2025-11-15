package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/HildaPosada/distributeai/worker/internal/client"
	"github.com/HildaPosada/distributeai/worker/internal/executor"
	"github.com/HildaPosada/distributeai/worker/internal/monitor"
	log "github.com/sirupsen/logrus"
)

type Worker struct {
	id              string
	name            string
	region          string
	cpuCores        int
	memoryGB        int
	gpuEnabled      bool
	client          *client.CoordinatorClient
	executor        *executor.DockerExecutor
	monitor         *monitor.SystemMonitor
	activeJobs      int
	stopChan        chan struct{}
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.Info("Starting DistributeAI Worker...")

	// Get configuration from environment
	coordinatorURL := getEnv("COORDINATOR_URL", "http://localhost:8080")
	workerID := getEnv("WORKER_ID", "worker-"+monitor.GetHostname())
	workerName := getEnv("WORKER_NAME", "Worker Node")
	region := getEnv("WORKER_REGION", "unknown")
	cpuCores := getEnvInt("CPU_CORES", 4)
	memoryGB := getEnvInt("MEMORY_GB", 8)
	gpuEnabled := getEnvBool("GPU_ENABLED", false)

	// Initialize components
	coordinatorClient := client.NewCoordinatorClient(coordinatorURL)

	dockerExecutor, err := executor.NewDockerExecutor()
	if err != nil {
		log.Fatalf("Failed to initialize Docker executor: %v", err)
	}
	defer dockerExecutor.Close()

	systemMonitor := monitor.NewSystemMonitor(cpuCores, memoryGB)

	worker := &Worker{
		id:         workerID,
		name:       workerName,
		region:     region,
		cpuCores:   cpuCores,
		memoryGB:   memoryGB,
		gpuEnabled: gpuEnabled,
		client:     coordinatorClient,
		executor:   dockerExecutor,
		monitor:    systemMonitor,
		stopChan:   make(chan struct{}),
	}

	// Register with coordinator
	if err := worker.register(); err != nil {
		log.Fatalf("Failed to register with coordinator: %v", err)
	}

	// Start worker loops
	go worker.heartbeatLoop()
	go worker.jobPollingLoop()

	log.Info("Worker started successfully")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	close(worker.stopChan)
	time.Sleep(2 * time.Second)
	log.Info("Worker stopped")
}

func (w *Worker) register() error {
	log.Infof("Registering worker %s with coordinator...", w.id)

	req := &client.NodeRegisterRequest{
		ID:         w.id,
		Name:       w.name,
		Region:     w.region,
		CPUCores:   w.cpuCores,
		MemoryGB:   w.memoryGB,
		GPUEnabled: w.gpuEnabled,
	}

	if err := w.client.RegisterNode(req); err != nil {
		return err
	}

	log.Info("Worker registered successfully")
	return nil
}

func (w *Worker) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.sendHeartbeat()
		case <-w.stopChan:
			return
		}
	}
}

func (w *Worker) sendHeartbeat() {
	heartbeat := &client.Heartbeat{
		CPUUsage:    w.monitor.GetCPUUsage(),
		MemoryUsage: w.monitor.GetMemoryUsage(),
		ActiveJobs:  w.activeJobs,
	}

	if err := w.client.SendHeartbeat(w.id, heartbeat); err != nil {
		log.Warnf("Failed to send heartbeat: %v", err)
	} else {
		log.Debug("Heartbeat sent")
	}
}

func (w *Worker) jobPollingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkForJobs()
		case <-w.stopChan:
			return
		}
	}
}

func (w *Worker) checkForJobs() {
	// Get pending jobs from coordinator
	jobs, err := w.client.GetPendingJobs(w.id)
	if err != nil {
		log.Warnf("Failed to get pending jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	log.Infof("Received %d pending job(s)", len(jobs))

	// Execute each job
	for _, pendingJob := range jobs {
		w.executeJob(pendingJob)
	}
}

func (w *Worker) executeJob(pendingJob client.PendingJob) {
	job := pendingJob.Job
	executionID := pendingJob.ExecutionID

	log.Infof("Executing job %s (%s)", job.ID, job.Name)

	w.activeJobs++
	defer func() { w.activeJobs-- }()

	// Execute the job
	ctx := context.Background()
	result := w.executor.ExecuteJob(
		ctx,
		job.DockerImage,
		job.Command,
		job.Environment,
		job.InputData,
	)

	// Prepare result submission
	submission := &client.JobResultSubmission{
		ExecutionID: executionID,
		JobID:       job.ID,
		NodeID:      w.id,
		Logs:        result.Logs,
	}

	if result.Success {
		submission.Result = result.Output
		submission.ResultHash = result.OutputHash
		log.Infof("Job %s completed successfully. Hash: %s", job.ID, result.OutputHash[:16])
	} else {
		submission.ErrorMessage = result.Error
		log.Errorf("Job %s failed: %s", job.ID, result.Error)
	}

	// Submit result to coordinator
	if err := w.client.SubmitJobResult(submission); err != nil {
		log.Errorf("Failed to submit job result: %v", err)
	} else {
		log.Infof("Job result submitted for %s", job.ID)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}
