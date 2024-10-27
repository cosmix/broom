package cleaners

import (
	"errors"
	"strings"
	"testing"
)

// FakeEnvironment represents a fake system environment for testing
type FakeEnvironment struct {
	InstalledKernels  []string
	InstalledPackages []string
	Files             map[string][]string // map of directory to files
}

func setupTestWithEnv() (*MockUtilsRunner, *FakeEnvironment) {
	mock := setupTest()
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
	mock, env := setupTestWithEnv()

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

func TestClearApt(t *testing.T) {
	mock, _ := setupTestWithEnv()

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
		case 3:
			if command != "apt-get clean" {
				t.Errorf("Unexpected command: %s", command)
			}
		default:
			t.Errorf("Unexpected call to RunWithIndicator")
		}
		return nil
	}

	err := clearApt()
	if err != nil {
		t.Errorf("clearApt() error = %v, wantErr %v", err, false)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls to RunWithIndicator, got %d", callCount)
	}
}

func TestRemoveOldLogs(t *testing.T) {
	mock, _ := setupTestWithEnv()

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
	mock, _ := setupTestWithEnv()

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
	mock, _ := setupTestWithEnv()

	expectedCalls := []struct {
		path string
		args string
	}{
		{"/tmp", "-type f -atime +10 -delete"},
		{"/var/tmp", "-type f -atime +10 -delete"},
	}

	callCount := 0
	mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
		if callCount >= len(expectedCalls) {
			t.Errorf("Unexpected call to RunFdOrFind: path=%s, args=%s", path, args)
			return nil
		}
		expected := expectedCalls[callCount]
		if path != expected.path || args != expected.args {
			t.Errorf("Unexpected RunFdOrFind call: got path=%s, args=%s; want path=%s, args=%s", path, args, expected.path, expected.args)
		}
		callCount++
		return nil
	}

	err := removeTemp()
	if err != nil {
		t.Errorf("removeTemp() error = %v, wantErr %v", err, false)
	}

	if callCount != len(expectedCalls) {
		t.Errorf("Expected %d calls to RunFdOrFind, got %d", len(expectedCalls), callCount)
	}
}

func TestCleanJournalLogs(t *testing.T) {
	mock, _ := setupTestWithEnv()

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
	mock, _ := setupTestWithEnv()

	testError := errors.New("test error")
	mock.RunWithIndicatorFunc = func(command, message string) error {
		return testError
	}

	err := removeOldKernels()
	if err != testError {
		t.Errorf("removeOldKernels() error = %v, wantErr %v", err, testError)
	}

	err = clearApt()
	if err != testError {
		t.Errorf("clearApt() error = %v, wantErr %v", err, testError)
	}

	err = cleanJournalLogs()
	if err != testError {
		t.Errorf("cleanJournalLogs() error = %v, wantErr %v", err, testError)
	}
}
