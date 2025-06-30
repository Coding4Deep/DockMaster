package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// getCPUMetrics gets CPU usage metrics
func getCPUMetrics() (*CPUMetrics, error) {
	// Read /proc/stat for CPU info
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read CPU stats")
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 || fields[0] != "cpu" {
		return nil, fmt.Errorf("invalid CPU stats format")
	}

	// Parse CPU times
	user, _ := strconv.ParseFloat(fields[1], 64)
	nice, _ := strconv.ParseFloat(fields[2], 64)
	system, _ := strconv.ParseFloat(fields[3], 64)
	idle, _ := strconv.ParseFloat(fields[4], 64)
	iowait, _ := strconv.ParseFloat(fields[5], 64)
	irq, _ := strconv.ParseFloat(fields[6], 64)
	softirq, _ := strconv.ParseFloat(fields[7], 64)

	total := user + nice + system + idle + iowait + irq + softirq
	usage := ((total - idle) / total) * 100

	// Get number of CPU cores
	cores := getCPUCores()

	return &CPUMetrics{
		Usage:      usage,
		UserTime:   user,
		SystemTime: system,
		IdleTime:   idle,
		Cores:      cores,
	}, nil
}

// getMemoryMetrics gets memory usage metrics
func getMemoryMetrics() (*MemoryMetrics, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memInfo := make(map[string]int64)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			value, err := strconv.ParseInt(fields[1], 10, 64)
			if err == nil {
				memInfo[key] = value * 1024 // Convert from KB to bytes
			}
		}
	}

	total := memInfo["MemTotal"]
	free := memInfo["MemFree"]
	available := memInfo["MemAvailable"]
	buffers := memInfo["Buffers"]
	cached := memInfo["Cached"]
	used := total - free

	var usage float64
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return &MemoryMetrics{
		Total:     total,
		Used:      used,
		Free:      free,
		Available: available,
		Usage:     usage,
		Buffers:   buffers,
		Cached:    cached,
	}, nil
}

// getDiskMetrics gets disk usage metrics for root filesystem
func getDiskMetrics() (*DiskMetrics, error) {
	// Use df command to get disk usage
	cmd := exec.Command("df", "-B1", "/")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return nil, fmt.Errorf("invalid df output format")
	}

	total, _ := strconv.ParseInt(fields[1], 10, 64)
	used, _ := strconv.ParseInt(fields[2], 10, 64)
	free, _ := strconv.ParseInt(fields[3], 10, 64)

	var usage float64
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	// Get disk I/O stats from /proc/diskstats
	readOps, writeOps, readBytes, writeBytes := getDiskIOStats()

	return &DiskMetrics{
		Total:      total,
		Used:       used,
		Free:       free,
		Usage:      usage,
		ReadOps:    readOps,
		WriteOps:   writeOps,
		ReadBytes:  readBytes,
		WriteBytes: writeBytes,
	}, nil
}

// getNetworkMetrics gets network usage metrics
func getNetworkMetrics() (*NetworkMetrics, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var totalRx, totalTx, totalRxPackets, totalTxPackets int64

	scanner := bufio.NewScanner(file)
	// Skip header lines
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 17 {
			// Skip loopback interface
			if strings.HasPrefix(fields[0], "lo:") {
				continue
			}

			rxBytes, _ := strconv.ParseInt(fields[1], 10, 64)
			rxPackets, _ := strconv.ParseInt(fields[2], 10, 64)
			txBytes, _ := strconv.ParseInt(fields[9], 10, 64)
			txPackets, _ := strconv.ParseInt(fields[10], 10, 64)

			totalRx += rxBytes
			totalTx += txBytes
			totalRxPackets += rxPackets
			totalTxPackets += txPackets
		}
	}

	return &NetworkMetrics{
		BytesReceived:   totalRx,
		BytesSent:       totalTx,
		PacketsReceived: totalRxPackets,
		PacketsSent:     totalTxPackets,
	}, nil
}

// getLoadMetrics gets system load averages
func getLoadMetrics() (*LoadMetrics, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read load average")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid load average format")
	}

	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)

	return &LoadMetrics{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}, nil
}

// getUptime gets system uptime in seconds
func getUptime() (int64, error) {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, fmt.Errorf("failed to read uptime")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}

	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	return int64(uptime), nil
}

// getCPUCores gets the number of CPU cores
func getCPUCores() int {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 1
	}
	defer file.Close()

	cores := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processor") {
			cores++
		}
	}

	if cores == 0 {
		return 1
	}
	return cores
}

// getDiskIOStats gets disk I/O statistics
func getDiskIOStats() (readOps, writeOps, readBytes, writeBytes int64) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return 0, 0, 0, 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 14 {
			// Skip loop devices and ram devices
			deviceName := fields[2]
			if strings.HasPrefix(deviceName, "loop") || strings.HasPrefix(deviceName, "ram") {
				continue
			}

			// Only consider main disk devices (sda, nvme0n1, etc.)
			if strings.Contains(deviceName, "sda") || strings.Contains(deviceName, "nvme") || strings.Contains(deviceName, "vda") {
				rOps, _ := strconv.ParseInt(fields[3], 10, 64)
				rBytes, _ := strconv.ParseInt(fields[5], 10, 64)
				wOps, _ := strconv.ParseInt(fields[7], 10, 64)
				wBytes, _ := strconv.ParseInt(fields[9], 10, 64)

				readOps += rOps
				writeOps += wOps
				readBytes += rBytes * 512  // sectors to bytes
				writeBytes += wBytes * 512 // sectors to bytes
			}
		}
	}

	return readOps, writeOps, readBytes, writeBytes
}
