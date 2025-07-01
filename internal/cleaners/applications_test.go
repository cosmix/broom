package cleaners

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cosmix/broom/internal/utils"
)

type condaEnvList struct {
	Envs []string `json:"envs"`
}

type MockRunner struct {
	RunFdOrFindCalls      []RunFdOrFindCall
	RunWithIndicatorCalls []RunWithIndicatorCall
	RunWithOutputCalls    []RunWithOutputCall
	fdOrFindErr           error
}

type RunFdOrFindCall struct {
	Dir     string
	Args    string
	Message string
	Sudo    bool
	Err     error
}

type RunWithIndicatorCall struct {
	Command string
	Message string
	Err     error
}

type RunWithOutputCall struct {
	Command string
	Output  string
	Err     error
}

func (m *MockRunner) RunFdOrFind(dir, args, message string, sudo bool) error {
	call := RunFdOrFindCall{Dir: dir, Args: args, Message: message, Sudo: sudo, Err: m.fdOrFindErr}
	m.RunFdOrFindCalls = append(m.RunFdOrFindCalls, call)
	return m.fdOrFindErr
}

func (m *MockRunner) RunWithIndicator(command, message string) error {
	if len(m.RunWithIndicatorCalls) > 0 {
		for _, call := range m.RunWithIndicatorCalls {
			if (call.Command != "" && call.Command == command) || (call.Command == "" && call.Err != nil) {
				if call.Err != nil {
					return call.Err
				}
				break
			}
		}
	}
	m.RunWithIndicatorCalls = append(m.RunWithIndicatorCalls, RunWithIndicatorCall{Command: command, Message: message})
	return nil
}

func (m *MockRunner) RunWithOutput(command string) (string, error) {
	if len(m.RunWithOutputCalls) > 0 {
		return m.RunWithOutputCalls[0].Output, m.RunWithOutputCalls[0].Err
	}
	return "", nil
}

func TestCleanDocker(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"DockerInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"DockerNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "docker system prune -af",
					Message: "Removing unused Docker data",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanDocker(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanDocker() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanSnap(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"SnapInstalled", func(cmd string) bool { return true }, nil, false, 2},
		{"SnapNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"FirstCommandError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "snap list --all | awk '/disabled/{print $1, $3}' | while read snapname revision; do sudo snap remove $snapname --revision=$revision; done",
					Message: "Removing old snap versions",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanSnap(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanSnap() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanFlatpak(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"FlatpakInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"FlatpakNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "flatpak uninstall --unused -y",
					Message: "Removing unused Flatpak runtimes",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanFlatpak(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanFlatpak() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanTimeshiftSnapshots(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"TimeshiftInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"TimeshiftNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "timeshift --list | grep -oP '(?<=\\s)\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}' | sort | head -n -3 | xargs -I {} timeshift --delete --snapshot '{}'",
					Message: "Removing old Timeshift snapshots",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanTimeshiftSnapshots(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanTimeshiftSnapshots() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanRubyGems(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"RubyGemsInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"RubyGemsNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "gem cleanup",
					Message: "Removing old Ruby gems",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanRubyGems(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanRubyGems() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanPythonCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		fdOrFindErr error
		expectErr   bool
	}{
		{"Success", nil, false},
		{"FdOrFindError", errors.New("fd error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			err := cleanPythonCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPythonCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != 2 {
				t.Errorf("Expected 2 calls to RunFdOrFind, got %d", len(mock.RunFdOrFindCalls))
			}
		})
	}
}

func TestCleanLibreOfficeCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		fdOrFindErr error
		expectErr   bool
	}{
		{"Success", nil, false},
		{"FdOrFindError", errors.New("fd error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			err := cleanLibreOfficeCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanLibreOfficeCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != 1 {
				t.Errorf("Expected 1 call to RunFdOrFind, got %d", len(mock.RunFdOrFindCalls))
			}
		})
	}
}

func TestClearBrowserCaches(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		fdOrFindErr error
		expectErr   bool
	}{
		{"Success", nil, false},
		{"FdOrFindError", errors.New("fd error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			err := clearBrowserCaches()

			if (err != nil) != tt.expectErr {
				t.Errorf("clearBrowserCaches() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != 3 {
				t.Errorf("Expected 3 calls to RunFdOrFind, got %d", len(mock.RunFdOrFindCalls))
			}
		})
	}
}

func TestCleanPackageManagerCaches(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{
			"AllPackageManagersInstalled",
			func(cmd string) bool { return true },
			nil,
			false,
			3,
		},
		{
			"NoPackageManagersInstalled",
			func(cmd string) bool { return false },
			nil,
			false,
			0,
		},
		{
			"AptGetError",
			func(cmd string) bool { return cmd == "apt-get" },
			errors.New("run error"),
			true,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "apt-get clean",
					Message: "Cleaning APT cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanPackageManagerCaches(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPackageManagerCaches() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanNpmCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"NpmInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"NpmNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "npm cache clean --force",
					Message: "Cleaning npm cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanNpmCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanNpmCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanYarnCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"YarnInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"YarnNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "yarn cache clean",
					Message: "Cleaning yarn cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanYarnCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanYarnCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanPnpmCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"PnpmInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"PnpmNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "pnpm store prune",
					Message: "Cleaning pnpm store",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanPnpmCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPnpmCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanDenoCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"RunWithIndicatorError", errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.cache/deno/*",
					Message: "Cleaning Deno cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanDenoCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanDenoCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanBunCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"RunWithIndicatorError", errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.bun/install/cache/*",
					Message: "Cleaning Bun cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanBunCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanBunCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanPipCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"PipInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"PipNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "pip cache purge",
					Message: "Cleaning pip cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanPipCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPipCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanPoetryCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"PoetryInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"PoetryNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "poetry cache clear . --all",
					Message: "Cleaning poetry cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanPoetryCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPoetryCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanPipenvCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"RunWithIndicatorError", errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.cache/pipenv/*",
					Message: "Cleaning pipenv cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanPipenvCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPipenvCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanUvCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"UvInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"UvNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "uv cache clean",
					Message: "Cleaning uv cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanUvCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUvCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanGradleCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"RunWithIndicatorError", errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.gradle/caches",
					Message: "Cleaning Gradle cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanGradleCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanGradleCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanComposerCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"ComposerInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"ComposerNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "composer clear-cache",
					Message: "Cleaning Composer cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanComposerCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanComposerCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestRemoveOldWinePrefixes(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"WineInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"WineNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"FdOrFindError", func(cmd string) bool { return true }, errors.New("fd error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			cleanFunc := removeOldWinePrefixes(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("removeOldWinePrefixes() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunFdOrFind, got %d", tt.expectedCalls, len(mock.RunFdOrFindCalls))
			}
		})
	}
}

func TestCleanElectronCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		fdOrFindErr error
		expectErr   bool
	}{
		{"Success", nil, false},
		{"FdOrFindError", errors.New("fd error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			err := cleanElectronCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanElectronCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != 1 {
				t.Errorf("Expected 1 call to RunFdOrFind, got %d", len(mock.RunFdOrFindCalls))
			}
		})
	}
}

func TestCleanKdenliveRenderFiles(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"KdenliveInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"KdenliveNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"FdOrFindError", func(cmd string) bool { return true }, errors.New("fd error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			cleanFunc := cleanKdenliveRenderFiles(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanKdenliveRenderFiles() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunFdOrFind, got %d", tt.expectedCalls, len(mock.RunFdOrFindCalls))
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[0]
				if call.Dir != "$HOME" || call.Args != "-type f -path '*/kdenlive/render/*' -delete" ||
					call.Message != "Removing Kdenlive render files" || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}
		})
	}
}

func TestCleanBlenderTempFiles(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"BlenderInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"BlenderNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"FdOrFindError", func(cmd string) bool { return true }, errors.New("fd error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			cleanFunc := cleanBlenderTempFiles(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanBlenderTempFiles() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunFdOrFind, got %d", tt.expectedCalls, len(mock.RunFdOrFindCalls))
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[0]
				if call.Dir != "$HOME" || call.Args != "-type f -path '*/blender_*_autosave.blend' -delete" ||
					call.Message != "Removing Blender temporary files" || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}
		})
	}
}

func TestCleanSteamDownloadCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", func(cmd string) bool { return true }, nil, false, 1},
		{"SteamNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.steam/steam/steamapps/downloading/*",
					Message: "Clearing Steam download cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanSteamDownloadCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanSteamDownloadCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf $HOME/.steam/steam/steamapps/downloading/*" || call.Message != "Clearing Steam download cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanMySQLMariaDBBinlogs(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"MySQLInstalled", func(cmd string) bool { return cmd == "mysql" }, nil, false, 1},
		{"MariaDBInstalled", func(cmd string) bool { return cmd == "mariadb" }, nil, false, 1},
		{"NeitherInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: `mysql -e "PURGE BINARY LOGS BEFORE DATE(NOW() - INTERVAL 7 DAY);"`,
					Message: "Removing old MySQL/MariaDB binary logs",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanMySQLMariaDBBinlogs(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanMySQLMariaDBBinlogs() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != `mysql -e "PURGE BINARY LOGS BEFORE DATE(NOW() - INTERVAL 7 DAY);"` || call.Message != "Removing old MySQL/MariaDB binary logs" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanThunderbirdCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"ThunderbirdInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"ThunderbirdNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"FdOrFindError", func(cmd string) bool { return true }, errors.New("fd error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			cleanFunc := cleanThunderbirdCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanThunderbirdCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunFdOrFind, got %d", tt.expectedCalls, len(mock.RunFdOrFindCalls))
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[0]
				if call.Dir != "$HOME/.thunderbird" || call.Args != "-type d -name 'Cache' -exec rm -rf {}/* \\;" ||
					call.Message != "Clearing Thunderbird cache" || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}
		})
	}
}

func TestCleanDropboxCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"RunWithIndicatorError", errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.dropbox/cache/*",
					Message: "Clearing Dropbox cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanDropboxCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanDropboxCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf $HOME/.dropbox/cache/*" || call.Message != "Clearing Dropbox cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanMavenCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", func(cmd string) bool { return true }, nil, false, 1},
		{"MavenNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"Error", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf ~/.m2/repository",
					Message: "Cleaning Maven local repository cache...",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanMavenCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanMavenCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf ~/.m2/repository" || call.Message != "Cleaning Maven local repository cache..." {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanGoCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", func(cmd string) bool { return true }, nil, false, 1},
		{"GoNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"Error", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "go clean -modcache",
					Message: "Cleaning old Go modules cache...",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanGoCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanGoCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "go clean -modcache" || call.Message != "Cleaning old Go modules cache..." {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanRustCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", func(cmd string) bool { return true }, nil, false, 2},
		{"RustNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"FirstCommandError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf ~/.cargo/registry",
					Message: "Cleaning Rust cargo registry...",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanRustCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanRustCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call1 := mock.RunWithIndicatorCalls[0]
				if call1.Command != "rm -rf ~/.cargo/registry" || call1.Message != "Cleaning Rust cargo registry..." {
					t.Errorf("Unexpected arguments to first RunWithIndicator: %+v", call1)
				}

				if len(mock.RunWithIndicatorCalls) > 1 {
					call2 := mock.RunWithIndicatorCalls[1]
					if call2.Command != "rm -rf ~/.cargo/git" || call2.Message != "Cleaning Rust cargo git cache..." {
						t.Errorf("Unexpected arguments to second RunWithIndicator: %+v", call2)
					}
				}
			}
		})
	}
}

func TestCleanAndroidSDK(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		execOutput    string
		execErr       error
		expectErr     bool
		expectedCalls int
	}{
		{
			name:          "SDKManagerNotInstalled",
			commandExists: func(cmd string) bool { return false },
			expectedCalls: 0,
		},
		{
			name:          "ListError",
			commandExists: func(cmd string) bool { return true },
			execErr:       errors.New("list error"),
			expectErr:     true,
			expectedCalls: 1,
		},
		{
			name:          "NoPackagesToRemove",
			commandExists: func(cmd string) bool { return true },
			execOutput:    "other-package\nsome-package\n",
			expectedCalls: 1,
		},
		{
			name:          "RemovePackages",
			commandExists: func(cmd string) bool { return true },
			execOutput:    "system-images;android-30;google_apis;x86    | 30.0.3 | Installed\nemulator                                  | 30.0.12 | Installed\n",
			expectedCalls: 5, // 1 for list + 2 packages * 2 calls each
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.commandExists("sdkmanager") {
				mock.RunWithOutputCalls = []RunWithOutputCall{{
					Command: "sdkmanager --list_installed",
					Output:  tt.execOutput,
					Err:     tt.execErr,
				}}

				if tt.execOutput != "" && !tt.expectErr {
					lines := strings.Split(tt.execOutput, "\n")
					for _, line := range lines {
						if strings.Contains(line, "system-images") || strings.Contains(line, "emulator") {
							packageName := strings.Fields(line)[0]
							mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls,
								RunWithIndicatorCall{
									Command: fmt.Sprintf("sdkmanager --uninstall %s", packageName),
									Message: fmt.Sprintf("Removing Android SDK package: %s", packageName),
								})
						}
					}
				}
			}

			cleanFunc := cleanAndroidSDK(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanAndroidSDK() error = %v, expectErr %v", err, tt.expectErr)
			}

			totalCalls := len(mock.RunWithOutputCalls)
			if tt.execErr == nil && tt.commandExists("sdkmanager") {
				totalCalls += len(mock.RunWithIndicatorCalls)
			}
			if totalCalls != tt.expectedCalls {
				t.Errorf("Expected %d total call(s), got %d", tt.expectedCalls, totalCalls)
			}

			if tt.commandExists("sdkmanager") && len(mock.RunWithOutputCalls) > 0 {
				call := mock.RunWithOutputCalls[0]
				if call.Command != "sdkmanager --list_installed" {
					t.Errorf("Expected command 'sdkmanager --list_installed', got %s", call.Command)
				}
			}
		})
	}
}

func TestCleanJetBrainsIDECaches(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"Error", errors.New("clean error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "find ~/.local/share/JetBrains -type d -name '.caches' -exec rm -rf {} +",
					Message: "Cleaning JetBrains IDE caches",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanJetBrainsIDECaches()
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanJetBrainsIDECaches() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				expectedCmd := "find ~/.local/share/JetBrains -type d -name '.caches' -exec rm -rf {} +"
				if call.Command != expectedCmd || call.Message != "Cleaning JetBrains IDE caches" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanRPackagesCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"RNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"Success", func(cmd string) bool { return true }, nil, false, 1},
		{"Error", func(cmd string) bool { return true }, errors.New("clean error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: `R -e "remove.packages(installed.packages()[,1])"`,
					Message: "Cleaning R packages cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanRPackagesCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanRPackagesCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				expectedCmd := `R -e "remove.packages(installed.packages()[,1])"`
				if call.Command != expectedCmd || call.Message != "Cleaning R packages cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanJuliaPackagesCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"JuliaNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"Success", func(cmd string) bool { return true }, nil, false, 1},
		{"Error", func(cmd string) bool { return true }, errors.New("clean error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "julia -e 'using Pkg; Pkg.gc()'",
					Message: "Cleaning Julia packages cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanJuliaPackagesCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanJuliaPackagesCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				expectedCmd := "julia -e 'using Pkg; Pkg.gc()'"
				if call.Command != expectedCmd || call.Message != "Cleaning Julia packages cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanUnusedCondaEnvironments(t *testing.T) {
	originalRunner := utils.Runner
	defer func() {
		utils.Runner = originalRunner
	}()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		execOutput    string
		execErr       error
		expectErr     bool
		expectedCalls int
	}{
		{
			name:          "CondaNotInstalled",
			commandExists: func(cmd string) bool { return false },
			expectedCalls: 0,
		},
		{
			name:          "ListError",
			commandExists: func(cmd string) bool { return true },
			execErr:       errors.New("list error"),
			expectErr:     true,
			expectedCalls: 1,
		},
		{
			name:          "NoEnvironments",
			commandExists: func(cmd string) bool { return true },
			execOutput:    `{"envs": []}`,
			expectedCalls: 2, // 1 for list + 1 for parsing
		},
		{
			name:          "RemoveEnvironments",
			commandExists: func(cmd string) bool { return true },
			execOutput:    `{"envs": ["/home/user/anaconda3/envs/base", "/home/user/anaconda3/envs/env1", "/home/user/anaconda3/envs/env2"]}`,
			expectedCalls: 4, // 1 for list + 1 for parsing + 2 remove calls (excluding base)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.commandExists("conda") {
				mock.RunWithOutputCalls = []RunWithOutputCall{{
					Command: "conda env list --json",
					Output:  tt.execOutput,
					Err:     tt.execErr,
				}}

				if tt.execOutput != "" && !tt.expectErr {
					var envList condaEnvList
					if err := json.Unmarshal([]byte(tt.execOutput), &envList); err == nil {
						for _, env := range envList.Envs {
							envName := filepath.Base(env)
							if envName != "base" {
								mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls,
									RunWithIndicatorCall{
										Command: fmt.Sprintf("conda env remove --name %s", envName),
										Message: fmt.Sprintf("Removing Conda environment: %s", envName),
									})
							}
						}
					}
				}
			}

			cleanFunc := cleanUnusedCondaEnvironments(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUnusedCondaEnvironments() error = %v, expectErr %v", err, tt.expectErr)
			}

			totalCalls := len(mock.RunWithOutputCalls)
			if tt.execErr == nil && tt.commandExists("conda") {
				totalCalls += len(mock.RunWithIndicatorCalls)
			}
			if totalCalls != tt.expectedCalls {
				t.Errorf("Expected %d total call(s), got %d", tt.expectedCalls, totalCalls)
			}

			if tt.commandExists("conda") && len(mock.RunWithOutputCalls) > 0 {
				call := mock.RunWithOutputCalls[0]
				if call.Command != "conda env list --json" {
					t.Errorf("Expected command 'conda env list --json', got %s", call.Command)
				}
			}

			if !tt.expectErr && tt.commandExists("conda") && tt.execOutput != "" {
				var envList condaEnvList
				if err := json.Unmarshal([]byte(tt.execOutput), &envList); err == nil {
					for _, env := range envList.Envs {
						envName := filepath.Base(env)
						if envName != "base" {
							found := false
							for _, call := range mock.RunWithIndicatorCalls {
								if call.Command == fmt.Sprintf("conda env remove --name %s", envName) {
									found = true
									break
								}
							}
							if !found {
								t.Errorf("Expected call to remove environment %s not found", envName)
							}
						}
					}
				}
			}
		})
	}
}

func TestCleanMercurialBackups(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		fdOrFindErr   error
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{
			name:          "MercurialInstalled",
			commandExists: func(cmd string) bool { return true },
			expectedCalls: 2,
		},
		{
			name:          "MercurialNotInstalled",
			commandExists: func(cmd string) bool { return false },
			expectedCalls: 0,
		},
		{
			name:          "FdOrFindError",
			commandExists: func(cmd string) bool { return true },
			fdOrFindErr:   errors.New("fd error"),
			expectedCalls: 2,
		},
		{
			name:          "RunWithIndicatorError",
			commandExists: func(cmd string) bool { return true },
			withIndErr:    errors.New("run error"),
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.hg/bundle-backup/*",
					Message: "Removing Mercurial bundle backups",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanMercurialBackups(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanMercurialBackups() error = %v, expectErr %v", err, tt.expectErr)
			}

			totalCalls := len(mock.RunFdOrFindCalls) + len(mock.RunWithIndicatorCalls)
			if totalCalls != tt.expectedCalls {
				t.Errorf("Expected %d total call(s), got %d", tt.expectedCalls, totalCalls)
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[0]
				if call.Dir != "/home" || call.Args != "-type f -name '*.hg*.bak' -delete" ||
					call.Message != "Removing Mercurial backup files" || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}
		})
	}
}

func TestCleanGitLFSCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{
			name:          "GitLFSInstalled",
			commandExists: func(cmd string) bool { return true },
			expectedCalls: 1,
		},
		{
			name:          "GitLFSNotInstalled",
			commandExists: func(cmd string) bool { return false },
			expectedCalls: 0,
		},
		{
			name:          "RunWithIndicatorError",
			commandExists: func(cmd string) bool { return true },
			withIndErr:    errors.New("run error"),
			expectErr:     true,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "git lfs prune",
					Message: "Cleaning Git LFS cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanGitLFSCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanGitLFSCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "git lfs prune" || call.Message != "Cleaning Git LFS cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanCMakeBuildDirs(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		fdOrFindErr error
		expectErr   bool
	}{
		{
			name: "Success",
		},
		{
			name:        "FdOrFindError",
			fdOrFindErr: errors.New("fd error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			err := cleanCMakeBuildDirs()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanCMakeBuildDirs() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != 2 {
				t.Errorf("Expected 2 calls to RunFdOrFind, got %d", len(mock.RunFdOrFindCalls))
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call1 := mock.RunFdOrFindCalls[0]
				if call1.Dir != "/home" || call1.Args != "-type d -name 'build' -exec test -f '{}/CMakeCache.txt' \\; -exec rm -rf {} \\;" ||
					call1.Message != "Removing old CMake build directories" || !call1.Sudo {
					t.Errorf("Unexpected arguments to first RunFdOrFind: %+v", call1)
				}

				if len(mock.RunFdOrFindCalls) > 1 {
					call2 := mock.RunFdOrFindCalls[1]
					if call2.Dir != "/home" || call2.Args != "-type d -name 'CMakeFiles' -exec rm -rf {} \\;" ||
						call2.Message != "Removing CMakeFiles directories" || !call2.Sudo {
						t.Errorf("Unexpected arguments to second RunFdOrFind: %+v", call2)
					}
				}
			}
		})
	}
}

func TestCleanAutotoolsFiles(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	patterns := []string{
		"autom4te.cache",
		"config.status",
		"config.log",
		"configure~",
		"Makefile.in",
		"aclocal.m4",
		".deps",
		".libs",
	}

	tests := []struct {
		name        string
		fdOrFindErr error
		expectErr   bool
	}{
		{
			name: "Success",
		},
		{
			name:        "FdOrFindError",
			fdOrFindErr: errors.New("fd error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{fdOrFindErr: tt.fdOrFindErr}
			utils.Runner = mock

			err := cleanAutotoolsFiles()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanAutotoolsFiles() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != len(patterns) {
				t.Errorf("Expected %d calls to RunFdOrFind, got %d", len(patterns), len(mock.RunFdOrFindCalls))
			}

			for i, pattern := range patterns {
				if i < len(mock.RunFdOrFindCalls) {
					call := mock.RunFdOrFindCalls[i]
					expectedArgs := fmt.Sprintf("-type d,f -name '%s' -exec rm -rf {} \\;", pattern)
					expectedMessage := fmt.Sprintf("Removing Autotools generated %s", pattern)
					if call.Dir != "/home" || call.Args != expectedArgs ||
						call.Message != expectedMessage || !call.Sudo {
						t.Errorf("Unexpected arguments for pattern %s: %+v", pattern, call)
					}
				}
			}
		})
	}
}

func TestCleanCCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{
			name:          "CCacheInstalled",
			commandExists: func(cmd string) bool { return true },
			expectedCalls: 1,
		},
		{
			name:          "CCacheNotInstalled",
			commandExists: func(cmd string) bool { return false },
			expectedCalls: 0,
		},
		{
			name:          "RunWithIndicatorError",
			commandExists: func(cmd string) bool { return true },
			withIndErr:    errors.New("run error"),
			expectErr:     true,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "ccache -C",
					Message: "Clearing ccache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanCCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanCCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "ccache -C" || call.Message != "Clearing ccache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

// Phase 2: Modern DevOps Tools Tests

func TestCleanKubectlCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		withIndErr  error
		expectErr   bool
		expectedMsg string
	}{
		{"Success", nil, false, ""},
		{"FirstCommandError", errors.New("rm error"), false, "Warning: Error while cleaning kubectl cache: rm error\n"},
		{"SecondCommandError", errors.New("rm error"), false, "Warning: Error while cleaning kubectl HTTP cache: rm error\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.kube/cache/*",
					Message: "Cleaning kubectl cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanKubectlCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanKubectlCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			expectedCalls := 2
			if len(mock.RunWithIndicatorCalls) != expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanHelmCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name        string
		withIndErr  error
		expectErr   bool
		expectedMsg string
	}{
		{"Success", nil, false, ""},
		{"FirstCommandError", errors.New("rm error"), false, "Warning: Error while cleaning Helm cache: rm error\n"},
		{"SecondCommandError", errors.New("rm error"), false, "Warning: Error while cleaning Helm data: rm error\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.cache/helm/*",
					Message: "Cleaning Helm cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanHelmCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanHelmCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			expectedCalls := 2
			if len(mock.RunWithIndicatorCalls) != expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", expectedCalls, len(mock.RunWithIndicatorCalls))
			}
		})
	}
}

func TestCleanMinikubeCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"MinikubeInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"MinikubeNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.minikube/cache/*",
					Message: "Cleaning minikube cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanMinikubeCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanMinikubeCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf $HOME/.minikube/cache/*" || call.Message != "Cleaning minikube cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanTerraformCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name       string
		withIndErr error
		expectErr  bool
	}{
		{"Success", nil, false},
		{"RunWithIndicatorError", errors.New("run error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.terraform.d/plugin-cache/*",
					Message: "Cleaning Terraform plugin cache",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanTerraformCache()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanTerraformCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			expectedCalls := 1
			if len(mock.RunWithIndicatorCalls) != expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf $HOME/.terraform.d/plugin-cache/*" || call.Message != "Cleaning Terraform plugin cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanAnsibleTemp(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name       string
		withIndErr error
		expectErr  bool
	}{
		{"Success", nil, false},
		{"RunWithIndicatorError", errors.New("run error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf $HOME/.ansible/tmp/*",
					Message: "Cleaning Ansible temporary files",
					Err:     tt.withIndErr,
				}}
			}

			err := cleanAnsibleTemp()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanAnsibleTemp() error = %v, expectErr %v", err, tt.expectErr)
			}

			expectedCalls := 1
			if len(mock.RunWithIndicatorCalls) != expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf $HOME/.ansible/tmp/*" || call.Message != "Cleaning Ansible temporary files" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanContainerdCache(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"ContainerdInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"ContainerdNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "rm -rf /var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/*",
					Message: "Cleaning containerd cache",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanContainerdCache(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanContainerdCache() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "rm -rf /var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/*" || call.Message != "Cleaning containerd cache" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanPodmanSystem(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		commandExists utils.CommandExistsFunc
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"PodmanInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"PodmanNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = []RunWithIndicatorCall{{
					Command: "podman system prune -af",
					Message: "Cleaning podman system",
					Err:     tt.withIndErr,
				}}
			}

			cleanFunc := cleanPodmanSystem(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanPodmanSystem() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[0]
				if call.Command != "podman system prune -af" || call.Message != "Cleaning podman system" {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}
