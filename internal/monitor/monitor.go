package monitor

import (
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID     int32
	Name    string
	CPU     float64
	RAM     float32
	MemMB   float64
}

var procCache = make(map[int32]*process.Process)

func GetTopProcesses(limit int) ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procs []ProcessInfo
	currentPids := make(map[int32]bool)

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

	// Sort by CPU (descending)
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].CPU > procs[j].CPU
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
