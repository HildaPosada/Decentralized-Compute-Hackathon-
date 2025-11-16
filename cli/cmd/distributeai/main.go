package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var coordinatorURL string

func main() {
	rootCmd := &cobra.Command{
		Use:   "distributeai",
		Short: "DistributeAI CLI - Decentralized Compute Network",
		Long:  `Command-line interface for interacting with the DistributeAI decentralized compute platform.`,
	}

	rootCmd.PersistentFlags().StringVar(&coordinatorURL, "coordinator", "http://localhost:8080", "Coordinator API URL")

	rootCmd.AddCommand(
		submitJobCmd(),
		listJobsCmd(),
		getJobCmd(),
		listNodesCmd(),
		statsCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func submitJobCmd() *cobra.Command {
	var (
		name        string
		description string
		dockerImage string
		command     []string
		cpu         int
		memory      int
	)

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a new job",
		RunE: func(cmd *cobra.Command, args []string) error {
			job := map[string]interface{}{
				"name":            name,
				"description":     description,
				"docker_image":    dockerImage,
				"command":         command,
				"required_cpu":    cpu,
				"required_memory": memory,
			}

			data, _ := json.Marshal(job)
			resp, err := http.Post(
				coordinatorURL+"/api/v1/jobs",
				"application/json",
				bytes.NewBuffer(data),
			)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to submit job: %s - %s", resp.Status, string(body))
			}

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			fmt.Printf("✅ Job submitted successfully!\n")
			fmt.Printf("   ID: %s\n", result["id"])
			fmt.Printf("   Name: %s\n", result["name"])
			fmt.Printf("   Status: %s\n", result["status"])
			fmt.Printf("\nMonitor progress with: distributeai get %s\n", result["id"])

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Job name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Job description")
	cmd.Flags().StringVar(&dockerImage, "image", "", "Docker image (required)")
	cmd.Flags().StringArrayVar(&command, "cmd", []string{}, "Command to run (can specify multiple times)")
	cmd.Flags().IntVar(&cpu, "cpu", 1, "Required CPU cores")
	cmd.Flags().IntVar(&memory, "memory", 1, "Required memory (GB)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("image")

	return cmd
}

func listJobsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get(coordinatorURL + "/api/v1/jobs")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var result struct {
				Jobs  []map[string]interface{} `json:"jobs"`
				Count int                      `json:"count"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			fmt.Printf("📋 Total Jobs: %d\n\n", result.Count)

			if result.Count == 0 {
				fmt.Println("No jobs found.")
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Status", "Submitted"})
			table.SetBorder(false)

			for _, job := range result.Jobs {
				id := truncate(fmt.Sprintf("%v", job["id"]), 20)
				name := truncate(fmt.Sprintf("%v", job["name"]), 30)
				status := fmt.Sprintf("%v", job["status"])

				submittedAt, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", job["submitted_at"]))
				submitted := submittedAt.Format("Jan 02 15:04")

				table.Append([]string{id, name, status, submitted})
			}

			table.Render()
			return nil
		},
	}
}

func getJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [job-id]",
		Short: "Get job details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := args[0]

			resp, err := http.Get(coordinatorURL + "/api/v1/jobs/" + jobID)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("job not found")
			}

			var job map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
				return err
			}

			// Print job details
			fmt.Printf("📦 Job Details\n")
			fmt.Printf("   ID:          %s\n", job["id"])
			fmt.Printf("   Name:        %s\n", job["name"])
			fmt.Printf("   Description: %s\n", job["description"])
			fmt.Printf("   Status:      %s\n", job["status"])
			fmt.Printf("   Image:       %s\n", job["docker_image"])
			fmt.Printf("   Submitted:   %s\n", job["submitted_at"])

			if job["completed_at"] != nil {
				fmt.Printf("   Completed:   %s\n", job["completed_at"])
			}

			if job["result"] != nil && job["result"] != "" {
				fmt.Printf("\n📤 Result:\n%s\n", job["result"])
			}

			if job["error_message"] != nil && job["error_message"] != "" {
				fmt.Printf("\n❌ Error:\n%s\n", job["error_message"])
			}

			// Get executions
			resp2, err := http.Get(coordinatorURL + "/api/v1/jobs/" + jobID + "/executions")
			if err == nil {
				var execResult struct {
					Executions []map[string]interface{} `json:"executions"`
				}

				if err := json.NewDecoder(resp2.Body).Decode(&execResult); err == nil {
					fmt.Printf("\n🔄 Executions: %d\n", len(execResult.Executions))

					if len(execResult.Executions) > 0 {
						table := tablewriter.NewWriter(os.Stdout)
						table.SetHeader([]string{"Node", "Status", "Result Hash"})
						table.SetBorder(false)

						for _, exec := range execResult.Executions {
							nodeID := truncate(fmt.Sprintf("%v", exec["node_id"]), 20)
							status := fmt.Sprintf("%v", exec["status"])
							hash := truncate(fmt.Sprintf("%v", exec["result_hash"]), 16)

							table.Append([]string{nodeID, status, hash})
						}

						table.Render()
					}
				}
				resp2.Body.Close()
			}

			return nil
		},
	}
}

func listNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "nodes",
		Short: "List all worker nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get(coordinatorURL + "/api/v1/nodes")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var result struct {
				Nodes []map[string]interface{} `json:"nodes"`
				Count int                      `json:"count"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			fmt.Printf("🖥️  Total Nodes: %d\n\n", result.Count)

			if result.Count == 0 {
				fmt.Println("No nodes registered.")
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Status", "CPU", "Memory", "Reputation", "Jobs"})
			table.SetBorder(false)

			for _, node := range result.Nodes {
				id := truncate(fmt.Sprintf("%v", node["id"]), 20)
				name := truncate(fmt.Sprintf("%v", node["name"]), 25)
				status := fmt.Sprintf("%v", node["status"])
				cpu := fmt.Sprintf("%v cores", node["cpu_cores"])
				memory := fmt.Sprintf("%v GB", node["memory_gb"])
				reputation := fmt.Sprintf("%.1f", node["reputation_score"])
				jobs := fmt.Sprintf("%v", node["total_jobs_run"])

				table.Append([]string{id, name, status, cpu, memory, reputation, jobs})
			}

			table.Render()
			return nil
		},
	}
}

func statsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show system statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get(coordinatorURL + "/stats")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var stats map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
				return err
			}

			nodes := stats["nodes"].(map[string]interface{})
			resources := stats["resources"].(map[string]interface{})
			jobs := stats["jobs"].(map[string]interface{})

			fmt.Println("📊 DistributeAI Statistics")
			fmt.Println()
			fmt.Printf("🖥️  Nodes:\n")
			fmt.Printf("   Total:  %v\n", nodes["total"])
			fmt.Printf("   Online: %v\n", nodes["online"])
			fmt.Printf("   Busy:   %v\n", nodes["busy"])
			fmt.Println()
			fmt.Printf("⚡ Resources:\n")
			fmt.Printf("   CPU Cores: %v\n", resources["total_cpu_cores"])
			fmt.Printf("   Memory:    %v GB\n", resources["total_memory_gb"])
			fmt.Println()
			fmt.Printf("📦 Jobs:\n")
			fmt.Printf("   Total:     %v\n", jobs["total"])
			fmt.Printf("   Completed: %v\n", jobs["completed"])
			fmt.Printf("   Running:   %v\n", jobs["running"])
			fmt.Printf("   Failed:    %v\n", jobs["failed"])

			return nil
		},
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
