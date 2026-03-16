package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill [pid]",
	Short: "Tuer un processus processus avec son PID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pidInt, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			fmt.Println("Erreur: PID invalide")
			os.Exit(1)
		}

		pid := int32(pidInt)
		p, err := process.NewProcess(pid)
		if err != nil {
			fmt.Println("Erreur: PID inexistant ou permission refusée")
			os.Exit(1)
		}

		name, err := p.Name()
		if err != nil {
			name = "Inconnu"
		}

		err = p.Kill()
		if err != nil {
			fmt.Printf("Erreur lors de l'arrêt du processus %s (PID %d): %v\n", name, pid, err)
			os.Exit(1)
		}

		fmt.Printf("Processus %s (PID %d) arrêté avec succès.\n", name, pid)
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
}
