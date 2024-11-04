package cleaners

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cosmix/broom/internal/utils"
)

const (
	defaultMinFileSize = 10 * 1024 * 1024     // 10MB
	defaultMaxFileAge  = 365 * 24 * time.Hour // 1 year
)

func init() {
	registerCleanup("old-files", Cleaner{CleanupFunc: cleanOldFiles, RequiresConfirmation: true})
}

type fileInfo struct {
	path         string
	size         int64
	lastModified time.Time
	selected     bool
}

type model struct {
	table    table.Model
	files    []fileInfo
	quitting bool
}

// FileCleanupOptions allows customization of file cleanup parameters
type FileCleanupOptions struct {
	MinFileSize int64
	MaxFileAge  time.Duration
	DryRun      bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case " ":
			row := m.table.Cursor()
			if row < len(m.files) {
				m.files[row].selected = !m.files[row].selected
				rows := makeRows(m.files)
				m.table.SetRows(rows)
			}
			return m, nil
		case "enter":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.table.View() + "\nUse arrow keys to navigate, space to select/deselect, enter to confirm, q to quit\n"
}

func makeRows(files []fileInfo) []table.Row {
	rows := make([]table.Row, len(files))
	for i, f := range files {
		selected := " "
		if f.selected {
			selected = "×"
		}
		rows[i] = table.Row{
			selected,
			utils.FormatBytes(uint64(f.size)),
			f.lastModified.Format("2006-01-02"),
			filepath.Base(f.path),
			f.path,
		}
	}
	return rows
}

func findOldFiles(opts ...FileCleanupOptions) ([]fileInfo, error) {
	var files []fileInfo
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Use default options if not provided
	options := FileCleanupOptions{
		MinFileSize: defaultMinFileSize,
		MaxFileAge:  defaultMaxFileAge,
	}
	if len(opts) > 0 {
		options = opts[0]
	}

	err = filepath.Walk(homeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and symlinks
		if info.IsDir() || (info.Mode()&os.ModeSymlink) != 0 {
			return nil
		}

		// Check file size and age
		if info.Size() > options.MinFileSize {
			// Check if file is older than specified age
			oldestAllowedTime := time.Now().Add(-options.MaxFileAge)
			if info.ModTime().Before(oldestAllowedTime) {
				files = append(files, fileInfo{
					path:         path,
					size:         info.Size(),
					lastModified: info.ModTime(),
					selected:     false,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking file system: %v", err)
	}

	// Sort by size (descending) and then by date (oldest first)
	sort.Slice(files, func(i, j int) bool {
		if files[i].size != files[j].size {
			return files[i].size > files[j].size
		}
		return files[i].lastModified.Before(files[j].lastModified)
	})

	return files, nil
}

func cleanOldFiles() error {
	files, err := findOldFiles()
	if err != nil {
		return fmt.Errorf("failed to find old files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No old files found")
		return nil
	}

	// In test environment, automatically select all files
	if os.Getenv("GO_TEST_ENV") == "true" {
		for i := range files {
			files[i].selected = true
		}
	} else {
		columns := []table.Column{
			{Title: " ", Width: 3},
			{Title: "Size", Width: 10},
			{Title: "Modified", Width: 10},
			{Title: "Name", Width: 30},
			{Title: "Path", Width: 50},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(makeRows(files)),
			table.WithFocused(true),
			table.WithHeight(10),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		m := model{table: t, files: files}

		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("error running program: %v", err)
		}

		finalM, ok := finalModel.(model)
		if !ok || finalM.quitting {
			return nil
		}

		files = finalM.files
	}

	var selectedFiles []fileInfo
	for _, f := range files {
		if f.selected {
			selectedFiles = append(selectedFiles, f)
		}
	}

	if len(selectedFiles) == 0 {
		fmt.Println("No files selected for deletion")
		return nil
	}

	fmt.Printf("\nDeleting %d selected files...\n", len(selectedFiles))
	var deletionErrors []error
	for _, f := range selectedFiles {
		err := os.Remove(f.path)
		if err != nil {
			deletionError := fmt.Errorf("error deleting %s: %v", f.path, err)
			deletionErrors = append(deletionErrors, deletionError)
			fmt.Println(deletionError)
		} else {
			fmt.Printf("Deleted %s\n", f.path)
		}
	}

	if len(deletionErrors) > 0 {
		return fmt.Errorf("%d files could not be deleted", len(deletionErrors))
	}

	return nil
}
