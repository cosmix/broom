package main

import (
	"runtime"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/cleaners/common"
	"github.com/cosmix/broom/internal/cleaners/linux"
	"github.com/cosmix/broom/internal/cleaners/macos"
)

var osSpecificCleaners map[string]cleaners.Cleaner

func init() {
	osSpecificCleaners = make(map[string]cleaners.Cleaner)

	// Initialize common cleaners
	for name, cleaner := range common.GetCleaners() {
		osSpecificCleaners[name] = cleaner
	}

	// Initialize OS-specific cleaners
	switch runtime.GOOS {
	case "linux":
		for name, cleaner := range linux.GetCleaners() {
			osSpecificCleaners[name] = cleaner
		}
	case "darwin":
		for name, cleaner := range macos.GetCleaners() {
			osSpecificCleaners[name] = cleaner
		}
	default:
		panic("Unsupported operating system")
	}
}

func getOSSpecificCleaners() map[string]cleaners.Cleaner {
	return osSpecificCleaners
}
