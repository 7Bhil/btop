package process

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID        int32
	Name       string
	CPUPercent float64
	MemPercent float64
	MemMB      float64
}

type Monitor struct {
	totalMemory uint64
}

func NewMonitor() (*Monitor, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des informations mémoire: %w", err)
	}

	return &Monitor{
		totalMemory: memInfo.Total,
	}, nil
}

func (m *Monitor) GetProcesses(limit int) ([]ProcessInfo, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des PIDs: %w", err)
	}

	var processes []ProcessInfo

	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		memPercent := float64(memInfo.RSS) / float64(m.totalMemory) * 100
		memMB := float64(memInfo.RSS) / 1024 / 1024

		processes = append(processes, ProcessInfo{
			PID:        pid,
			Name:       name,
			CPUPercent: cpuPercent,
			MemPercent: memPercent,
			MemMB:      memMB,
		})
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPUPercent > processes[j].CPUPercent
	})

	if limit > 0 && len(processes) > limit {
		processes = processes[:limit]
	}

	return processes, nil
}

func (m *Monitor) KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("processus introuvable: %w", err)
	}

	name, err := p.Name()
	if err != nil {
		name = "inconnu"
	}

	err = p.Kill()
	if err != nil {
		return fmt.Errorf("impossible de tuer le processus %d: %w", pid, err)
	}

	fmt.Printf("✅ Processus %s (PID %d) arrêté avec succès.\n", name, pid)
	return nil
}

func (m *Monitor) PrintProcesses(processes []ProcessInfo) {
	color.Cyan("%-6s %-15s %-8s %-8s %-10s", "PID", "NOM", "CPU%", "RAM%", "MEM(MB)")
	fmt.Println(strings.Repeat("-", 55))

	for _, proc := range processes {
		pidStr := fmt.Sprintf("%d", proc.PID)
		name := proc.Name
		cpuStr := fmt.Sprintf("%.1f", proc.CPUPercent)
		memStr := fmt.Sprintf("%.1f", proc.MemPercent)
		memMBStr := fmt.Sprintf("%.0f", proc.MemMB)

		if proc.CPUPercent > 70 {
			cpuStr = color.RedString(cpuStr)
		} else if proc.CPUPercent > 40 {
			cpuStr = color.YellowString(cpuStr)
		} else {
			cpuStr = color.GreenString(cpuStr)
		}

		fmt.Printf("%-6s %-15s %-8s %-8s %-10s\n", pidStr, name, cpuStr, memStr, memMBStr)
	}
}

func (m *Monitor) CheckThresholds(cpuThreshold, memThreshold float64, autoKill bool) ([]ProcessInfo, error) {
	processes, err := m.GetProcesses(50)
	if err != nil {
		return nil, err
	}

	var alertProcesses []ProcessInfo

	for _, proc := range processes {
		if proc.CPUPercent > cpuThreshold || proc.MemPercent > memThreshold {
			alertProcesses = append(alertProcesses, proc)

			color.Yellow("⚠  Utilisation élevée détectée")
			fmt.Printf("Processus : %s\n", proc.Name)
			fmt.Printf("PID : %d\n", proc.PID)
			fmt.Printf("CPU : %.1f %%\n", proc.CPUPercent)
			fmt.Printf("RAM : %.1f %%\n", proc.MemPercent)
			fmt.Println(strings.Repeat("-", 30))

			if autoKill {
				err := m.KillProcess(proc.PID)
				if err != nil {
					color.Red("Erreur lors du kill automatique: %v", err)
				}
			}
		}
	}

	return alertProcesses, nil
}
