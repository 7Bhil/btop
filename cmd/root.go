package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"btop/internal/monitor"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	interval int
	limit    int
)

var rootCmd = &cobra.Command{
	Use:   "btop",
	Short: "Un utilitaire CLI léger pour surveiller les processus du système",
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		for {
			printTable(red, yellow, green)
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
}

func printTable(red, yellow, green func(a ...interface{}) string) {
	procs, err := monitor.GetTopProcesses(limit)
	if err != nil {
		fmt.Println("Erreur de récupération des processus:", err)
		return
	}

	// Clear screen reliably
	cmd := exec.Command("clear") // For Linux/macOS
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Printf("%-8s %-20s %-8s %-8s %-10s\n", "PID", "NOM", "CPU%", "RAM%", "MEM(MB)")
	for _, p := range procs {
		cpuStr := fmt.Sprintf("%-8.1f", p.CPU)
		if p.CPU > 70 {
			cpuStr = red(cpuStr)
		} else if p.CPU > 40 {
			cpuStr = yellow(cpuStr)
		} else {
			cpuStr = green(cpuStr)
		}

		fmt.Printf("%-8d %-20s %s %-8.1f %-10.0f\n",
			p.PID,
			truncateString(p.Name, 19),
			cpuStr,
			p.RAM,
			p.MemMB)
	}
}

func truncateString(str string, length int) string {
	if len(str) > length {
		return str[:length]
	}
	return str
}
