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
		{
			name: "negative system delta returns 0",
			stats: &container.StatsResponse{
				CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 200}, SystemUsage: 500},
				PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 1000},
			},
			expected: 0.0,
		},
		{
			name: "zero online CPUs defaults to 1",
			stats: &container.StatsResponse{
				PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 1000},
				CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 150}, SystemUsage: 1100, OnlineCPUs: 0},
			},
			// (50/100) * 1 * 100 = 50.0%
			expected: 50.0,
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

func TestCalculateNetworkBytes(t *testing.T) {
	tests := []struct {
		name     string
		stats    *container.StatsResponse
		expected uint64
	}{
		{
			name:     "empty networks returns 0",
			stats:    &container.StatsResponse{},
			expected: 0,
		},
		{
			name: "single interface",
			stats: &container.StatsResponse{
				Networks: map[string]container.NetworkStats{
					"eth0": {RxBytes: 1000, TxBytes: 500},
				},
			},
			expected: 1500,
		},
		{
			name: "multiple interfaces",
			stats: &container.StatsResponse{
				Networks: map[string]container.NetworkStats{
					"eth0": {RxBytes: 2000, TxBytes: 1000},
					"lo":   {RxBytes: 500, TxBytes: 500},
				},
			},
			expected: 4000,
		},
		{
			name: "zero bytes",
			stats: &container.StatsResponse{
				Networks: map[string]container.NetworkStats{
					"eth0": {RxBytes: 0, TxBytes: 0},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := calculateNetworkBytes(tt.stats)
			if res != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, res)
			}
		})
	}
}
