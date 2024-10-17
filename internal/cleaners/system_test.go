package cleaners

import (
	"errors"
	"strings"
	"testing"

	"github.com/cosmix/broom/internal/utils"
)

// MockUtilsRunner implements utils.UtilsRunner for testing
type MockUtilsRunner struct {
	RunWithIndicatorFunc func(command, message string) error
	RunFdOrFindFunc      func(path, args, message string, sudo bool) error
	Commands             []string
}

func (m *MockUtilsRunner) RunWithIndicator(command, message string) error {
	m.Commands = append(m.Commands, command)
	return m.RunWithIndicatorFunc(command, message)
}

func (m *MockUtilsRunner) RunFdOrFind(path, args, message string, sudo bool) error {
	command := "fd/find " + path + " " + args
	m.Commands = append(m.Commands, command)
	return m.RunFdOrFindFunc(path, args, message, sudo)
}

// FakeEnvironment represents a fake system environment for testing
type FakeEnvironment struct {
	InstalledKernels  []string
	InstalledPackages []string
	Files             map[string][]string // map of directory to files
}

func setupTest() (*MockUtilsRunner, *FakeEnvironment) {
	mock := &MockUtilsRunner{
		RunWithIndicatorFunc: func(command, message string) error { return nil },
		RunFdOrFindFunc:      func(path, args, message string, sudo bool) error { return nil },
	}
	utils.SetUtilsRunner(mock)

	env := &FakeEnvironment{
		InstalledKernels:  []string{"linux-image-5.4.0-42-generic", "linux-image-5.4.0-45-generic", "linux-image-5.4.0-47-generic"},
		InstalledPackages: []string{"nano", "vim-tiny", "other-package"},
		Files: map[string][]string{
			"/var/log":                  {"system.log", "auth.log"},
			"/var/crash":                {"crash1.log", "crash2.log"},
			"/var/lib/systemd/coredump": {"core1", "core2"},
			"/tmp":                      {"old_file", "new_file"},
			"/var/tmp":                  {"old_file", "new_file"},
		},
	}

	return mock, env
}

func TestRemoveOldKernels(t *testing.T) {
	mock, env := setupTest()

	mock.RunWithIndicatorFunc = func(command, message string) error {
		if !strings.Contains(command, "dpkg --list | grep linux-image") {
			t.Errorf("Unexpected command: %s", command)
		}
		return nil
	}

	err := removeOldKernels()
	if err != nil {
		t.Errorf("removeOldKernels() error = %v, wantErr %v", err, false)
	}

	if len(mock.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mock.Commands))
	}

	// Check if the current kernel is not removed
	if strings.Contains(mock.Commands[0], env.InstalledKernels[len(env.InstalledKernels)-1]) {
		t.Errorf("Current kernel should not be removed: %s", mock.Commands[0])
	}
}

func TestRemoveUnnecessaryPackages(t *testing.T) {
	mock, _ := setupTest()

	callCount := 0
	mock.RunWithIndicatorFunc = func(command, message string) error {
		callCount++
		switch callCount {
		case 1:
			if command != "apt-get autoremove -y" {
				t.Errorf("Unexpected command: %s", command)
			}
		case 2:
			if command != "apt-get purge -y nano vim-tiny" {
				t.Errorf("Unexpected command: %s", command)
			}
		default:
			t.Errorf("Unexpected call to RunWithIndicator")
		}
		return nil
	}

	err := removeUnnecessaryPackages()
	if err != nil {
		t.Errorf("removeUnnecessaryPackages() error = %v, wantErr %v", err, false)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls to RunWithIndicator, got %d", callCount)
	}
}

func TestClearAptCache(t *testing.T) {
	mock, _ := setupTest()

	mock.RunWithIndicatorFunc = func(command, message string) error {
		if command != "apt-get clean" {
			t.Errorf("Unexpected command: %s", command)
		}
		return nil
	}

	err := clearAptCache()
	if err != nil {
		t.Errorf("clearAptCache() error = %v, wantErr %v", err, false)
	}

	if len(mock.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mock.Commands))
	}
}

func TestRemoveOldLogs(t *testing.T) {
	mock, _ := setupTest()

	callCount := 0
	mock.RunWithIndicatorFunc = func(command, message string) error {
		if command != "journalctl --vacuum-time=3d" {
			t.Errorf("Unexpected command: %s", command)
		}
		callCount++
		return nil
	}

	mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
		if path != "/var/log" || !strings.Contains(args, "-type f -name \"*.log\" -mtime +30 -delete") {
			t.Errorf("Unexpected RunFdOrFind call: path=%s, args=%s", path, args)
		}
		callCount++
		return nil
	}

	err := removeOldLogs()
	if err != nil {
		t.Errorf("removeOldLogs() error = %v, wantErr %v", err, false)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls (RunWithIndicator and RunFdOrFind), got %d", callCount)
	}
}

func TestRemoveCrashReports(t *testing.T) {
	mock, _ := setupTest()

	callCount := 0
	mock.RunWithIndicatorFunc = func(command, message string) error {
		if command != "rm -rf /var/crash/*" {
			t.Errorf("Unexpected command: %s", command)
		}
		callCount++
		return nil
	}

	mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
		if path != "/var/lib/systemd/coredump" || args != "-type f -delete" {
			t.Errorf("Unexpected RunFdOrFind call: path=%s, args=%s", path, args)
		}
		callCount++
		return nil
	}

	err := removeCrashReports()
	if err != nil {
		t.Errorf("removeCrashReports() error = %v, wantErr %v", err, false)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls (RunWithIndicator and RunFdOrFind), got %d", callCount)
	}
}

func TestRemoveTemp(t *testing.T) {
	mock, _ := setupTest()

	callCount := 0
	mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
		callCount++
		if (path != "/tmp" && path != "/var/tmp") || args != "-type f -atime +10 -delete" {
			t.Errorf("Unexpected RunFdOrFind call: path=%s, args=%s", path, args)
		}
		return nil
	}

	err := removeTemp()
	if err != nil {
		t.Errorf("removeTemp() error = %v, wantErr %v", err, false)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls to RunFdOrFind, got %d", callCount)
	}
}

func TestCleanJournalLogs(t *testing.T) {
	mock, _ := setupTest()

	mock.RunWithIndicatorFunc = func(command, message string) error {
		if command != "journalctl --vacuum-size=100M" {
			t.Errorf("Unexpected command: %s", command)
		}
		return nil
	}

	err := cleanJournalLogs()
	if err != nil {
		t.Errorf("cleanJournalLogs() error = %v, wantErr %v", err, false)
	}

	if len(mock.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mock.Commands))
	}
}

func TestErrorHandling(t *testing.T) {
	mock, _ := setupTest()

	testError := errors.New("test error")
	mock.RunWithIndicatorFunc = func(command, message string) error {
		return testError
	}

	err := removeOldKernels()
	if err != testError {
		t.Errorf("removeOldKernels() error = %v, wantErr %v", err, testError)
	}

	err = removeUnnecessaryPackages()
	if err != testError {
		t.Errorf("removeUnnecessaryPackages() error = %v, wantErr %v", err, testError)
	}

	err = clearAptCache()
	if err != testError {
		t.Errorf("clearAptCache() error = %v, wantErr %v", err, testError)
	}

	err = cleanJournalLogs()
	if err != testError {
		t.Errorf("cleanJournalLogs() error = %v, wantErr %v", err, testError)
	}
}
