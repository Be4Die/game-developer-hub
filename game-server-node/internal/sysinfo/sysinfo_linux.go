//go:build linux

package sysinfo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"golang.org/x/sys/unix"
)

type LinuxProvider struct {
	mu           sync.Mutex
	ethName      string
	lastCPUIdle  uint64
	lastCPUTotal uint64
	lastNetBytes uint64
	lastCheck    time.Time
}

func NewProvider(ethName string) *LinuxProvider {
	if ethName == "" {
		ethName = findDefaultInterface()
	}

	p := &LinuxProvider{
		ethName: ethName,
	}

	// Инициализируем базовые счетчики
	p.lastCPUIdle, p.lastCPUTotal = p.getCPUCounters()
	p.lastNetBytes = p.getNetBytes()
	p.lastCheck = time.Now()

	return p
}

func (p *LinuxProvider) GetMax() (domain.ResourcesMax, error) {
	cpu, _ := p.getMaxCPU()
	ram, _ := p.getMaxRAM()
	disk, _ := p.getMaxDisk()
	net, _ := p.getMaxNet()

	return domain.ResourcesMax{
		CPUCores:         cpu,
		TotalMemorySize:  ram,
		TotalDiskSpace:   disk,
		NetworkBandwidth: net,
	}, nil
}

func (p *LinuxProvider) GetUsage() (domain.ResourcesUsage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var usage domain.ResourcesUsage

	memTotal, memAvail := p.getMemoryUsageBase()
	if memTotal > 0 && memTotal >= memAvail {
		usage.Memory = (memTotal - memAvail) * 1024
	}

	// 2. Disk
	var stat unix.Statfs_t
	if err := unix.Statfs("/host/root", &stat); err == nil {
		usage.Disk = (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)
	}

	now := time.Now()
	elapsed := now.Sub(p.lastCheck).Seconds()

	cpuIdle, cpuTotal := p.getCPUCounters()
	idleDiff := float64(cpuIdle) - float64(p.lastCPUIdle)
	totalDiff := float64(cpuTotal) - float64(p.lastCPUTotal)

	if totalDiff > 0 {
		usage.CPU = (1.0 - (idleDiff / totalDiff)) * 100.0
	}

	netBytes := p.getNetBytes()
	byteDiff := float64(netBytes) - float64(p.lastNetBytes)
	if elapsed > 0 && byteDiff > 0 {
		usage.Network = uint64(byteDiff / elapsed)
	}

	p.lastCPUIdle = cpuIdle
	p.lastCPUTotal = cpuTotal
	p.lastNetBytes = netBytes
	p.lastCheck = now

	return usage, nil
}

func (p *LinuxProvider) getMaxRAM() (uint64, error) {
	data, err := os.ReadFile("/host/proc/meminfo")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, _ := strconv.ParseUint(fields[1], 10, 64)
				return val * 1024, nil
			}
		}
	}
	return 0, fmt.Errorf("ram property not found")
}

func (p *LinuxProvider) getMaxDisk() (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs("/host/root", &stat); err == nil {
		return stat.Blocks * uint64(stat.Bsize), nil
	} else {
		return 0, err
	}
}

func (p *LinuxProvider) getMaxCPU() (uint32, error) {
	data, err := os.ReadFile("/host/proc/cpuinfo")
	if err != nil {
		return 0, err
	}
	return uint32(strings.Count(string(data), "processor\t:")), nil
}

func (p *LinuxProvider) getMaxNet() (uint64, error) {
	if p.ethName == "" {
		return 0, fmt.Errorf("network interface not configured")
	}

	data, err := os.ReadFile("/host/sys/class/net/" + p.ethName + "/speed")
	if err == nil {
		speed, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if speed > 0 {
			return speed * 125000, nil // Конвертация Mbits в Bytes
		}
	}
	return 0, fmt.Errorf("net info not found for interface: %s", p.ethName)
}

func findDefaultInterface() string {
	entries, err := os.ReadDir("/host/sys/class/net")
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.Name() != "lo" {
			return e.Name()
		}
	}
	return ""
}

func (p *LinuxProvider) getMemoryUsageBase() (total, avail uint64) {
	data, err := os.ReadFile("/host/proc/meminfo")
	if err != nil {
		return
	}
	var free uint64
	var hasAvail bool
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				total, _ = strconv.ParseUint(f[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				avail, _ = strconv.ParseUint(f[1], 10, 64)
				hasAvail = true
			}
		} else if strings.HasPrefix(line, "MemFree:") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				free, _ = strconv.ParseUint(f[1], 10, 64)
			}
		}
	}
	if !hasAvail {
		avail = free // Фолбэк на MemFree
	}
	return
}

func (p *LinuxProvider) getCPUCounters() (idle, total uint64) {
	data, err := os.ReadFile("/host/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "cpu ") {
		fields := strings.Fields(lines[0])[1:]
		for i, f := range fields {
			val, _ := strconv.ParseUint(f, 10, 64)
			total += val
			if i == 3 || i == 4 { // idle и iowait
				idle += val
			}
		}
	}
	return
}

func (p *LinuxProvider) getNetBytes() uint64 {
	if p.ethName == "" {
		return 0
	}

	rxPath := "/host/sys/class/net/" + p.ethName + "/statistics/rx_bytes"
	txPath := "/host/sys/class/net/" + p.ethName + "/statistics/tx_bytes"

	rxData, _ := os.ReadFile(rxPath)
	txData, _ := os.ReadFile(txPath)

	rx, _ := strconv.ParseUint(strings.TrimSpace(string(rxData)), 10, 64)
	tx, _ := strconv.ParseUint(strings.TrimSpace(string(txData)), 10, 64)

	return rx + tx
}
