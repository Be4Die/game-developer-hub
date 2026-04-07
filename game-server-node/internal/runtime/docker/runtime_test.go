package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestCalculateCPUPercent(t *testing.T) {
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected float64
	}{
		{
			name: "zero delta returns 0",
			stats: &container.StatsResponse{
				CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 100},
				PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 100},
			},
			expected: 0.0,
		},
		{
			name: "valid calculation (50% usage on 2 cores)",
			stats: &container.StatsResponse{
				PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 1000},
				CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 200}, SystemUsage: 1200, OnlineCPUs: 2},
			},
			// cpuDelta = 100, systemDelta = 200. (100/200) * 2 cores * 100 = 100.0%
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := calculateCPUPercent(tt.stats)
			if res != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, res)
			}
		})
	}
}
