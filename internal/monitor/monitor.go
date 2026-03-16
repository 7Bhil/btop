package monitor

import (
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID     int32
	Name    string
	User    string
	CPU     float64
	RAM     float32
	MemMB   float64
}

type SystemMetrics struct {
	CPUPercent float64
	TotalRAM   uint64
	UsedRAM    uint64
}

var procCache = make(map[int32]*process.Process)

func GetSystemMetrics() (SystemMetrics, error) {
	var metrics SystemMetrics
	
	cpuPercents, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercents) > 0 {
		metrics.CPUPercent = cpuPercents[0]
	}

	vmStat, err := mem.VirtualMemory()
	if err == nil {
		metrics.TotalRAM = vmStat.Total
		metrics.UsedRAM = vmStat.Used
	}

	return metrics, nil
}

func GetTopProcesses(limit int, sortBy string, filterName string, filterUser string) ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procs []ProcessInfo
	currentPids := make(map[int32]bool)
	filterNameVal := strings.ToLower(filterName)
	filterUserVal := strings.ToLower(filterUser)

	for _, p := range processes {
		pid := p.Pid
		currentPids[pid] = true

		if _, exists := procCache[pid]; !exists {
			procCache[pid] = p
			// Initial call to set the baseline for CPU calculation
			p.CPUPercent()
		}

		cachedProc := procCache[pid]

		name, err := cachedProc.Name()
		if err != nil || len(strings.TrimSpace(name)) == 0 {
			name = "unknown"
		}

		// Apply name filter early
		if filterNameVal != "" && !strings.Contains(strings.ToLower(name), filterNameVal) {
			continue
		}

		user, err := cachedProc.Username()
		if err != nil {
			user = "unknown"
		}
		
		// Apply user filter early
		if filterUserVal != "" && !strings.Contains(strings.ToLower(user), filterUserVal) {
			continue
		}

		cpu, err := cachedProc.CPUPercent()
		if err != nil {
			continue
		}

		mem, err := cachedProc.MemoryPercent()
		if err != nil {
			continue
		}

		memInfo, err := cachedProc.MemoryInfo()
		var rssMB float64 = 0
		if err == nil && memInfo != nil {
			rssMB = float64(memInfo.RSS) / 1024 / 1024
		}

		procs = append(procs, ProcessInfo{
			PID:   pid,
			Name:  name,
			User:  user,
			CPU:   cpu,
			RAM:   mem,
			MemMB: rssMB,
		})
	}

	// Clean up dead processes from cache
	for pid := range procCache {
		if !currentPids[pid] {
			delete(procCache, pid)
		}
	}

	// Sort logic
	sort.Slice(procs, func(i, j int) bool {
		switch strings.ToLower(sortBy) {
		case "ram":
			return procs[i].RAM > procs[j].RAM
		case "mem":
			return procs[i].MemMB > procs[j].MemMB
		case "pid":
			return procs[i].PID < procs[j].PID
		case "name":
			return procs[i].Name < procs[j].Name
		default: // "cpu"
			return procs[i].CPU > procs[j].CPU
		}
	})

	if limit > 0 && len(procs) > limit {
		procs = procs[:limit]
	}

	return procs, nil
}

func KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
