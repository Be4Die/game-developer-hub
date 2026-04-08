//go:build !linux

package sysinfo

import (
	"testing"
)

func TestNewProvider_EmptyEthName(t *testing.T) {
	p := NewProvider("")

	if p.ethName != "eth0_mock" {
		t.Errorf("expected ethName 'eth0_mock', got '%s'", p.ethName)
	}
	if p.rng == nil {
		t.Errorf("expected non-nil rng")
	}
}

func TestNewProvider_WithEthName(t *testing.T) {
	p := NewProvider("eth1")

	if p.ethName != "eth1" {
		t.Errorf("expected ethName 'eth1', got '%s'", p.ethName)
	}
}

func TestStubProvider_GetMax(t *testing.T) {
	p := NewProvider("eth0")

	res, err := p.GetMax()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.CPUCores != 8 {
		t.Errorf("expected 8 CPU cores, got %d", res.CPUCores)
	}
	if res.TotalMemorySize != 16*1024*1024*1024 {
		t.Errorf("expected 16GB memory, got %d", res.TotalMemorySize)
	}
	if res.TotalDiskSpace != 500*1024*1024*1024 {
		t.Errorf("expected 500GB disk, got %d", res.TotalDiskSpace)
	}
	if res.NetworkBandwidth != 1000*125000 {
		t.Errorf("expected network bandwidth %d, got %d", 1000*125000, res.NetworkBandwidth)
	}
}

func TestStubProvider_GetUsage(t *testing.T) {
	p := NewProvider("eth0")

	res, err := p.GetUsage()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CPU should be between 30 and 50
	if res.CPU < 30.0 || res.CPU > 50.0 {
		t.Errorf("expected CPU between 30-50, got %f", res.CPU)
	}

	// Memory should be 6, 7, or 8 GB (in bytes)
	expectedMem := []uint64{
		6 * 1024 * 1024 * 1024,
		7 * 1024 * 1024 * 1024,
		8 * 1024 * 1024 * 1024,
	}
	memValid := false
	for _, m := range expectedMem {
		if res.Memory == m {
			memValid = true
			break
		}
	}
	if !memValid {
		t.Errorf("expected memory to be one of %v, got %d", expectedMem, res.Memory)
	}

	// Disk should be fixed at 250 GB
	if res.Disk != 250*1024*1024*1024 {
		t.Errorf("expected disk 250GB, got %d", res.Disk)
	}

	// Network should be between 0 and 5000*1024
	if res.Network > 5000*1024 {
		t.Errorf("expected network <= %d, got %d", 5000*1024, res.Network)
	}
}
