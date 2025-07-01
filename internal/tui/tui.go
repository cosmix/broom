package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

// ViewState represents the current view in the TUI
type ViewState int

const (
	WelcomeView ViewState = iota
	SelectionView
	PreviewView
	ProgressView
	ResultsView
)

// CleanerItem represents a cleaner in the selection list
type CleanerItem struct {
	ID          string
	Name        string
	Desc        string
	Category    string
	Selected    bool
	NeedsConfirm bool
}

func (c CleanerItem) Title() string       { return c.Name }
func (c CleanerItem) Description() string { return c.Desc }
func (c CleanerItem) FilterValue() string { return c.Name + " " + c.Desc }

// ProgressItem represents progress for a specific cleaner
type ProgressItem struct {
	Name       string
	Status     string
	Progress   float64
	SpaceFreed uint64
	Duration   time.Duration
	Error      error
}

// Model represents the TUI application state
type Model struct {
	state           ViewState
	cleanerList     list.Model
	selectedCleaners []CleanerItem
	progressItems   []ProgressItem
	progressSpinner spinner.Model
	progressBar     progress.Model
	currentCleaner  int
	totalProgress   float64
	startSpace      uint64
	endSpace        uint64
	err             error
	quitting        bool
	width           int
	height          int
}

// Styling
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true).
			Padding(1, 2)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(1, 2).
			Margin(1, 0)

	selectedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Padding(1, 2).
			Margin(1, 0)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Align(lipgloss.Center).
			Padding(1, 0)
)

// Messages for the Update function
type tickMsg time.Time
type cleanerDoneMsg struct {
	index      int
	spaceFreed uint64
	duration   time.Duration
	err        error
}

// InitialModel creates the initial model for the TUI
func InitialModel() Model {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// Create progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Get all cleaners
	cleanerTypes := cleaners.GetAllCleanupTypes()
	items := make([]list.Item, len(cleanerTypes))
	
	for i, cleanerType := range cleanerTypes {
		cleaner, exists := cleaners.GetCleaner(cleanerType)
		items[i] = CleanerItem{
			ID:          cleanerType,
			Name:        strings.Title(strings.ReplaceAll(cleanerType, "-", " ")),
			Desc:        getCleanerDescription(cleanerType),
			Category:    getCleanerCategory(cleanerType),
			Selected:    false,
			NeedsConfirm: exists && cleaner.RequiresConfirmation,
		}
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Cleaners to Run"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return Model{
		state:           WelcomeView,
		cleanerList:     l,
		selectedCleaners: []CleanerItem{},
		progressItems:   []ProgressItem{},
		progressSpinner: s,
		progressBar:     prog,
		startSpace:      utils.GetFreeDiskSpace(),
	}
}

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.progressSpinner.Tick,
		tickCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.cleanerList.SetSize(msg.Width-4, msg.Height-8)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case WelcomeView:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "enter", " ":
				m.state = SelectionView
				return m, nil
			}

		case SelectionView:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				// Start preview/execution
				m.updateSelectedCleaners()
				if len(m.selectedCleaners) > 0 {
					m.state = PreviewView
				}
				return m, nil
			case " ":
				// Toggle selection
				if item, ok := m.cleanerList.SelectedItem().(CleanerItem); ok {
					item.Selected = !item.Selected
					m.cleanerList.SetItem(m.cleanerList.Index(), item)
				}
				return m, nil
			}

		case PreviewView:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "enter", "y":
				m.state = ProgressView
				return m, m.startCleanup()
			case "esc", "n":
				m.state = SelectionView
				return m, nil
			}

		case ProgressView:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			}

		case ResultsView:
			switch msg.String() {
			case "q", "ctrl+c", "enter", "esc":
				m.quitting = true
				return m, tea.Quit
			case "r":
				// Restart
				return InitialModel(), nil
			}
		}

	case tickMsg:
		var cmd tea.Cmd
		m.progressSpinner, cmd = m.progressSpinner.Update(msg)
		cmds = append(cmds, cmd, tickCmd())

	case cleanerDoneMsg:
		if msg.index < len(m.progressItems) {
			m.progressItems[msg.index].SpaceFreed = msg.spaceFreed
			m.progressItems[msg.index].Duration = msg.duration
			m.progressItems[msg.index].Error = msg.err
			m.progressItems[msg.index].Progress = 1.0
			
			if msg.err != nil {
				m.progressItems[msg.index].Status = "Error"
			} else {
				m.progressItems[msg.index].Status = "Complete"
			}
		}
		
		m.currentCleaner++
		m.totalProgress = float64(m.currentCleaner) / float64(len(m.selectedCleaners))
		
		if m.currentCleaner >= len(m.selectedCleaners) {
			m.endSpace = utils.GetFreeDiskSpace()
			m.state = ResultsView
		}
		
		return m, nil
	}

	// Update list when in selection view
	if m.state == SelectionView {
		var cmd tea.Cmd
		m.cleanerList, cmd = m.cleanerList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current view
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case WelcomeView:
		return m.welcomeView()
	case SelectionView:
		return m.selectionView()
	case PreviewView:
		return m.previewView()
	case ProgressView:
		return m.progressView()
	case ResultsView:
		return m.resultsView()
	default:
		return "Unknown view"
	}
}

func (m Model) welcomeView() string {
	banner := bannerStyle.Render(`
╔══════════════════════════════════════╗
║  ▗▄▄▖ ▗▄▄▖  ▗▄▖  ▗▄▖ ▗▖  ▗▖         ║
║  ▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▛▚▞▜▌         ║
║  ▐▛▀▚▖▐▛▀▚▖▐▌ ▐▌▐▌ ▐▌▐▌  ▐▌         ║
║  ▐▙▄▞▘▐▌ ▐▌▝▚▄▞▘▝▚▄▞▘▐▌  ▐▌         ║
╚══════════════════════════════════════╝`)

	info := cardStyle.Render(fmt.Sprintf(`
🧹 System Cleanup Utility

Current free space: %s
Available cleaners: %d

This interactive tool will help you safely clean up your system
and free disk space by removing unnecessary files.

Press Enter to begin or 'q' to quit`,
		utils.FormatBytes(m.startSpace),
		len(cleaners.GetAllCleanupTypes())))

	help := subtleStyle.Render("Press Enter to continue • q to quit")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		banner,
		"",
		info,
		"",
		help,
	)
}

func (m Model) selectionView() string {
	header := headerStyle.Render("🔧 Select Cleaners")
	
	help := subtleStyle.Render("Use ↑↓ to navigate • Space to select • Enter to continue • q to quit")
	
	listView := m.cleanerList.View()
	
	var selectedInfo string
	selectedCount := len(m.getSelectedItems())
	if selectedCount > 0 {
		selectedInfo = successStyle.Render(fmt.Sprintf("Selected: %d cleaners", selectedCount))
	} else {
		selectedInfo = warningStyle.Render("No cleaners selected")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		listView,
		"",
		selectedInfo,
		help,
	)
}

func (m Model) previewView() string {
	header := headerStyle.Render("📋 Preview Selected Cleaners")
	
	var preview strings.Builder
	for _, cleaner := range m.selectedCleaners {
		style := cardStyle
		if cleaner.NeedsConfirm {
			style = selectedStyle
		}
		
		status := ""
		if cleaner.NeedsConfirm {
			status = warningStyle.Render(" (requires confirmation)")
		}
		
		preview.WriteString(style.Render(fmt.Sprintf("• %s%s\n  %s", 
			cleaner.Name, status, cleaner.Desc)))
		preview.WriteString("\n")
	}
	
	help := subtleStyle.Render("Enter/y to proceed • Esc/n to go back • q to quit")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		preview.String(),
		help,
	)
}

func (m Model) progressView() string {
	header := headerStyle.Render(fmt.Sprintf("🚀 Running Cleanup (%d/%d)", 
		m.currentCleaner, len(m.selectedCleaners)))
	
	overallProgress := m.progressBar.ViewAs(m.totalProgress)
	
	var progress strings.Builder
	for i, item := range m.progressItems {
		var status string
		var style lipgloss.Style
		
		switch item.Status {
		case "Running":
			status = fmt.Sprintf("%s %s", m.progressSpinner.View(), item.Name)
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))
		case "Complete":
			status = fmt.Sprintf("✓ %s - %s freed", item.Name, utils.FormatBytes(item.SpaceFreed))
			style = successStyle
		case "Error":
			status = fmt.Sprintf("✗ %s - Error", item.Name)
			style = errorStyle
		default:
			status = fmt.Sprintf("⋯ %s - Pending", item.Name)
			style = subtleStyle
		}
		
		if i == m.currentCleaner && item.Status == "Running" {
			progress.WriteString(selectedStyle.Render(style.Render(status)))
		} else {
			progress.WriteString(style.Render(status))
		}
		progress.WriteString("\n")
	}
	
	help := subtleStyle.Render("Cleanup in progress... • q to quit")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		overallProgress,
		"",
		progress.String(),
		"",
		help,
	)
}

func (m Model) resultsView() string {
	header := headerStyle.Render("📊 Cleanup Results")
	
	totalFreed := uint64(0)
	successCount := 0
	errorCount := 0
	
	for _, item := range m.progressItems {
		totalFreed += item.SpaceFreed
		if item.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}
	
	summary := cardStyle.Render(fmt.Sprintf(`
🎉 Cleanup Complete!

Total space freed: %s
Successful cleaners: %d
Failed cleaners: %d

Free space before: %s
Free space after: %s`,
		utils.FormatBytes(totalFreed),
		successCount,
		errorCount,
		utils.FormatBytes(m.startSpace),
		utils.FormatBytes(m.endSpace)))
	
	var details strings.Builder
	details.WriteString("Detailed Results:\n\n")
	
	for _, item := range m.progressItems {
		var status string
		var style lipgloss.Style
		
		if item.Error != nil {
			status = fmt.Sprintf("✗ %s - Error: %v", item.Name, item.Error)
			style = errorStyle
		} else {
			status = fmt.Sprintf("✓ %s - %s freed in %v", 
				item.Name, utils.FormatBytes(item.SpaceFreed), item.Duration)
			style = successStyle
		}
		
		details.WriteString(style.Render(status))
		details.WriteString("\n")
	}
	
	help := subtleStyle.Render("Enter/q to quit • r to restart")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		"",
		details.String(),
		help,
	)
}

// Helper functions
func (m *Model) updateSelectedCleaners() {
	selected := m.getSelectedItems()
	m.selectedCleaners = make([]CleanerItem, len(selected))
	copy(m.selectedCleaners, selected)
	
	// Initialize progress items
	m.progressItems = make([]ProgressItem, len(m.selectedCleaners))
	for i, cleaner := range m.selectedCleaners {
		m.progressItems[i] = ProgressItem{
			Name:     cleaner.Name,
			Status:   "Pending",
			Progress: 0.0,
		}
	}
}

func (m *Model) getSelectedItems() []CleanerItem {
	var selected []CleanerItem
	for _, item := range m.cleanerList.Items() {
		if cleanerItem, ok := item.(CleanerItem); ok && cleanerItem.Selected {
			selected = append(selected, cleanerItem)
		}
	}
	return selected
}

func (m Model) startCleanup() tea.Cmd {
	return func() tea.Msg {
		// This would start the actual cleanup process
		// For now, simulate it
		return cleanerDoneMsg{
			index:      0,
			spaceFreed: 1024 * 1024, // 1MB for demo
			duration:   time.Second,
			err:        nil,
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func getCleanerDescription(cleanerType string) string {
	descriptions := map[string]string{
		"apt-cache":          "Remove APT package cache files",
		"journal-logs":       "Clean systemd journal logs",
		"old-kernels":        "Remove old kernel packages",
		"temp-files":         "Remove temporary files",
		"user-cache":         "Clean user cache directories",
		"trash":              "Empty trash/recycle bin",
		"browser-cache":      "Clean web browser caches",
		"docker-system":      "Clean Docker system data",
		"yarn-cache":         "Clean Yarn package cache",
		"npm-cache":          "Clean npm package cache",
		"pip-cache":          "Clean pip package cache",
	}
	
	if desc, exists := descriptions[cleanerType]; exists {
		return desc
	}
	return "System cleanup operation"
}

func getCleanerCategory(cleanerType string) string {
	categories := map[string]string{
		"apt-cache":          "System",
		"journal-logs":       "System", 
		"old-kernels":        "System",
		"temp-files":         "User",
		"user-cache":         "User",
		"trash":              "User",
		"browser-cache":      "Applications",
		"docker-system":      "Virtualization",
		"yarn-cache":         "Applications",
		"npm-cache":          "Applications",
		"pip-cache":          "Applications",
	}
	
	if cat, exists := categories[cleanerType]; exists {
		return cat
	}
	return "General"
}

// RunTUI starts the TUI application
func RunTUI(ctx context.Context) error {
	// Use the simpler TUI for now
	return RunSimpleTUI(ctx)
}