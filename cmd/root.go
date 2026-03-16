package cmd

import (
	"fmt"
	"os"

	"btop/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	limit      int
	sortBy     string
	filterName string
	filterUser string
)

var rootCmd = &cobra.Command{
	Use:   "btop",
	Short: "Un utilitaire CLI interactif pour surveiller les processus du système",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(
			tui.NewModel(limit, sortBy, filterName, filterUser),
			tea.WithAltScreen(),
		)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Erreur lors de l'exécution: %v", err)
			os.Exit(1)
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
	rootCmd.Flags().IntVar(&limit, "limit", 15, "nombre par défaut de processus affichés")
	rootCmd.Flags().StringVar(&sortBy, "sort", "cpu", "champ de tri (cpu, ram, mem, pid, name)")
	rootCmd.Flags().StringVar(&filterName, "name", "", "filtrer par nom de processus")
	rootCmd.Flags().StringVar(&filterUser, "user", "", "filtrer par nom d'utilisateur")
}
