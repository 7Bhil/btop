package cmd

import (
	"fmt"
	"time"

	"btop/internal/monitor"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var autoKill bool

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Surveille les processus en continu et affiche une alerte",
	Run: func(cmd *cobra.Command, args []string) {
		yellow := color.New(color.FgYellow).SprintFunc()
		for {
			procs, err := monitor.GetTopProcesses(0) // Limite 0 = tous les processus
			if err == nil {
				for _, p := range procs {
					if p.CPU > 80 || p.RAM > 80 {
						fmt.Printf("%s\n", yellow("⚠ Utilisation élevée détectée"))
						fmt.Printf("Processus : %s\n", p.Name)
						fmt.Printf("PID : %d\n", p.PID)
						if p.CPU > 80 {
							fmt.Printf("CPU : %.1f %%\n", p.CPU)
						} else {
							fmt.Printf("RAM : %.1f %%\n", p.RAM)
						}

						if autoKill {
							err := monitor.KillProcess(p.PID)
							if err == nil {
								fmt.Printf("Processus %s (PID %d) tué automatiquement.\n", p.Name, p.PID)
							} else {
								fmt.Printf("Échec de l'arrêt automatique: %v\n", err)
							}
						}
						fmt.Println("---")
					}
				}
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	},
}

func init() {
	watchCmd.Flags().BoolVar(&autoKill, "auto-kill", false, "Tuer automatiquement les processus dépassant le seuil")
	rootCmd.AddCommand(watchCmd)
}
