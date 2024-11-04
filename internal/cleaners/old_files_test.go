package cleaners

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindOldFiles(t *testing.T) {
	tests := []struct {
		name      string
		options   FileCleanupOptions
		mockFiles []struct {
			name    string
			size    int64
			modTime time.Time
		}
		expectedFiles int
		expectedError bool
	}{
		{
			name: "NoFiles",
			options: FileCleanupOptions{
				MinFileSize: 10 * 1024 * 1024,
				MaxFileAge:  365 * 24 * time.Hour,
			},
			mockFiles: []struct {
				name    string
				size    int64
				modTime time.Time
			}{},
			expectedFiles: 0,
			expectedError: false,
		},
		{
			name: "MixedFiles",
			options: FileCleanupOptions{
				MinFileSize: 10 * 1024 * 1024,
				MaxFileAge:  365 * 24 * time.Hour,
			},
			mockFiles: []struct {
				name    string
				size    int64
				modTime time.Time
			}{
				{"small_new.txt", 5 * 1024 * 1024, time.Now()},
				{"big_old.dat", 15 * 1024 * 1024, time.Now().AddDate(-2, 0, 0)},
				{"big_new.dat", 20 * 1024 * 1024, time.Now()},
			},
			expectedFiles: 1,
			expectedError: false,
		},
		{
			name: "CustomThresholds",
			options: FileCleanupOptions{
				MinFileSize: 5 * 1024 * 1024,
				MaxFileAge:  180 * 24 * time.Hour,
			},
			mockFiles: []struct {
				name    string
				size    int64
				modTime time.Time
			}{
				{"small_old.txt", 6 * 1024 * 1024, time.Now().AddDate(0, -7, 0)},
				{"big_old.dat", 15 * 1024 * 1024, time.Now().AddDate(-1, 0, 0)},
			},
			expectedFiles: 2,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir, err := os.MkdirTemp("", "old_files_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test files
			for _, tf := range tt.mockFiles {
				path := filepath.Join(tmpDir, tf.name)
				f, err := os.Create(path)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", tf.name, err)
				}
				f.Close()
				os.Chtimes(path, tf.modTime, tf.modTime)
			}

			// Override home directory for testing
			originalHomeDir := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", originalHomeDir)

			files, err := findOldFiles(tt.options)

			if (err != nil) != tt.expectedError {
				t.Errorf("findOldFiles() error = %v, expectedError %v", err, tt.expectedError)
			}

			if len(files) != tt.expectedFiles {
				t.Errorf("findOldFiles() got %d files, expected %d", len(files), tt.expectedFiles)
			}
		})
	}
}

func TestCleanOldFiles(t *testing.T) {
	tests := []struct {
		name      string
		testFiles []struct {
			name    string
			size    int64
			modTime time.Time
		}
		expectedDeleted int
	}{
		{
			name: "DeleteOldFiles",
			testFiles: []struct {
				name    string
				size    int64
				modTime time.Time
			}{
				{"big_old.dat", 15000000, time.Now().AddDate(-2, 0, 0)},
				{"small_old.dat", 5000000, time.Now().AddDate(-2, 0, 0)},
				{"big_new.dat", 15000000, time.Now()},
			},
			expectedDeleted: 1,
		},
		{
			name: "NoFilesToDelete",
			testFiles: []struct {
				name    string
				size    int64
				modTime time.Time
			}{
				{"big_new.dat", 15000000, time.Now()},
				{"small_new.dat", 5000000, time.Now()},
			},
			expectedDeleted: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test directory
			tmpDir, err := os.MkdirTemp("", "old_files_cleanup_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test files
			var testFilePaths []string
			for _, tf := range tt.testFiles {
				path := filepath.Join(tmpDir, tf.name)
				f, err := os.Create(path)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", tf.name, err)
				}
				f.Close()
				os.Chtimes(path, tf.modTime, tf.modTime)
				testFilePaths = append(testFilePaths, path)
			}

			// Override home directory for testing
			originalHomeDir := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", originalHomeDir)

			// Set test environment to automatically select files
			os.Setenv("GO_TEST_ENV", "true")
			defer os.Unsetenv("GO_TEST_ENV")

			err = cleanOldFiles()
			if err != nil {
				t.Errorf("cleanOldFiles() error = %v", err)
			}

			// Count remaining files
			remainingFiles, err := os.ReadDir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to read temp dir: %v", err)
			}

			expectedRemainingFiles := len(tt.testFiles) - tt.expectedDeleted
			if len(remainingFiles) != expectedRemainingFiles {
				t.Errorf("Expected %d remaining files, got %d", expectedRemainingFiles, len(remainingFiles))
			}
		})
	}
}
