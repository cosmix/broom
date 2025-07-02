// Package tui provides terminal user interface components for broom
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

// SimpleModel represents a simplified TUI model
type SimpleModel struct {
	choices       []string
	selected      map[int]bool
	cursor        int
	stage         Stage
	executing     bool
	results       []CleanupResult
	currentIndex  int
	progress      string
	startSpace    uint64
	endSpace      uint64
	quitting      bool
	
	// Viewport tracking
	width         int
	height        int
	viewportStart int
	
	// Progress log for execution view
	progressLog   []string
	logViewStart  int
}

type Stage int

const (
	Selection Stage = iota
	Confirmation
	Execution
	Results
)

type CleanupResult struct {
	Name       string
	Success    bool
	SpaceFreed uint64
	Duration   time.Duration
	Error      error
}

var (
	simpleFocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	simpleBlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C7C7C"))
	simpleCursorStyle  = simpleFocusedStyle.Copy()
	simpleNoStyle      = lipgloss.NewStyle()
	
	simpleHelpStyle = simpleBlurredStyle.Copy()
	
	simpleSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6"))
	
	simpleTitleStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#7C3AED")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1).
		Bold(true)
		
	simpleHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Margin(1, 0)
)

func InitialSimpleModel() SimpleModel {
	cleanerTypes := cleaners.GetAllCleanupTypes()
	
	return SimpleModel{
		choices:    cleanerTypes,
		selected:   make(map[int]bool),
		cursor:     0,
		stage:      Selection,
		executing:  false,
		results:    []CleanupResult{},
		startSpace: utils.GetFreeDiskSpace(),
	}
}

func (m SimpleModel) Init() tea.Cmd {
	return nil
}

func (m SimpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		switch m.stage {
		case Selection:
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			
			case "up", "k":
				cols, _ := m.getGridDimensions()
				newCursor := m.cursor - cols
				if newCursor >= 0 {
					m.cursor = newCursor
				} else {
					// Move to first item
					m.cursor = 0
				}
				m.updateViewport()
			
			case "down", "j":
				cols, _ := m.getGridDimensions()
				newCursor := m.cursor + cols
				if newCursor < len(m.choices) {
					m.cursor = newCursor
				} else {
					// Move to last item
					m.cursor = len(m.choices) - 1
				}
				m.updateViewport()
			
			case "left", "h":
				if m.cursor > 0 {
					m.cursor--
					m.updateViewport()
				}
			
			case "right", "l":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
					m.updateViewport()
				}
			
			case "pgup":
				// Page up - move cursor up by viewport height
				availableHeight := m.height - 6
				if availableHeight <= 0 {
					availableHeight = 10
				}
				m.cursor -= availableHeight
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.updateViewport()
			
			case "pgdown":
				// Page down - move cursor down by viewport height
				availableHeight := m.height - 6
				if availableHeight <= 0 {
					availableHeight = 10
				}
				m.cursor += availableHeight
				if m.cursor >= len(m.choices) {
					m.cursor = len(m.choices) - 1
				}
				m.updateViewport()
			
			case " ":
				m.selected[m.cursor] = !m.selected[m.cursor]
			
			case "a":
				// Select all
				for i := range m.choices {
					m.selected[i] = true
				}
			
			case "A":
				// Deselect all
				m.selected = make(map[int]bool)
			
			case "enter":
				if len(m.getSelectedChoices()) > 0 {
					m.stage = Confirmation
				}
			}
		
		case Confirmation:
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			
			case "y", "enter":
				m.stage = Execution
				m.executing = true
				return m, m.executeCleanups()
			
			case "n", "esc":
				m.stage = Selection
			}
		
		case Execution:
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			}
		
		case Results:
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q", "enter":
				m.quitting = true
				return m, tea.Quit
			case "r":
				// Reset to selection view while preserving model structure
				m.stage = Selection
				m.selected = make(map[int]bool)
				m.cursor = 0
				m.viewportStart = 0
				m.results = []CleanupResult{}
				m.currentIndex = 0
				m.progressLog = []string{}
				m.logViewStart = 0
				m.startSpace = utils.GetFreeDiskSpace()
				return m, nil
			}
		}
		
	case CleanupCompleteMsg:
		m.results = msg.Results
		m.endSpace = utils.GetFreeDiskSpace()
		m.executing = false
		m.stage = Results
		return m, nil
		
	case CleanupProgressMsg:
		// Only update current progress, don't add to log
		m.progress = sanitizeLogMessage(msg.Message, m.width-10)
		return m, nil
		
	case CleanupStartMsg:
		m.currentIndex = msg.Index
		selectedChoices := m.getSelectedChoices()
		if m.currentIndex < len(selectedChoices) {
			choice := selectedChoices[m.currentIndex]
			// Standardized format: just the cleaner name being processed
			cleanerName := sanitizeLogMessage(strings.Title(strings.ReplaceAll(choice, "-", " ")), 25)
			logEntry := fmt.Sprintf("→ %s", cleanerName)
			m.addLogEntry(logEntry)
			return m, m.performSingleCleanup(choice)
		}
		return m, nil
		
	case CleanupItemCompleteMsg:
		m.results = append(m.results, msg.Result)
		
		cleanerName := sanitizeLogMessage(strings.Title(strings.ReplaceAll(msg.Result.Name, "-", " ")), 25)
		var logEntry string
		
		if msg.Result.Success {
			if msg.Result.SpaceFreed > 0 {
				logEntry = fmt.Sprintf("✓ %s (%s)", cleanerName, utils.FormatBytes(msg.Result.SpaceFreed))
			} else {
				logEntry = fmt.Sprintf("✓ %s (clean)", cleanerName)
			}
		} else {
			logEntry = fmt.Sprintf("✗ %s (failed)", cleanerName)
		}
		
		m.addLogEntry(logEntry)
		
		m.currentIndex++
		return m, m.executeNextCleanup()
	}
	
	return m, nil
}

func (m SimpleModel) View() string {
	if m.quitting {
		return ""
	}
	
	var b strings.Builder
	
	switch m.stage {
	case Selection:
		return m.selectionView(&b)
	case Confirmation:
		// Add title for other views
		b.WriteString(simpleTitleStyle.Render("🧹 BROOM - Interactive System Cleanup"))
		b.WriteString("\n\n")
		return m.confirmationView(&b)
	case Execution:
		b.WriteString(simpleTitleStyle.Render("🧹 BROOM - Interactive System Cleanup"))
		b.WriteString("\n\n")
		return m.executionView(&b)
	case Results:
		b.WriteString(simpleTitleStyle.Render("🧹 BROOM - Interactive System Cleanup"))
		b.WriteString("\n\n")
		return m.resultsView(&b)
	}
	
	return b.String()
}

func (m SimpleModel) selectionView(b *strings.Builder) string {
	b.WriteString(simpleTitleStyle.Render("🧹 BROOM - Interactive System Cleanup"))
	b.WriteString("\n")
	b.WriteString(simpleHeaderStyle.Render("Select cleaners to run:"))
	b.WriteString("\n\n")
	
	cols, _ := m.getGridDimensions()
	availableHeight := m.height - 6
	if availableHeight <= 0 {
		availableHeight = 10
	}
	
	// Calculate visible rows
	startRow := m.viewportStart / cols
	endRow := startRow + availableHeight
	totalRows := (len(m.choices) + cols - 1) / cols
	
	// Show scroll indicators if needed
	if startRow > 0 {
		b.WriteString(simpleBlurredStyle.Render("... (more above)"))
		b.WriteString("\n")
	}
	
	// Render grid
	for row := startRow; row < endRow && row < totalRows; row++ {
		var rowItems []string
		
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx >= len(m.choices) {
				break
			}
			
			choice := m.choices[idx]
			cursor := " "
			if m.cursor == idx {
				cursor = simpleCursorStyle.Render(">")
			}
			
			checked := " "
			if m.selected[idx] {
				checked = simpleSelectedStyle.Render("✓")
			}
			
			cleanerName := toTitle(strings.ReplaceAll(choice, "-", " "))
			// Truncate long names to fit in columns
			if len(cleanerName) > 18 {
				cleanerName = cleanerName[:15] + "..."
			}
			
			item := fmt.Sprintf("%s [%s] %-18s", cursor, checked, cleanerName)
			
			// Apply style to the entire item if focused
			if m.cursor == idx {
				item = simpleFocusedStyle.Render(item)
			}
			
			rowItems = append(rowItems, item)
		}
		
		// Join items in this row with spacing
		b.WriteString(strings.Join(rowItems, "  "))
		b.WriteString("\n")
	}
	
	if endRow < totalRows {
		b.WriteString(simpleBlurredStyle.Render("... (more below)"))
		b.WriteString("\n")
	}
	
	selectedCount := len(m.getSelectedChoices())
	totalCount := len(m.choices)
	fmt.Fprintf(b, "\nSelected: %d/%d cleaners", selectedCount, totalCount)
	
	b.WriteString("\n\n")
	b.WriteString(simpleHelpStyle.Render("↑/k up • ↓/j down • ←/h left • →/l right • PgUp/PgDn page • space select • a select all • A deselect all • enter continue • q quit"))
	
	return b.String()
}

func (m SimpleModel) confirmationView(b *strings.Builder) string {
	b.WriteString(simpleHeaderStyle.Render("Confirm cleanup operations:"))
	b.WriteString("\n\n")
	
	selectedChoices := m.getSelectedChoices()
	
	// Show summary first
	fmt.Fprintf(b, "About to run %d cleaners:\n", len(selectedChoices))
	fmt.Fprintf(b, "Current free space: %s\n\n", utils.FormatBytes(m.startSpace))
	
	// Grid layout for selected cleaners (3-4 columns)
	cols := 3
	if m.width > 120 {
		cols = 4
	} else if m.width < 90 {
		cols = 2
	} else if m.width < 60 {
		cols = 1
	}
	
	// Render selected cleaners in grid
	for i := 0; i < len(selectedChoices); i += cols {
		var rowItems []string
		
		for col := 0; col < cols && i+col < len(selectedChoices); col++ {
			choice := selectedChoices[i+col]
			cleanerName := toTitle(strings.ReplaceAll(choice, "-", " "))
			
			// Truncate name to fit in column
			maxLen := 20
			if cols >= 3 {
				maxLen = 16
			}
			if len(cleanerName) > maxLen {
				cleanerName = cleanerName[:maxLen-3] + "..."
			}
			
			item := fmt.Sprintf("✓ %-*s", maxLen, cleanerName)
			item = simpleSelectedStyle.Render(item)
			rowItems = append(rowItems, item)
		}
		
		// Join items in this row with spacing
		if cols > 1 {
			b.WriteString(strings.Join(rowItems, "  "))
		} else {
			b.WriteString(strings.Join(rowItems, ""))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString(simpleHelpStyle.Render("y/enter proceed • n/esc back • q quit"))
	
	return b.String()
}

func (m SimpleModel) executionView(b *strings.Builder) string {
	selectedChoices := m.getSelectedChoices()
	
	// Header
	b.WriteString(simpleHeaderStyle.Render("Executing cleanup operations..."))
	b.WriteString("\n\n")
	
	// Status grid header
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(0, 1).
		Render("CLEANER STATUS"))
	b.WriteString("\n\n")
	
	// Use responsive columns
	cols := 4
	if m.width < 120 {
		cols = 3
	} else if m.width < 90 {
		cols = 2
	} else if m.width < 60 {
		cols = 1
	}
	
	// Render status grid
	for i := 0; i < len(selectedChoices); i += cols {
		var rowItems []string
		
		for col := 0; col < cols && i+col < len(selectedChoices); col++ {
			idx := i + col
			choice := selectedChoices[idx]
			cleanerName := toTitle(strings.ReplaceAll(choice, "-", " "))
			
			// Truncate name to fit in compact grid
			maxLen := 12
			if cols <= 2 {
				maxLen = 20
			}
			if len(cleanerName) > maxLen {
				cleanerName = cleanerName[:maxLen-3] + "..."
			}
			
			var statusIcon string
			if idx < m.currentIndex {
				statusIcon = simpleSelectedStyle.Render("✓")
			} else if idx == m.currentIndex {
				statusIcon = simpleFocusedStyle.Render("⚡")
			} else {
				statusIcon = simpleBlurredStyle.Render("⋯")
			}
			
			item := fmt.Sprintf("%s %-*s", statusIcon, maxLen, cleanerName)
			rowItems = append(rowItems, item)
		}
		
		b.WriteString(strings.Join(rowItems, " "))
		b.WriteString("\n")
	}
	
	// Separator
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(m.width, 80)))
	b.WriteString("\n")
	
	// Progress log header
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#01FAC6")).
		Padding(0, 1).
		Render("PROGRESS LOG"))
	b.WriteString("\n\n")
	
	// Calculate log space - be more conservative with height
	statusRows := (len(selectedChoices) + cols - 1) / cols
	usedHeight := 8 + statusRows + 3 // header + status + log header + help
	logHeight := m.height - usedHeight
	if logHeight < 3 {
		logHeight = 3
	} else if logHeight > 8 {
		logHeight = 8 // Limit log to prevent overflow
	}
	
	// Show recent log entries
	if len(m.progressLog) == 0 {
		b.WriteString(simpleBlurredStyle.Render("Ready to start cleanup operations..."))
	} else {
		start := len(m.progressLog) - logHeight
		if start < 0 {
			start = 0
		}
		
		for i := start; i < len(m.progressLog) && i < start+logHeight; i++ {
			logLine := sanitizeLogMessage(m.progressLog[i], m.width-5)
			b.WriteString(logLine)
			b.WriteString("\n")
		}
		
		// Show scroll indicator if needed
		if start > 0 {
			fmt.Fprintf(b, "%s", simpleBlurredStyle.Render(fmt.Sprintf("... (%d more entries above)", start)))
		}
	}
	
	b.WriteString("\n\n")
	b.WriteString(simpleHelpStyle.Render("Please wait... • q to quit"))
	
	return b.String()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// addLogEntry adds a log entry with proper length control to prevent layout issues
func (m *SimpleModel) addLogEntry(entry string) {
	// Sanitize and limit to prevent layout breakage
	sanitized := sanitizeLogMessage(entry, m.width-5)
	m.progressLog = append(m.progressLog, sanitized)
	
	// Limit log history to prevent memory issues
	maxLogEntries := 100
	if len(m.progressLog) > maxLogEntries {
		m.progressLog = m.progressLog[len(m.progressLog)-maxLogEntries:]
	}
}

// sanitizeLogMessage cleans up error messages to prevent layout issues
func sanitizeLogMessage(msg string, maxWidth int) string {
	// Remove newlines and replace with spaces
	cleaned := strings.ReplaceAll(msg, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	
	// Replace multiple spaces with single space
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	
	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	// Truncate if too long
	if maxWidth > 0 && len(cleaned) > maxWidth {
		if maxWidth > 3 {
			cleaned = cleaned[:maxWidth-3] + "..."
		} else {
			cleaned = cleaned[:maxWidth]
		}
	}
	
	return cleaned
}

func (m SimpleModel) resultsView(b *strings.Builder) string {
	totalFreed := uint64(0)
	successCount := 0
	
	for _, result := range m.results {
		totalFreed += result.SpaceFreed
		if result.Success {
			successCount++
		}
	}
	
	// Prominent header with space savings
	b.WriteString(simpleTitleStyle.Render("🎉 CLEANUP COMPLETE"))
	b.WriteString("\n\n")
	
	// Create a prominent summary box
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Margin(0, 2)
	
	summary := fmt.Sprintf(
		"%s SPACE FREED\n\n"+
		"Before: %s\n"+
		"After:  %s\n"+
		"Freed:  %s\n\n"+
		"Success: %d/%d cleaners",
		simpleSelectedStyle.Render("📈 "+utils.FormatBytes(totalFreed)),
		utils.FormatBytes(m.startSpace),
		utils.FormatBytes(m.endSpace),
		simpleSelectedStyle.Render(utils.FormatBytes(totalFreed)),
		successCount, len(m.results))
	
	b.WriteString(summaryStyle.Render(summary))
	b.WriteString("\n\n")
	
	// Grid layout for results (2-3 columns depending on width)
	cols := 2
	if m.width > 120 {
		cols = 3
	} else if m.width < 80 {
		cols = 1
	}
	
	b.WriteString(simpleHeaderStyle.Render("Results Summary:"))
	b.WriteString("\n\n")
	
	// Render results in grid
	for i := 0; i < len(m.results); i += cols {
		var rowItems []string
		
		for col := 0; col < cols && i+col < len(m.results); col++ {
			result := m.results[i+col]
			cleanerName := strings.Title(strings.ReplaceAll(result.Name, "-", " "))
			
			// Calculate column widths based on terminal size
			maxNameLen := 16
			maxSpaceLen := 10
			if cols == 1 {
				maxNameLen = 25
				maxSpaceLen = 12
			} else if cols == 2 {
				maxNameLen = 20
				maxSpaceLen = 11
			}
			
			// Truncate name to fit in column
			if len(cleanerName) > maxNameLen {
				cleanerName = cleanerName[:maxNameLen-3] + "..."
			}
			
			var statusIcon, spaceInfo string
			if result.Success {
				statusIcon = simpleSelectedStyle.Render("✓")
				if result.SpaceFreed > 0 {
					spaceInfo = utils.FormatBytes(result.SpaceFreed)
				} else {
					spaceInfo = "No data"
				}
			} else {
				statusIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5733")).Render("✗")
				spaceInfo = "Failed"
			}
			
			// Format with consistent spacing for both name and space columns
			item := fmt.Sprintf("%s %-*s %-*s", statusIcon, maxNameLen, cleanerName, maxSpaceLen, spaceInfo)
			rowItems = append(rowItems, item)
		}
		
		// Join items in this row with spacing
		if cols > 1 {
			b.WriteString(strings.Join(rowItems, "  "))
		} else {
			b.WriteString(strings.Join(rowItems, ""))
		}
		b.WriteString("\n")
	}
	
	// Show any errors separately
	errorCount := len(m.results) - successCount
	if errorCount > 0 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5733")).Render(
			fmt.Sprintf("⚠ %d cleaner(s) failed - check logs for details", errorCount)))
		b.WriteString("\n")
	}
	
	b.WriteString("\n\n")
	b.WriteString(simpleHelpStyle.Render("enter/q quit • r restart"))
	
	return b.String()
}

func (m SimpleModel) getSelectedChoices() []string {
	var selected []string
	for i, choice := range m.choices {
		if m.selected[i] {
			selected = append(selected, choice)
		}
	}
	return selected
}

// updateViewport ensures the cursor is visible within the viewport
func (m *SimpleModel) updateViewport() {
	if m.height <= 0 {
		return
	}
	
	// Reserve space for header, help text, and selected count (approximately 6 lines)
	availableHeight := m.height - 6
	if availableHeight <= 0 {
		availableHeight = 10 // Minimum viewport size
	}
	
	// Get grid dimensions and cursor position
	cols, _ := m.getGridDimensions()
	cursorRow, _ := m.getGridPosition()
	
	// Calculate viewport in terms of rows
	viewportStartRow := m.viewportStart / cols
	viewportEndRow := viewportStartRow + availableHeight - 1
	
	// Ensure cursor row is visible
	if cursorRow < viewportStartRow {
		viewportStartRow = cursorRow
	} else if cursorRow > viewportEndRow {
		viewportStartRow = cursorRow - availableHeight + 1
	}
	
	// Keep viewport bounds in check
	if viewportStartRow < 0 {
		viewportStartRow = 0
	}
	
	// Convert back to linear index
	m.viewportStart = viewportStartRow * cols
	
	// Ensure we don't go past the end
	if m.viewportStart >= len(m.choices) {
		m.viewportStart = ((len(m.choices) - 1) / cols) * cols
	}
}

// getGridDimensions calculates optimal grid layout for the given terminal size
func (m SimpleModel) getGridDimensions() (cols, rows int) {
	if m.width <= 0 {
		return 1, len(m.choices)
	}
	
	// Estimate space needed per item (cursor + checkbox + name + padding)
	avgNameLength := 20 // Reasonable estimate for cleaner names
	itemWidth := 4 + avgNameLength // "> [✓] " + name
	
	// Calculate how many columns fit
	cols = m.width / itemWidth
	if cols < 1 {
		cols = 1
	}
	if cols > 4 { // Limit to max 4 columns for readability
		cols = 4
	}
	
	// Calculate rows needed
	rows = (len(m.choices) + cols - 1) / cols
	
	return cols, rows
}

// getGridPosition converts linear cursor position to grid coordinates
func (m SimpleModel) getGridPosition() (row, col int) {
	cols, _ := m.getGridDimensions()
	row = m.cursor / cols
	col = m.cursor % cols
	return row, col
}


// CleanupCompleteMsg represents completion of all cleanup operations
type CleanupCompleteMsg struct {
	Results []CleanupResult
}

type CleanupProgressMsg struct {
	Index   int
	Message string
}

type CleanupStartMsg struct {
	Index int
	Name  string
}

type CleanupItemCompleteMsg struct {
	Result CleanupResult
}

func (m SimpleModel) executeCleanups() tea.Cmd {
	return m.executeNextCleanup()
}

func (m SimpleModel) executeNextCleanup() tea.Cmd {
	return func() tea.Msg {
		selectedChoices := m.getSelectedChoices()
		
		if m.currentIndex >= len(selectedChoices) {
			// All done
			return CleanupCompleteMsg{Results: m.results}
		}
		
		choice := selectedChoices[m.currentIndex]
		cleanerName := strings.Title(strings.ReplaceAll(choice, "-", " "))
		
		// Send start message
		return CleanupStartMsg{
			Index: m.currentIndex,
			Name:  cleanerName,
		}
	}
}

func (m SimpleModel) performSingleCleanup(choice string) tea.Cmd {
	return func() tea.Msg {
		// Execute the cleanup
		start := time.Now()
		spaceFreed, err := cleaners.PerformCleanup(choice)
		duration := time.Since(start)
		
		result := CleanupResult{
			Name:       choice,
			Success:    err == nil,
			SpaceFreed: spaceFreed,
			Duration:   duration,
			Error:      err,
		}
		
		return CleanupItemCompleteMsg{Result: result}
	}
}


// toTitle converts a string to title case (simple implementation)
func toTitle(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// RunSimpleTUI starts the simplified TUI
func RunSimpleTUI(ctx context.Context) error {
	p := tea.NewProgram(InitialSimpleModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}