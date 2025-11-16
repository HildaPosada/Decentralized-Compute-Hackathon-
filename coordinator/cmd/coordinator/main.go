package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/HildaPosada/distributeai/coordinator/internal/api"
	"github.com/HildaPosada/distributeai/coordinator/internal/repository"
	"github.com/HildaPosada/distributeai/coordinator/internal/scheduler"
	"github.com/HildaPosada/distributeai/coordinator/internal/verification"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Configure logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.Info("Starting DistributeAI Coordinator...")

	// Get database connection string from environment
	dbURL := getEnv("DATABASE_URL", "postgres://distributeai:distributeai_dev@localhost:5432/distributeai?sslmode=disable")

	// Initialize database
	db, err := repository.NewDatabase(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Info("Database connected successfully")

	// Initialize verifier and scheduler
	verifier := verification.NewVerifier(db)
	sched := scheduler.NewScheduler(db, verifier)

	// Start scheduler in background
	go sched.Start()

	// Initialize API handler
	handler := api.NewHandler(db)

	// Setup Gin router
	router := gin.Default()

	// Enable CORS for dashboard
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", handler.HealthCheck)
	router.GET("/stats", handler.GetStats)

	// Job endpoints
	jobs := router.Group("/api/v1/jobs")
	{
		jobs.POST("", handler.SubmitJob)
		jobs.GET("", handler.ListJobs)
		jobs.GET("/:id", handler.GetJob)
		jobs.GET("/:id/executions", handler.GetJobExecutions)
	}

	// Node endpoints
	nodes := router.Group("/api/v1/nodes")
	{
		nodes.POST("/register", handler.RegisterNode)
		nodes.GET("", handler.ListNodes)
		nodes.GET("/:id", handler.GetNode)
		nodes.POST("/:id/heartbeat", handler.NodeHeartbeat)
		nodes.GET("/:nodeId/pending-jobs", handler.GetPendingJobs)
	}

	// Worker endpoints
	worker := router.Group("/api/v1/worker")
	{
		worker.POST("/result", handler.SubmitJobResult)
	}

	// Prometheus metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Get port from environment
	port := getEnv("COORDINATOR_PORT", "8080")

	log.Infof("Coordinator API listening on port %s", port)

	// Handle graceful shutdown
	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down coordinator...")
	sched.Stop()
	log.Info("Coordinator stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
