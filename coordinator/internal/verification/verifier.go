package verification

import (
	"fmt"
	"github.com/HildaPosada/distributeai/coordinator/internal/models"
	"github.com/HildaPosada/distributeai/coordinator/internal/repository"
	log "github.com/sirupsen/logrus"
)

// Verifier handles k-of-n verification for job results
type Verifier struct {
	db *repository.Database
}

func NewVerifier(db *repository.Database) *Verifier {
	return &Verifier{db: db}
}

// VerifyJob performs k-of-n verification on completed job executions
// Returns true if consensus is reached, and the consensus result
func (v *Verifier) VerifyJob(jobID string) (*models.VerificationResult, error) {
	// Get the job to know required consensus
	job, err := v.db.GetJob(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Get all executions for this job
	executions, err := v.db.GetJobExecutions(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	// Filter only completed executions
	var completedExecutions []*models.JobExecution
	for _, exec := range executions {
		if exec.Status == models.JobStatusCompleted && exec.ResultHash != "" {
			completedExecutions = append(completedExecutions, exec)
		}
	}

	log.Infof("Job %s: %d/%d executions completed", jobID, len(completedExecutions), job.Redundancy)

	// Check if we have enough completed executions
	if len(completedExecutions) < job.Consensus {
		return &models.VerificationResult{
			JobID:            jobID,
			TotalExecutions:  len(completedExecutions),
			ConsensusReached: false,
		}, nil
	}

	// Count result hashes
	resultCounts := make(map[string]int)
	resultData := make(map[string]string) // hash -> actual result
	hashToNodes := make(map[string][]string) // hash -> node IDs

	for _, exec := range completedExecutions {
		resultCounts[exec.ResultHash]++
		resultData[exec.ResultHash] = exec.Result
		hashToNodes[exec.ResultHash] = append(hashToNodes[exec.ResultHash], exec.NodeID)
	}

	// Find the result with most votes
	var consensusHash string
	maxVotes := 0

	for hash, count := range resultCounts {
		if count > maxVotes {
			maxVotes = count
			consensusHash = hash
		}
	}

	// Check if consensus threshold is met
	consensusReached := maxVotes >= job.Consensus

	// Identify agreeing and disagreeing nodes
	var agreementNodes, disagreementNodes []string

	for hash, nodes := range hashToNodes {
		if hash == consensusHash {
			agreementNodes = append(agreementNodes, nodes...)
		} else {
			disagreementNodes = append(disagreementNodes, nodes...)
		}
	}

	result := &models.VerificationResult{
		JobID:             jobID,
		TotalExecutions:   len(completedExecutions),
		ResultCounts:      resultCounts,
		ConsensusResult:   resultData[consensusHash],
		ConsensusReached:  consensusReached,
		AgreementNodes:    agreementNodes,
		DisagreementNodes: disagreementNodes,
	}

	// Update node reputations based on agreement/disagreement
	if consensusReached {
		v.updateNodeReputations(agreementNodes, disagreementNodes)
	}

	return result, nil
}

// updateNodeReputations adjusts reputation scores based on verification results
func (v *Verifier) updateNodeReputations(agreementNodes, disagreementNodes []string) {
	// Reward nodes that agreed with consensus
	for _, nodeID := range agreementNodes {
		if err := v.db.UpdateNodeReputation(nodeID, 5.0); err != nil {
			log.Warnf("Failed to update reputation for node %s: %v", nodeID, err)
		}
		if err := v.db.IncrementNodeStats(nodeID, true, 1); err != nil {
			log.Warnf("Failed to update stats for node %s: %v", nodeID, err)
		}
	}

	// Penalize nodes that disagreed
	for _, nodeID := range disagreementNodes {
		if err := v.db.UpdateNodeReputation(nodeID, -10.0); err != nil {
			log.Warnf("Failed to update reputation for node %s: %v", nodeID, err)
		}
		if err := v.db.IncrementNodeStats(nodeID, false, 0); err != nil {
			log.Warnf("Failed to update stats for node %s: %v", nodeID, err)
		}
	}

	log.Infof("Updated reputations: %d rewarded, %d penalized",
		len(agreementNodes), len(disagreementNodes))
}

// CheckAndFinalizeJob checks if a job is ready for verification and finalizes it
func (v *Verifier) CheckAndFinalizeJob(jobID string) error {
	result, err := v.VerifyJob(jobID)
	if err != nil {
		return err
	}

	if !result.ConsensusReached {
		log.Infof("Job %s: Consensus not yet reached (%d/%d executions)",
			jobID, result.TotalExecutions, len(result.ResultCounts))
		return nil
	}

	// Finalize the job with the consensus result
	log.Infof("Job %s: Consensus reached! %d nodes agreed", jobID, len(result.AgreementNodes))

	err = v.db.UpdateJobStatus(jobID, models.JobStatusCompleted, result.ConsensusResult, "")
	if err != nil {
		return fmt.Errorf("failed to finalize job: %w", err)
	}

	return nil
}
