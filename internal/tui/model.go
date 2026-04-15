package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"killerprocess/internal/monitor"
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

	searchInput textinput.Model
	isSearching bool

	detailedView   bool
	currentDetails monitor.ProcessDetails
	confirmKillPID int32

	cpuHistory []float64
	ramHistory []float64
}

func NewModel(limit int, sortBy, filterName, filterUser string) *Model {
	ti := textinput.New()
	ti.Placeholder = "Rechercher par nom (Entrée/Esc pour quitter)..."
	ti.Prompt = "🔍 "
	ti.CharLimit = 50
	ti.Width = 40

	if filterName != "" {
		ti.SetValue(filterName)
	}

	return &Model{
		limit:          limit,
		sortBy:         sortBy,
		filterName:     filterName,
		filterUser:     filterUser,
		cursor:         0,
		searchInput:    ti,
		isSearching:    false,
		confirmKillPID: -1,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		tickCmd(),
		textinput.Blink,
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
		if m.isSearching {
			switch msg.String() {
			case "esc", "enter":
				m.isSearching = false
				m.searchInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				
				newFilter := m.searchInput.Value()
				if newFilter != m.filterName {
					m.filterName = newFilter
					return m, tea.Batch(cmd, m.fetchData())
				}
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.processes)-1 && m.cursor < m.limit-1 {
				m.cursor++
			}
		case "esc":
			if m.detailedView {
				m.detailedView = false
				return m, nil
			}
			if m.confirmKillPID != -1 {
				m.confirmKillPID = -1
				m.message = "Erreur: Annulé."
				m.messageTimer = 2
				return m, nil
			}
		case "o", "y":
			if m.confirmKillPID != -1 {
				err := monitor.KillProcess(m.confirmKillPID)
				if err != nil {
					m.message = fmt.Sprintf("Erreur: %v", err)
				} else {
					m.message = fmt.Sprintf("Processus %d tué", m.confirmKillPID)
				}
				m.messageTimer = 3
				m.confirmKillPID = -1
				return m, m.fetchData()
			}
		case "n":
			if m.confirmKillPID != -1 {
				m.confirmKillPID = -1
				m.message = "Erreur: Annulé."
				m.messageTimer = 2
				return m, nil
			}
		case "left", "<":
			m.cycleSort(-1)
			return m, m.fetchData()
		case "right", ">":
			m.cycleSort(1)
			return m, m.fetchData()
		case "/", "f":
			m.isSearching = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case "k":
			if len(m.processes) > 0 && m.cursor < len(m.processes) && !m.detailedView {
				m.confirmKillPID = m.processes[m.cursor].PID
			}
			return m, nil
		case "enter":
			if m.confirmKillPID != -1 { return m, nil }
			if len(m.processes) > 0 && m.cursor < len(m.processes) && !m.detailedView {
				m.detailedView = true
				m.currentDetails = monitor.GetProcessDetails(m.processes[m.cursor].PID)
			}
			return m, nil
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Ajustement de la limite des processus affichés selon la hauteur
		newLimit := m.height - 16 // Hauteur totale moins le header/footer (beaucoup plus grand maintenant avec la V3)
		if newLimit > 0 {
			m.limit = newLimit
			// Fetch de nouvelles données si la limite grandit, mais safe to just refetch next tick
		}
		
	case dataMsg:
		m.processes = msg.procs
		m.metrics = msg.metrics

		m.cpuHistory = append(m.cpuHistory, m.metrics.CPUPercent)
		if len(m.cpuHistory) > 30 {
			m.cpuHistory = m.cpuHistory[1:]
		}
		
		ramPct := 0.0
		if m.metrics.TotalRAM > 0 {
			ramPct = float64(m.metrics.UsedRAM) / float64(m.metrics.TotalRAM) * 100
		}
		m.ramHistory = append(m.ramHistory, ramPct)
		if len(m.ramHistory) > 30 {
			m.ramHistory = m.ramHistory[1:]
		}

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

	if m.detailedView {
		b.WriteString(titleStyle.Render(fmt.Sprintf("=== KILLERPROCESS - Détails du PID %d ===", m.currentDetails.PID)) + "\n\n")
		b.WriteString(fmt.Sprintf("Nom/Exe : %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(m.currentDetails.Exe)))
		b.WriteString(fmt.Sprintf("Status  : %s\n", m.currentDetails.Status))
		b.WriteString(fmt.Sprintf("Threads : %d\n", m.currentDetails.Threads))
		b.WriteString(fmt.Sprintf("Conn.   : %d\n", m.currentDetails.Connections))
		cwd := m.currentDetails.Cwd
		if cwd == "" { cwd = "Inconnu (Permission refusée?)" }
		b.WriteString(fmt.Sprintf("Dossier : %s\n", cwd))
		cmdLine := m.currentDetails.Cmdline
		if cmdLine == "" { cmdLine = "Inconnu" }
		b.WriteString(fmt.Sprintf("\nCommande exacte :\n%s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(cmdLine)))
		b.WriteString("\n" + footerStyle.Render("[Esc: Retour à la liste] [q: Quitter]"))
		return b.String()
	}

	// Titre ASCII Art
	banner := `  _  ___ _ _               _____                                
 | |/ (_) | |             |  __ \                               
 | ' / _| | | ___ _ __    | |__) | __ ___   ___ ___  ___ ___    
 |  < | | | |/ _ \ '__|   |  ___/ '__/ _ \ / __/ _ \/ __/ __|   
 | . \| | | |  __/ |      | |   | | | (_) | (_|  __/\__ \__ \   
 |_|\_\_|_|_|\___|_|      |_|   |_|  \___/ \___\___||___/___/   `
 
	b.WriteString(titleStyle.Render(banner) + "\n\n")

	// Métriques Système
	ramUsagePct := float64(m.metrics.UsedRAM) / float64(m.metrics.TotalRAM) * 100
	
	cpuSparkline := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(generateSparkline(m.cpuHistory))
	ramSparkline := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(generateSparkline(m.ramHistory))

	b.WriteString(fmt.Sprintf("CPU Global : %s  [%-30s]\n", formatBar(m.metrics.CPUPercent), cpuSparkline))
	b.WriteString(fmt.Sprintf("RAM Globale: %s (%.1f GB / %.1f GB)  [%-30s]\n\n", 
		formatBar(ramUsagePct), 
		float64(m.metrics.UsedRAM)/1024/1024/1024, 
		float64(m.metrics.TotalRAM)/1024/1024/1024,
		ramSparkline))

	// Hardware Stats
	netStr := fmt.Sprintf("Net (RX/TX): %s / %s", lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(formatBytes(m.metrics.NetRxPerSec)+"/s"), lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(formatBytes(m.metrics.NetTxPerSec)+"/s"))
	diskStr := fmt.Sprintf("Disk (R/W): %s / %s", lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(formatBytes(m.metrics.DiskReadPerSec)+"/s"), lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(formatBytes(m.metrics.DiskWritePerSec)+"/s"))
	
	tempStr := "Temp : "
	for i, t := range m.metrics.Temperatures {
		if i >= 3 { break }
		color := "46"
		if t.Temperature > 60 { color = "220" }
		if t.Temperature > 80 { color = "196" }
		tempStr += lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(fmt.Sprintf("%.0f°C ", t.Temperature))
	}
	if len(m.metrics.Temperatures) == 0 {
		tempStr += "N/A"
	}

	b.WriteString(fmt.Sprintf("%-45s %-45s %s\n\n", netStr, diskStr, tempStr))

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
	if m.confirmKillPID != -1 {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render(
			fmt.Sprintf("⚠️  Tuer le PID %d ? [o/n]", m.confirmKillPID)))
	} else if m.message != "" {
		if strings.HasPrefix(m.message, "Erreur") {
			b.WriteString("\n" + errorStyle.Render(m.message))
		} else {
			b.WriteString("\n" + successStyle.Render(m.message))
		}
	}

	// Footer
	if m.isSearching {
		b.WriteString("\n" + footerStyle.Render(m.searchInput.View()))
	} else {
		b.WriteString("\n" + footerStyle.Render("[↑: Haut] [↓/j: Bas] [Enter: Détails] [k: Tuer] [</>: Trier] [/ ou f: Chercher] [q: Quitter]"))
	}

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

func generateSparkline(history []float64) string {
	sparks := []rune{'\u2800', '\u2840', '\u2844', '\u2846', '\u2847', '\u28c7', '\u28e7', '\u28f7', '\u28ff'}
	var b strings.Builder
	for _, val := range history {
		idx := int((val / 100.0) * 8.0)
		if idx < 0 { idx = 0 }
		if idx > 8 { idx = 8 }
		b.WriteRune(sparks[idx])
	}
	return b.String()
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
