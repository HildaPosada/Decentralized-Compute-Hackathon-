package monitor

import (
	"os"
	"runtime"
)

type SystemMonitor struct {
	cpuCores int
	memoryGB int
}

func NewSystemMonitor(cpuCores, memoryGB int) *SystemMonitor {
	return &SystemMonitor{
		cpuCores: cpuCores,
		memoryGB: memoryGB,
	}
}

// GetCPUUsage returns current CPU usage percentage (simplified)
func (m *SystemMonitor) GetCPUUsage() float64 {
	// Simplified - in production would use system metrics
	return float64(runtime.NumGoroutine()) / float64(m.cpuCores*10) * 100
}

// GetMemoryUsage returns current memory usage percentage
func (m *SystemMonitor) GetMemoryUsage() float64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	// Convert to GB and calculate percentage
	usedGB := float64(mem.Alloc) / 1024 / 1024 / 1024
	return (usedGB / float64(m.memoryGB)) * 100
}

// GetHostname returns the system hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
