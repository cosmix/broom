package main

import (
	"runtime"
	"testing"

	"github.com/cosmix/broom/internal/cleaners/common"
	"github.com/cosmix/broom/internal/cleaners/linux"
	"github.com/cosmix/broom/internal/cleaners/macos"
)

func TestGetOSSpecificCleaners(t *testing.T) {
	cleaners := getOSSpecificCleaners()

	// Check if common cleaners are present
	for name := range common.GetCleaners() {
		if _, ok := cleaners[name]; !ok {
			t.Errorf("Common cleaner %s not found in OS-specific cleaners", name)
		}
	}

	// Check OS-specific cleaners
	switch runtime.GOOS {
	case "linux":
		for name := range linux.GetCleaners() {
			if _, ok := cleaners[name]; !ok {
				t.Errorf("Linux cleaner %s not found in OS-specific cleaners", name)
			}
		}
	case "darwin":
		for name := range macos.GetCleaners() {
			if _, ok := cleaners[name]; !ok {
				t.Errorf("macOS cleaner %s not found in OS-specific cleaners", name)
			}
		}
	default:
		t.Errorf("Unsupported operating system: %s", runtime.GOOS)
	}

	// Check that there are no unexpected cleaners
	expectedCount := len(common.GetCleaners())
	switch runtime.GOOS {
	case "linux":
		expectedCount += len(linux.GetCleaners())
	case "darwin":
		expectedCount += len(macos.GetCleaners())
	}

	if len(cleaners) != expectedCount {
		t.Errorf("Unexpected number of cleaners. Got %d, expected %d", len(cleaners), expectedCount)
	}
}

func TestCleanerStructure(t *testing.T) {
	cleaners := getOSSpecificCleaners()

	for name, cleaner := range cleaners {
		if cleaner.CleanupFunc == nil {
			t.Errorf("Cleaner %s has nil CleanupFunc", name)
		}
	}
}
