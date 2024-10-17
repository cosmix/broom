package cleaners

import (
	"errors"
	"testing"

	"github.com/cosmix/broom/internal/utils"
)

type MockRunner struct {
	RunFdOrFindCalls      []RunFdOrFindCall
	RunWithIndicatorCalls []RunWithIndicatorCall
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

func (m *MockRunner) RunFdOrFind(dir, args, message string, sudo bool) error {
	call := RunFdOrFindCall{Dir: dir, Args: args, Message: message, Sudo: sudo}
	m.RunFdOrFindCalls = append(m.RunFdOrFindCalls, call)
	if len(m.RunFdOrFindCalls) > 0 && m.RunFdOrFindCalls[0].Err != nil {
		return m.RunFdOrFindCalls[0].Err
	}
	return nil
}

func (m *MockRunner) RunWithIndicator(command, message string) error {
	call := RunWithIndicatorCall{Command: command, Message: message}
	m.RunWithIndicatorCalls = append(m.RunWithIndicatorCalls, call)
	if len(m.RunWithIndicatorCalls) > 0 && m.RunWithIndicatorCalls[0].Err != nil {
		return m.RunWithIndicatorCalls[0].Err
	}
	return nil
}

func TestCleanHomeDirectory(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		fdOrFindErr   error
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, nil, false, 2},
		{"FdOrFindError", errors.New("fd error"), nil, false, 3},
		{"RunWithIndError", nil, errors.New("run error"), true, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindCalls = append(mock.RunFdOrFindCalls, RunFdOrFindCall{Err: tt.fdOrFindErr})
			}
			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
			}

			err := cleanHomeDirectory()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanHomeDirectory() error = %v, expectErr %v", err, tt.expectErr)
			}

			totalCalls := len(mock.RunFdOrFindCalls) + len(mock.RunWithIndicatorCalls)
			if totalCalls != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, totalCalls)
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[len(mock.RunFdOrFindCalls)-1]
				if call.Dir != "/home" || call.Args != "-type f \\( -name '*.tmp' -o -name '*.temp' -o -name '*.swp' -o -name '*~' \\) -delete" ||
					call.Message != "Removing temporary files in home directory..." || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[len(mock.RunWithIndicatorCalls)-1]
				if call.Command != "rm -rf /home/*/.cache/thumbnails/*" || call.Message != "Clearing thumbnail cache..." {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanUserCaches(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"Error", errors.New("fd error"), false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindCalls = append(mock.RunFdOrFindCalls, RunFdOrFindCall{Err: tt.fdOrFindErr})
			}

			err := cleanUserCaches()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUserCaches() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunFdOrFind, got %d", tt.expectedCalls, len(mock.RunFdOrFindCalls))
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[len(mock.RunFdOrFindCalls)-1]
				if call.Dir != "/home" || call.Args != "-type d -name '.cache' -exec rm -rf {}/* \\;" ||
					call.Message != "Clearing user caches..." || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}
		})
	}
}

func TestCleanUserTrash(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		fdOrFindErr   error
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, nil, false, 2},
		{"FdOrFindError", errors.New("fd error"), nil, false, 3},
		{"RunWithIndError", nil, errors.New("run error"), true, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindCalls = append(mock.RunFdOrFindCalls, RunFdOrFindCall{Err: tt.fdOrFindErr})
			}
			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
			}

			err := cleanUserTrash()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUserTrash() error = %v, expectErr %v", err, tt.expectErr)
			}

			totalCalls := len(mock.RunFdOrFindCalls) + len(mock.RunWithIndicatorCalls)
			if totalCalls != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, totalCalls)
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[len(mock.RunFdOrFindCalls)-1]
				if call.Dir != "/home" || call.Args != "-type d -name 'Trash' -exec rm -rf {}/* \\;" ||
					call.Message != "Emptying user trash folders..." || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}

			if len(mock.RunWithIndicatorCalls) > 0 {
				call := mock.RunWithIndicatorCalls[len(mock.RunWithIndicatorCalls)-1]
				if call.Command != "rm -rf /root/.local/share/Trash/*" || call.Message != "Emptying trash for root..." {
					t.Errorf("Unexpected arguments to RunWithIndicator: %+v", call)
				}
			}
		})
	}
}

func TestCleanUserHomeLogs(t *testing.T) {
	originalRunner := utils.Runner
	defer func() { utils.Runner = originalRunner }()

	tests := []struct {
		name          string
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"Error", errors.New("fd error"), false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindCalls = append(mock.RunFdOrFindCalls, RunFdOrFindCall{Err: tt.fdOrFindErr})
			}

			err := cleanUserHomeLogs()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUserHomeLogs() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunFdOrFindCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunFdOrFind, got %d", tt.expectedCalls, len(mock.RunFdOrFindCalls))
			}

			if len(mock.RunFdOrFindCalls) > 0 {
				call := mock.RunFdOrFindCalls[len(mock.RunFdOrFindCalls)-1]
				if call.Dir != "/home" || call.Args != "-type f -name '*.log' -size +10M -delete" ||
					call.Message != "Removing large log files in user home directories..." || !call.Sudo {
					t.Errorf("Unexpected arguments to RunFdOrFind: %+v", call)
				}
			}
		})
	}
}
