package tui

import (
	"fmt"
	"strings"
	"time"

	"btop/internal/monitor"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type Model struct {
	processes  []monitor.ProcessInfo
	metrics    monitor.SystemMetrics
	limit      int
	sortBy     string
	filterName string
	filterUser string
	
	cursor     int
	width      int
	height     int
	
	message    string
	messageTimer int
}

func NewModel(limit int, sortBy, filterName, filterUser string) *Model {
	return &Model{
		limit:      limit,
		sortBy:     sortBy,
		filterName: filterName,
		filterUser: filterUser,
		cursor:     0,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		tickCmd(),
	)
}

func (m *Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		metrics, _ := monitor.GetSystemMetrics()
		procs, _ := monitor.GetTopProcesses(m.limit, m.sortBy, m.filterName, m.filterUser)
		return dataMsg{
			metrics: metrics,
			procs:   procs,
		}
	}
}

type dataMsg struct {
	metrics monitor.SystemMetrics
	procs   []monitor.ProcessInfo
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.processes)-1 && m.cursor < m.limit-1 {
				m.cursor++
			}
		case "left", "<":
			m.cycleSort(-1)
			return m, m.fetchData()
		case "right", ">":
			m.cycleSort(1)
			return m, m.fetchData()
		case "x", "enter":
			if len(m.processes) > 0 && m.cursor < len(m.processes) {
				pid := m.processes[m.cursor].PID
				err := monitor.KillProcess(pid)
				if err != nil {
					m.message = fmt.Sprintf("Erreur: %v", err)
				} else {
					m.message = fmt.Sprintf("Processus %d tué", pid)
				}
				m.messageTimer = 3 // Afficher le message pendant 3 ticks (6 sec)
			}
			return m, m.fetchData()
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Ajustement de la limite des processus affichés selon la hauteur
		newLimit := m.height - 8 // Hauteur totale moins le header/footer
		if newLimit > 0 {
			m.limit = newLimit
			// Fetch de nouvelles données si la limite grandit, mais safe to just refetch next tick
		}
		
	case dataMsg:
		m.processes = msg.procs
		m.metrics = msg.metrics
		// Ajustement du curseur s'il dépasse
		if m.cursor >= len(m.processes) && len(m.processes) > 0 {
			m.cursor = len(m.processes) - 1
		} else if len(m.processes) == 0 {
			m.cursor = 0
		}
		
	case tickMsg:
		if m.messageTimer > 0 {
			m.messageTimer--
			if m.messageTimer == 0 {
				m.message = ""
			}
		}
		return m, tea.Batch(m.fetchData(), tickCmd())
	}

	return m, nil
}

func (m *Model) cycleSort(dir int) {
	sorts := []string{"cpu", "ram", "mem", "pid", "name"}
	idx := 0
	for i, s := range sorts {
		if s == m.sortBy {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(sorts) - 1
	} else if idx >= len(sorts) {
		idx = 0
	}
	m.sortBy = sorts[idx]
}

// Styles
var (
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	footerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
)

func (m *Model) View() string {
	if len(m.processes) == 0 {
		return "Chargement..."
	}

	var b strings.Builder

	// Titre
	b.WriteString(titleStyle.Render("=== btop - Moniteur de Processus ===") + "\n")

	// Métriques Système
	ramUsagePct := float64(m.metrics.UsedRAM) / float64(m.metrics.TotalRAM) * 100
	b.WriteString(fmt.Sprintf("CPU Global : %s\n", formatBar(m.metrics.CPUPercent)))
	b.WriteString(fmt.Sprintf("RAM Globale: %s (%.1f GB / %.1f GB)\n\n", 
		formatBar(ramUsagePct), 
		float64(m.metrics.UsedRAM)/1024/1024/1024, 
		float64(m.metrics.TotalRAM)/1024/1024/1024))

	// En-tête Tableau
	sortInd := fmt.Sprintf("(Tri: %s)", strings.ToUpper(m.sortBy))
	header := fmt.Sprintf("%-8s %-12s %-20s %-8s %-8s %-10s %s\n", "PID", "USER", "NOM", "CPU%", "RAM%", "MEM(MB)", sortInd)
	b.WriteString(headerStyle.Render(header))

	// Lignes des Processus
	for i, p := range m.processes {
		cpuStr := fmt.Sprintf("%-8.1f", p.CPU)
		
		row := fmt.Sprintf("%-8d %-12s %-20s %s %-8.1f %-10.0f",
			p.PID,
			truncateString(p.User, 11),
			truncateString(p.Name, 19),
			cpuStr,
			p.RAM,
			p.MemMB)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(row) + "\n")
		} else {
			// Couleur du texte CPU selon l'utilisation si non sélectionné
			if p.CPU > 70 {
				row = strings.Replace(row, cpuStr, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(cpuStr), 1)
			} else if p.CPU > 40 {
				row = strings.Replace(row, cpuStr, lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(cpuStr), 1)
			} else {
				row = strings.Replace(row, cpuStr, lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render(cpuStr), 1)
			}
			b.WriteString(normalStyle.Render(row) + "\n")
		}
	}

	// Message de statut
	if m.message != "" {
		if strings.HasPrefix(m.message, "Erreur") {
			b.WriteString("\n" + errorStyle.Render(m.message))
		} else {
			b.WriteString("\n" + successStyle.Render(m.message))
		}
	}

	// Footer
	b.WriteString("\n" + footerStyle.Render("[↑/k: Haut] [↓/j: Bas] [x/Entrée: Tuer] [</>: Trier] [q: Quitter]"))

	return b.String()
}

func formatBar(percent float64) string {
	bars := int(percent / 5)
	if bars > 20 {
		bars = 20
	}
	
	barStr := fmt.Sprintf("[%s%s] %5.1f%%", strings.Repeat("|", bars), strings.Repeat(" ", 20-bars), percent)
	style := lipgloss.NewStyle()
	
	if percent > 80 {
		style = style.Foreground(lipgloss.Color("196")) // Rouge
	} else if percent > 50 {
		style = style.Foreground(lipgloss.Color("220")) // Jaune
	} else {
		style = style.Foreground(lipgloss.Color("46"))  // Vert
	}
	return style.Render(barStr)
}

func truncateString(str string, length int) string {
	if len(str) > length {
		return str[:length]
	}
	return str
}
