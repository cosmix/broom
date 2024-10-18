package linux

import (
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
}

type RunWithIndicatorCall struct {
	Command string
	Message string
}

func (m *MockRunner) RunFdOrFind(dir, args, message string, sudo bool) error {
	m.RunFdOrFindCalls = append(m.RunFdOrFindCalls, RunFdOrFindCall{dir, args, message, sudo})
	return nil
}

func (m *MockRunner) RunWithIndicator(command, message string) error {
	m.RunWithIndicatorCalls = append(m.RunWithIndicatorCalls, RunWithIndicatorCall{command, message})
	return nil
}

func TestCleanHomeDirectory(t *testing.T) {
	mock := &MockRunner{}
	utils.Runner = mock

	err := cleanHomeDirectory()
	if err != nil {
		t.Errorf("cleanHomeDirectory() returned an error: %v", err)
	}

	expectedFdOrFindCalls := []RunFdOrFindCall{
		{
			Dir:     "/home",
			Args:    "-type f \\( -name '*.tmp' -o -name '*.temp' -o -name '*.swp' -o -name '*~' \\) -delete",
			Message: "Removing temporary files in home directory...",
			Sudo:    true,
		},
	}

	if len(mock.RunFdOrFindCalls) != len(expectedFdOrFindCalls) {
		t.Errorf("Expected %d RunFdOrFind calls, got %d", len(expectedFdOrFindCalls), len(mock.RunFdOrFindCalls))
	}

	for i, call := range mock.RunFdOrFindCalls {
		if call != expectedFdOrFindCalls[i] {
			t.Errorf("Unexpected RunFdOrFind call %d: got %+v, want %+v", i, call, expectedFdOrFindCalls[i])
		}
	}

	expectedRunWithIndicatorCalls := []RunWithIndicatorCall{
		{
			Command: "rm -rf /home/*/.cache/thumbnails/*",
			Message: "Clearing thumbnail cache...",
		},
	}

	if len(mock.RunWithIndicatorCalls) != len(expectedRunWithIndicatorCalls) {
		t.Errorf("Expected %d RunWithIndicator calls, got %d", len(expectedRunWithIndicatorCalls), len(mock.RunWithIndicatorCalls))
	}

	for i, call := range mock.RunWithIndicatorCalls {
		if call != expectedRunWithIndicatorCalls[i] {
			t.Errorf("Unexpected RunWithIndicator call %d: got %+v, want %+v", i, call, expectedRunWithIndicatorCalls[i])
		}
	}
}

func TestCleanUserCaches(t *testing.T) {
	mock := &MockRunner{}
	utils.Runner = mock

	err := cleanUserCaches()
	if err != nil {
		t.Errorf("cleanUserCaches() returned an error: %v", err)
	}

	expectedCalls := []RunFdOrFindCall{
		{
			Dir:     "/home",
			Args:    "-type d -name '.cache' -exec rm -rf {}/* \\;",
			Message: "Clearing user caches...",
			Sudo:    true,
		},
	}

	if len(mock.RunFdOrFindCalls) != len(expectedCalls) {
		t.Errorf("Expected %d RunFdOrFind calls, got %d", len(expectedCalls), len(mock.RunFdOrFindCalls))
	}

	for i, call := range mock.RunFdOrFindCalls {
		if call != expectedCalls[i] {
			t.Errorf("Unexpected RunFdOrFind call %d: got %+v, want %+v", i, call, expectedCalls[i])
		}
	}
}

func TestCleanUserTrash(t *testing.T) {
	mock := &MockRunner{}
	utils.Runner = mock

	err := cleanUserTrash()
	if err != nil {
		t.Errorf("cleanUserTrash() returned an error: %v", err)
	}

	expectedFdOrFindCalls := []RunFdOrFindCall{
		{
			Dir:     "/home",
			Args:    "-type d -name 'Trash' -exec rm -rf {}/* \\;",
			Message: "Emptying user trash folders...",
			Sudo:    true,
		},
	}

	if len(mock.RunFdOrFindCalls) != len(expectedFdOrFindCalls) {
		t.Errorf("Expected %d RunFdOrFind calls, got %d", len(expectedFdOrFindCalls), len(mock.RunFdOrFindCalls))
	}

	for i, call := range mock.RunFdOrFindCalls {
		if call != expectedFdOrFindCalls[i] {
			t.Errorf("Unexpected RunFdOrFind call %d: got %+v, want %+v", i, call, expectedFdOrFindCalls[i])
		}
	}

	expectedRunWithIndicatorCalls := []RunWithIndicatorCall{
		{
			Command: "rm -rf /root/.local/share/Trash/*",
			Message: "Emptying trash for root...",
		},
	}

	if len(mock.RunWithIndicatorCalls) != len(expectedRunWithIndicatorCalls) {
		t.Errorf("Expected %d RunWithIndicator calls, got %d", len(expectedRunWithIndicatorCalls), len(mock.RunWithIndicatorCalls))
	}

	for i, call := range mock.RunWithIndicatorCalls {
		if call != expectedRunWithIndicatorCalls[i] {
			t.Errorf("Unexpected RunWithIndicator call %d: got %+v, want %+v", i, call, expectedRunWithIndicatorCalls[i])
		}
	}
}

func TestCleanUserHomeLogs(t *testing.T) {
	mock := &MockRunner{}
	utils.Runner = mock

	err := cleanUserHomeLogs()
	if err != nil {
		t.Errorf("cleanUserHomeLogs() returned an error: %v", err)
	}

	expectedCalls := []RunFdOrFindCall{
		{
			Dir:     "/home",
			Args:    "-type f -name '*.log' -size +10M -delete",
			Message: "Removing large log files in user home directories...",
			Sudo:    true,
		},
	}

	if len(mock.RunFdOrFindCalls) != len(expectedCalls) {
		t.Errorf("Expected %d RunFdOrFind calls, got %d", len(expectedCalls), len(mock.RunFdOrFindCalls))
	}

	for i, call := range mock.RunFdOrFindCalls {
		if call != expectedCalls[i] {
			t.Errorf("Unexpected RunFdOrFind call %d: got %+v, want %+v", i, call, expectedCalls[i])
		}
	}
}
