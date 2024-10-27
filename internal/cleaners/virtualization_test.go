package cleaners

import (
	"errors"
	"testing"
)

func TestCleanLXCLXD(t *testing.T) {
	tests := []struct {
		name          string
		commandExists bool
		withIndErr    error
		expectedCalls int
	}{
		{"LXCInstalled", true, nil, 2},
		{"LXCNotInstalled", false, nil, 0},
		{"FirstCommandError", true, errors.New("run error"), 2},
		{"SecondCommandError", true, errors.New("run error"), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			commandExists := func(command string) bool {
				return tt.commandExists
			}

			callCount := 0
			mock.RunWithIndicatorFunc = func(command, message string) error {
				if !tt.commandExists {
					t.Errorf("RunWithIndicator called when command doesn't exist")
				}
				callCount++
				return tt.withIndErr
			}

			cleanLXCLXDWithCheck(commandExists)

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, callCount)
			}

			expectedCommands := []string{
				"lxc image list --format csv | cut -d',' -f1 | xargs -I {} lxc image delete {}",
				"lxc list --format csv | cut -d',' -f1 | xargs -I {} lxc delete --force {}",
			}

			for i, cmd := range mock.Commands {
				if i < len(expectedCommands) && i < callCount && cmd != expectedCommands[i] {
					t.Errorf("Unexpected command: got %s, want %s", cmd, expectedCommands[i])
				}
			}
		})
	}
}

func TestCleanPodman(t *testing.T) {
	tests := []struct {
		name          string
		commandExists bool
		withIndErr    error
		expectedCalls int
	}{
		{"PodmanInstalled", true, nil, 2},
		{"PodmanNotInstalled", false, nil, 0},
		{"FirstCommandError", true, errors.New("run error"), 2},
		{"SecondCommandError", true, errors.New("run error"), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			commandExists := func(command string) bool {
				return tt.commandExists
			}

			callCount := 0
			mock.RunWithIndicatorFunc = func(command, message string) error {
				if !tt.commandExists {
					t.Errorf("RunWithIndicator called when command doesn't exist")
				}
				callCount++
				return tt.withIndErr
			}

			cleanPodmanWithCheck(commandExists)

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, callCount)
			}

			expectedCommands := []string{
				"podman image prune -af",
				"podman container prune -f",
			}

			for i, cmd := range mock.Commands {
				if i < len(expectedCommands) && i < callCount && cmd != expectedCommands[i] {
					t.Errorf("Unexpected command: got %s, want %s", cmd, expectedCommands[i])
				}
			}
		})
	}
}

func TestCleanVagrant(t *testing.T) {
	tests := []struct {
		name          string
		commandExists bool
		withIndErr    error
		expectedCalls int
	}{
		{"VagrantInstalled", true, nil, 2},
		{"VagrantNotInstalled", false, nil, 0},
		{"FirstCommandError", true, errors.New("run error"), 2},
		{"SecondCommandError", true, errors.New("run error"), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			commandExists := func(command string) bool {
				return tt.commandExists
			}

			callCount := 0
			mock.RunWithIndicatorFunc = func(command, message string) error {
				if !tt.commandExists {
					t.Errorf("RunWithIndicator called when command doesn't exist")
				}
				callCount++
				return tt.withIndErr
			}

			cleanVagrantWithCheck(commandExists)

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, callCount)
			}

			expectedCommands := []string{
				"vagrant global-status --prune",
				"rm -rf ~/.vagrant.d/boxes/*",
			}

			for i, cmd := range mock.Commands {
				if i < len(expectedCommands) && i < callCount && cmd != expectedCommands[i] {
					t.Errorf("Unexpected command: got %s, want %s", cmd, expectedCommands[i])
				}
			}
		})
	}
}

func TestCleanBuildah(t *testing.T) {
	tests := []struct {
		name          string
		commandExists bool
		withIndErr    error
		expectedCalls int
	}{
		{"BuildahInstalled", true, nil, 1},
		{"BuildahNotInstalled", false, nil, 0},
		{"CommandError", true, errors.New("run error"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			commandExists := func(command string) bool {
				return tt.commandExists
			}

			callCount := 0
			mock.RunWithIndicatorFunc = func(command, message string) error {
				if !tt.commandExists {
					t.Errorf("RunWithIndicator called when command doesn't exist")
				}
				callCount++
				return tt.withIndErr
			}

			cleanBuildahWithCheck(commandExists)

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, callCount)
			}

			if callCount > 0 {
				expectedCommand := "buildah rmi --all"
				if mock.Commands[0] != expectedCommand {
					t.Errorf("Unexpected command: got %s, want %s", mock.Commands[0], expectedCommand)
				}
			}
		})
	}
}
