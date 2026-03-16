package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"btop/internal/monitor"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	interval   int
	limit      int
	sortBy     string
	filterName string
	filterUser string
)

var rootCmd = &cobra.Command{
	Use:   "btop",
	Short: "Un utilitaire CLI léger pour surveiller les processus du système",
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()

		for {
			printTable(red, yellow, green, cyan)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVar(&interval, "interval", 2, "intervalle de rafraîchissement en secondes")
	rootCmd.Flags().IntVar(&limit, "limit", 10, "nombre de processus affichés")
	rootCmd.Flags().StringVar(&sortBy, "sort", "cpu", "champ de tri (cpu, ram, pid, name, mem)")
	rootCmd.Flags().StringVar(&filterName, "name", "", "filtrer par nom de processus")
	rootCmd.Flags().StringVar(&filterUser, "user", "", "filtrer par nom d'utilisateur")
}

func printTable(red, yellow, green, cyan func(a ...interface{}) string) {
	metrics, _ := monitor.GetSystemMetrics()
	procs, err := monitor.GetTopProcesses(limit, sortBy, filterName, filterUser)
	if err != nil {
		fmt.Println("Erreur de récupération des processus:", err)
		return
	}

	// Clear screen reliably
	cmd := exec.Command("clear") // For Linux/macOS
	cmd.Stdout = os.Stdout
	cmd.Run()

	ramUsagePct := float64(metrics.UsedRAM) / float64(metrics.TotalRAM) * 100
	fmt.Printf("%s\n", cyan(fmt.Sprintf("=== btop - Métriques Système ===")))
	fmt.Printf("CPU Global : %s\n", formatBar(metrics.CPUPercent, red, yellow, green))
	fmt.Printf("RAM Globale: %s (%.1f GB / %.1f GB)\n", formatBar(ramUsagePct, red, yellow, green), float64(metrics.UsedRAM)/1024/1024/1024, float64(metrics.TotalRAM)/1024/1024/1024)
	fmt.Println()

	fmt.Printf("%-8s %-12s %-20s %-8s %-8s %-10s\n", "PID", "USER", "NOM", "CPU%", "RAM%", "MEM(MB)")
	for _, p := range procs {
		cpuStr := fmt.Sprintf("%-8.1f", p.CPU)
		if p.CPU > 70 {
			cpuStr = red(cpuStr)
		} else if p.CPU > 40 {
			cpuStr = yellow(cpuStr)
		} else {
			cpuStr = green(cpuStr)
		}

		fmt.Printf("%-8d %-12s %-20s %s %-8.1f %-10.0f\n",
			p.PID,
			truncateString(p.User, 11),
			truncateString(p.Name, 19),
			cpuStr,
			p.RAM,
			p.MemMB)
	}
}

func formatBar(percent float64, red, yellow, green func(a ...interface{}) string) string {
	bars := int(percent / 5)
	if bars > 20 {
		bars = 20
	}
	
	barStr := fmt.Sprintf("[%s%s] %5.1f%%", strings.Repeat("|", bars), strings.Repeat(" ", 20-bars), percent)
	if percent > 80 {
		return red(barStr)
	} else if percent > 50 {
		return yellow(barStr)
	}
	return green(barStr)
}

func truncateString(str string, length int) string {
	if len(str) > length {
		return str[:length]
	}
	return str
}
