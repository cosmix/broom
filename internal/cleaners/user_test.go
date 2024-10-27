package cleaners

import (
	"errors"
	"testing"
)

func TestCleanHomeDirectory(t *testing.T) {
	tests := []struct {
		name          string
		fdOrFindErr   error
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, nil, false, 2},
		{"FdOrFindError", errors.New("fd error"), nil, false, 2},
		{"RunWithIndError", nil, errors.New("run error"), true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
					return tt.fdOrFindErr
				}
			}
			if tt.withIndErr != nil {
				mock.RunWithIndicatorFunc = func(command, message string) error {
					return tt.withIndErr
				}
			}

			err := cleanHomeDirectory()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanHomeDirectory() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.Commands) != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, len(mock.Commands))
			}

			if len(mock.Commands) > 0 {
				expectedFdCmd := "fd/find /home -type f \\( -name '*.tmp' -o -name '*.temp' -o -name '*.swp' -o -name '*~' \\) -delete"
				if mock.Commands[0] != expectedFdCmd {
					t.Errorf("Unexpected fd command: got %s, want %s", mock.Commands[0], expectedFdCmd)
				}

				if len(mock.Commands) > 1 {
					expectedRmCmd := "rm -rf /home/*/.cache/thumbnails/*"
					if mock.Commands[1] != expectedRmCmd {
						t.Errorf("Unexpected rm command: got %s, want %s", mock.Commands[1], expectedRmCmd)
					}
				}
			}
		})
	}
}

func TestCleanUserCaches(t *testing.T) {
	tests := []struct {
		name          string
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"Error", errors.New("fd error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
					return tt.fdOrFindErr
				}
			}

			err := cleanUserCaches()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUserCaches() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.Commands) != tt.expectedCalls {
				t.Errorf("Expected %d call(s), got %d", tt.expectedCalls, len(mock.Commands))
			}

			if len(mock.Commands) > 0 {
				expectedCmd := "fd/find /home -type d -name '.cache' -exec rm -rf {}/* \\;"
				if mock.Commands[0] != expectedCmd {
					t.Errorf("Unexpected command: got %s, want %s", mock.Commands[0], expectedCmd)
				}
			}
		})
	}
}

func TestCleanUserTrash(t *testing.T) {
	tests := []struct {
		name          string
		fdOrFindErr   error
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, nil, false, 2},
		{"FdOrFindError", errors.New("fd error"), nil, false, 2},
		{"RunWithIndError", nil, errors.New("run error"), true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
					return tt.fdOrFindErr
				}
			}
			if tt.withIndErr != nil {
				mock.RunWithIndicatorFunc = func(command, message string) error {
					return tt.withIndErr
				}
			}

			err := cleanUserTrash()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUserTrash() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.Commands) != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, len(mock.Commands))
			}

			if len(mock.Commands) > 0 {
				expectedFdCmd := "fd/find /home -type d -name 'Trash' -exec rm -rf {}/* \\;"
				if mock.Commands[0] != expectedFdCmd {
					t.Errorf("Unexpected fd command: got %s, want %s", mock.Commands[0], expectedFdCmd)
				}

				if len(mock.Commands) > 1 {
					expectedRmCmd := "rm -rf /root/.local/share/Trash/*"
					if mock.Commands[1] != expectedRmCmd {
						t.Errorf("Unexpected rm command: got %s, want %s", mock.Commands[1], expectedRmCmd)
					}
				}
			}
		})
	}
}

func TestCleanUserHomeLogs(t *testing.T) {
	tests := []struct {
		name          string
		fdOrFindErr   error
		expectErr     bool
		expectedCalls int
	}{
		{"Success", nil, false, 1},
		{"Error", errors.New("fd error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := setupTest()
			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindFunc = func(path, args, message string, sudo bool) error {
					return tt.fdOrFindErr
				}
			}

			err := cleanUserHomeLogs()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUserHomeLogs() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.Commands) != tt.expectedCalls {
				t.Errorf("Expected %d call(s), got %d", tt.expectedCalls, len(mock.Commands))
			}

			if len(mock.Commands) > 0 {
				expectedCmd := "fd/find /home -type f -name '*.log' -size +10M -delete"
				if mock.Commands[0] != expectedCmd {
					t.Errorf("Unexpected command: got %s, want %s", mock.Commands[0], expectedCmd)
				}
			}
		})
	}
}
