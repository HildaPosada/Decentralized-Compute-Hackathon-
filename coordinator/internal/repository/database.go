package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/lib/pq"
	"github.com/HildaPosada/distributeai/coordinator/internal/models"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		docker_image VARCHAR(255) NOT NULL,
		command JSONB NOT NULL,
		environment JSONB,
		input_data TEXT,
		required_cpu INTEGER DEFAULT 1,
		required_memory INTEGER DEFAULT 1,
		required_gpu BOOLEAN DEFAULT FALSE,
		redundancy INTEGER DEFAULT 3,
		consensus INTEGER DEFAULT 2,
		status VARCHAR(50) NOT NULL,
		submitted_by VARCHAR(255),
		submitted_at TIMESTAMP NOT NULL,
		started_at TIMESTAMP,
		completed_at TIMESTAMP,
		result TEXT,
		error_message TEXT,
		credits_required INTEGER DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS nodes (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		region VARCHAR(100),
		cpu_cores INTEGER NOT NULL,
		memory_gb INTEGER NOT NULL,
		gpu_enabled BOOLEAN DEFAULT FALSE,
		gpu_model VARCHAR(255),
		status VARCHAR(50) NOT NULL,
		reputation_score REAL DEFAULT 100.0,
		total_jobs_run INTEGER DEFAULT 0,
		successful_jobs_run INTEGER DEFAULT 0,
		failed_jobs INTEGER DEFAULT 0,
		credits_earned INTEGER DEFAULT 0,
		last_heartbeat TIMESTAMP NOT NULL,
		registered_at TIMESTAMP NOT NULL,
		current_job_id VARCHAR(36)
	);

	CREATE TABLE IF NOT EXISTS job_executions (
		id VARCHAR(36) PRIMARY KEY,
		job_id VARCHAR(36) NOT NULL REFERENCES jobs(id),
		node_id VARCHAR(36) NOT NULL REFERENCES nodes(id),
		status VARCHAR(50) NOT NULL,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		result TEXT,
		result_hash VARCHAR(64),
		error_message TEXT,
		logs TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
	CREATE INDEX IF NOT EXISTS idx_executions_job_id ON job_executions(job_id);
	CREATE INDEX IF NOT EXISTS idx_executions_node_id ON job_executions(node_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// Job operations
func (d *Database) CreateJob(job *models.Job) error {
	commandJSON, _ := json.Marshal(job.Command)
	envJSON, _ := json.Marshal(job.Environment)

	_, err := d.db.Exec(`
		INSERT INTO jobs (id, name, description, docker_image, command, environment,
			input_data, required_cpu, required_memory, required_gpu, redundancy, consensus,
			status, submitted_by, submitted_at, credits_required)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		job.ID, job.Name, job.Description, job.DockerImage, commandJSON, envJSON,
		job.InputData, job.RequiredCPU, job.RequiredMemory, job.RequiredGPU,
		job.Redundancy, job.Consensus, job.Status, job.SubmittedBy,
		job.SubmittedAt, job.CreditsRequired,
	)
	return err
}

func (d *Database) GetJob(id string) (*models.Job, error) {
	var job models.Job
	var commandJSON, envJSON []byte

	err := d.db.QueryRow(`
		SELECT id, name, description, docker_image, command, environment, input_data,
			required_cpu, required_memory, required_gpu, redundancy, consensus, status,
			submitted_by, submitted_at, started_at, completed_at, result, error_message, credits_required
		FROM jobs WHERE id = $1`, id).Scan(
		&job.ID, &job.Name, &job.Description, &job.DockerImage, &commandJSON, &envJSON,
		&job.InputData, &job.RequiredCPU, &job.RequiredMemory, &job.RequiredGPU,
		&job.Redundancy, &job.Consensus, &job.Status, &job.SubmittedBy, &job.SubmittedAt,
		&job.StartedAt, &job.CompletedAt, &job.Result, &job.ErrorMessage, &job.CreditsRequired,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(commandJSON, &job.Command)
	json.Unmarshal(envJSON, &job.Environment)

	return &job, nil
}

func (d *Database) UpdateJobStatus(id string, status models.JobStatus, result, errorMsg string) error {
	now := time.Now()

	if status == models.JobStatusCompleted || status == models.JobStatusFailed {
		_, err := d.db.Exec(`
			UPDATE jobs SET status = $1, result = $2, error_message = $3, completed_at = $4
			WHERE id = $5`,
			status, result, errorMsg, now, id,
		)
		return err
	}

	_, err := d.db.Exec(`
		UPDATE jobs SET status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}

func (d *Database) GetPendingJobs() ([]*models.Job, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, docker_image, command, environment, input_data,
			required_cpu, required_memory, required_gpu, redundancy, consensus, status,
			submitted_by, submitted_at, credits_required
		FROM jobs WHERE status = $1 ORDER BY submitted_at ASC`,
		models.JobStatusPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		var commandJSON, envJSON []byte

		err := rows.Scan(
			&job.ID, &job.Name, &job.Description, &job.DockerImage, &commandJSON, &envJSON,
			&job.InputData, &job.RequiredCPU, &job.RequiredMemory, &job.RequiredGPU,
			&job.Redundancy, &job.Consensus, &job.Status, &job.SubmittedBy,
			&job.SubmittedAt, &job.CreditsRequired,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(commandJSON, &job.Command)
		json.Unmarshal(envJSON, &job.Environment)
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (d *Database) GetAllJobs() ([]*models.Job, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, docker_image, command, environment, input_data,
			required_cpu, required_memory, required_gpu, redundancy, consensus, status,
			submitted_by, submitted_at, started_at, completed_at, result, error_message, credits_required
		FROM jobs ORDER BY submitted_at DESC LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		var commandJSON, envJSON []byte

		err := rows.Scan(
			&job.ID, &job.Name, &job.Description, &job.DockerImage, &commandJSON, &envJSON,
			&job.InputData, &job.RequiredCPU, &job.RequiredMemory, &job.RequiredGPU,
			&job.Redundancy, &job.Consensus, &job.Status, &job.SubmittedBy,
			&job.SubmittedAt, &job.StartedAt, &job.CompletedAt, &job.Result,
			&job.ErrorMessage, &job.CreditsRequired,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(commandJSON, &job.Command)
		json.Unmarshal(envJSON, &job.Environment)
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// Node operations
func (d *Database) RegisterNode(node *models.Node) error {
	_, err := d.db.Exec(`
		INSERT INTO nodes (id, name, region, cpu_cores, memory_gb, gpu_enabled, gpu_model,
			status, reputation_score, last_heartbeat, registered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			region = EXCLUDED.region,
			cpu_cores = EXCLUDED.cpu_cores,
			memory_gb = EXCLUDED.memory_gb,
			gpu_enabled = EXCLUDED.gpu_enabled,
			gpu_model = EXCLUDED.gpu_model,
			status = EXCLUDED.status,
			last_heartbeat = EXCLUDED.last_heartbeat`,
		node.ID, node.Name, node.Region, node.CPUCores, node.MemoryGB,
		node.GPUEnabled, node.GPUModel, node.Status, node.ReputationScore,
		node.LastHeartbeat, node.RegisteredAt,
	)
	return err
}

func (d *Database) GetNode(id string) (*models.Node, error) {
	var node models.Node
	err := d.db.QueryRow(`
		SELECT id, name, region, cpu_cores, memory_gb, gpu_enabled, gpu_model, status,
			reputation_score, total_jobs_run, successful_jobs_run, failed_jobs,
			credits_earned, last_heartbeat, registered_at, current_job_id
		FROM nodes WHERE id = $1`, id).Scan(
		&node.ID, &node.Name, &node.Region, &node.CPUCores, &node.MemoryGB,
		&node.GPUEnabled, &node.GPUModel, &node.Status, &node.ReputationScore,
		&node.TotalJobsRun, &node.SuccessfulJobs, &node.FailedJobs,
		&node.CreditsEarned, &node.LastHeartbeat, &node.RegisteredAt, &node.CurrentJobID,
	)
	return &node, err
}

func (d *Database) GetAvailableNodes(requiredCPU, requiredMemory int, requiredGPU bool) ([]*models.Node, error) {
	query := `
		SELECT id, name, region, cpu_cores, memory_gb, gpu_enabled, gpu_model, status,
			reputation_score, total_jobs_run, successful_jobs_run, failed_jobs,
			credits_earned, last_heartbeat, registered_at, current_job_id
		FROM nodes
		WHERE status = $1
			AND cpu_cores >= $2
			AND memory_gb >= $3
			AND ($4 = FALSE OR gpu_enabled = TRUE)
		ORDER BY reputation_score DESC, total_jobs_run ASC
	`

	rows, err := d.db.Query(query, models.NodeStatusOnline, requiredCPU, requiredMemory, requiredGPU)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		err := rows.Scan(
			&node.ID, &node.Name, &node.Region, &node.CPUCores, &node.MemoryGB,
			&node.GPUEnabled, &node.GPUModel, &node.Status, &node.ReputationScore,
			&node.TotalJobsRun, &node.SuccessfulJobs, &node.FailedJobs,
			&node.CreditsEarned, &node.LastHeartbeat, &node.RegisteredAt, &node.CurrentJobID,
		)
		if err != nil {
			continue
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

func (d *Database) GetAllNodes() ([]*models.Node, error) {
	rows, err := d.db.Query(`
		SELECT id, name, region, cpu_cores, memory_gb, gpu_enabled, gpu_model, status,
			reputation_score, total_jobs_run, successful_jobs_run, failed_jobs,
			credits_earned, last_heartbeat, registered_at, current_job_id
		FROM nodes ORDER BY registered_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		err := rows.Scan(
			&node.ID, &node.Name, &node.Region, &node.CPUCores, &node.MemoryGB,
			&node.GPUEnabled, &node.GPUModel, &node.Status, &node.ReputationScore,
			&node.TotalJobsRun, &node.SuccessfulJobs, &node.FailedJobs,
			&node.CreditsEarned, &node.LastHeartbeat, &node.RegisteredAt, &node.CurrentJobID,
		)
		if err != nil {
			continue
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

func (d *Database) UpdateNodeHeartbeat(nodeID string, heartbeat *models.Heartbeat) error {
	_, err := d.db.Exec(`
		UPDATE nodes SET last_heartbeat = $1, status = $2 WHERE id = $3`,
		heartbeat.Timestamp, models.NodeStatusOnline, nodeID,
	)
	return err
}

func (d *Database) UpdateNodeStatus(nodeID string, status models.NodeStatus) error {
	_, err := d.db.Exec(`UPDATE nodes SET status = $1 WHERE id = $2`, status, nodeID)
	return err
}

func (d *Database) UpdateNodeReputation(nodeID string, delta float64) error {
	_, err := d.db.Exec(`
		UPDATE nodes SET reputation_score = GREATEST(0, reputation_score + $1) WHERE id = $2`,
		delta, nodeID,
	)
	return err
}

func (d *Database) IncrementNodeStats(nodeID string, success bool, creditsEarned int) error {
	if success {
		_, err := d.db.Exec(`
			UPDATE nodes SET
				total_jobs_run = total_jobs_run + 1,
				successful_jobs_run = successful_jobs_run + 1,
				credits_earned = credits_earned + $1
			WHERE id = $2`,
			creditsEarned, nodeID,
		)
		return err
	}

	_, err := d.db.Exec(`
		UPDATE nodes SET
			total_jobs_run = total_jobs_run + 1,
			failed_jobs = failed_jobs + 1
		WHERE id = $1`,
		nodeID,
	)
	return err
}

// JobExecution operations
func (d *Database) CreateJobExecution(execution *models.JobExecution) error {
	_, err := d.db.Exec(`
		INSERT INTO job_executions (id, job_id, node_id, status, started_at)
		VALUES ($1, $2, $3, $4, $5)`,
		execution.ID, execution.JobID, execution.NodeID, execution.Status, execution.StartedAt,
	)
	return err
}

func (d *Database) UpdateJobExecution(execution *models.JobExecution) error {
	_, err := d.db.Exec(`
		UPDATE job_executions
		SET status = $1, completed_at = $2, result = $3, result_hash = $4,
		    error_message = $5, logs = $6
		WHERE id = $7`,
		execution.Status, execution.CompletedAt, execution.Result, execution.ResultHash,
		execution.ErrorMessage, execution.Logs, execution.ID,
	)
	return err
}

func (d *Database) GetJobExecutions(jobID string) ([]*models.JobExecution, error) {
	rows, err := d.db.Query(`
		SELECT id, job_id, node_id, status, started_at, completed_at,
		       result, result_hash, error_message, logs
		FROM job_executions WHERE job_id = $1`,
		jobID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*models.JobExecution
	for rows.Next() {
		var exec models.JobExecution
		err := rows.Scan(
			&exec.ID, &exec.JobID, &exec.NodeID, &exec.Status, &exec.StartedAt,
			&exec.CompletedAt, &exec.Result, &exec.ResultHash, &exec.ErrorMessage, &exec.Logs,
		)
		if err != nil {
			continue
		}
		executions = append(executions, &exec)
	}

	return executions, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
