package cleaners

import (
	"errors"
	"testing"

	"github.com/cosmix/broom/internal/utils"
)

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
	call := RunWithIndicatorCall{Command: command, Message: message}
	m.RunWithIndicatorCalls = append(m.RunWithIndicatorCalls, call)
	if len(m.RunWithIndicatorCalls) > 0 && m.RunWithIndicatorCalls[0].Err != nil {
		return m.RunWithIndicatorCalls[0].Err
	}
	return nil
}

func (m *MockRunner) RunWithOutput(command string) (string, error) {
	call := RunWithOutputCall{Command: command}
	m.RunWithOutputCalls = append(m.RunWithOutputCalls, call)
	if len(m.RunWithOutputCalls) > 0 {
		return m.RunWithOutputCalls[0].Output, m.RunWithOutputCalls[0].Err
	}
	return "", nil
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
		{"SteamInstalled", func(cmd string) bool { return true }, nil, false, 1},
		{"SteamNotInstalled", func(cmd string) bool { return false }, nil, false, 0},
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
		{"RunWithIndicatorError", func(cmd string) bool { return true }, errors.New("run error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
		{"FdOrFindError", func(cmd string) bool { return true }, errors.New("fd error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.fdOrFindErr != nil {
				mock.RunFdOrFindCalls = append(mock.RunFdOrFindCalls, RunFdOrFindCall{Err: tt.fdOrFindErr})
			}

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
		{"RunWithIndicatorError", errors.New("run error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
		{"SecondCommandError", func(cmd string) bool { return true }, errors.New("run error"), true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
				if tt.name == "SecondCommandError" {
					mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
				}
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
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"SDKManagerNotInstalled", func(cmd string) bool { return false }, "", nil, nil, false, 0},
		{"ListError", func(cmd string) bool { return true }, "", errors.New("list error"), nil, true, 0},
		{"NoPackagesToRemove", func(cmd string) bool { return true }, "package1\npackage2\n", nil, nil, false, 0},
		{"RemovePackages", func(cmd string) bool { return true }, "package1\nsystem-images;android-30;google_apis;x86\nemulator\n", nil, nil, false, 2},
		{"RemoveError", func(cmd string) bool { return true }, "system-images;android-30;google_apis;x86\n", nil, errors.New("remove error"), false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
			}

			cleanFunc := cleanAndroidSDK(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanAndroidSDK() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(mock.RunWithIndicatorCalls) != tt.expectedCalls {
				t.Errorf("Expected %d call(s) to RunWithIndicator, got %d", tt.expectedCalls, len(mock.RunWithIndicatorCalls))
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
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
				expectedCmd := "R -e \"remove.packages(installed.packages()[,1])\""
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
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
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
		withIndErr    error
		expectErr     bool
		expectedCalls int
	}{
		{"CondaNotInstalled", func(cmd string) bool { return false }, "", nil, nil, false, 0},
		{"ListError", func(cmd string) bool { return true }, "", errors.New("list error"), nil, true, 1},
		{"NoEnvironments", func(cmd string) bool { return true }, "{\"envs\": []}", nil, nil, false, 1},
		{"RemoveEnvironments", func(cmd string) bool { return true }, "{\"envs\": [\"base\", \"env1\", \"env2\"]}", nil, nil, false, 3},
		{"RemoveError", func(cmd string) bool { return true }, "{\"envs\": [\"env1\"]}", nil, errors.New("remove error"), false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRunner{}
			utils.Runner = mock

			mock.RunWithOutputCalls = append(mock.RunWithOutputCalls, RunWithOutputCall{Output: tt.execOutput, Err: tt.execErr})

			if tt.withIndErr != nil {
				mock.RunWithIndicatorCalls = append(mock.RunWithIndicatorCalls, RunWithIndicatorCall{Err: tt.withIndErr})
			}

			cleanFunc := cleanUnusedCondaEnvironments(tt.commandExists)
			err := cleanFunc()

			if (err != nil) != tt.expectErr {
				t.Errorf("cleanUnusedCondaEnvironments() error = %v, expectErr %v", err, tt.expectErr)
			}

			totalCalls := len(mock.RunWithOutputCalls) + len(mock.RunWithIndicatorCalls)
			if totalCalls != tt.expectedCalls {
				t.Errorf("Expected %d total call(s), got %d", tt.expectedCalls, totalCalls)
			}
		})
	}
}
