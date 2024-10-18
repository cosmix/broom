package macos

import "github.com/cosmix/broom/internal/cleaners"

var macosCleaners map[string]cleaners.Cleaner

func Init() {
	// Initialize macOS-specific cleaners
	macosCleaners = map[string]cleaners.Cleaner{
		"system_cache": {CleanupFunc: cleanSystemCache, RequiresConfirmation: false},
		"system_logs":  {CleanupFunc: cleanSystemLogs, RequiresConfirmation: false},
		"xcode_cache":  {CleanupFunc: cleanXcodeCache, RequiresConfirmation: false},
	}
}

func GetCleaners() map[string]cleaners.Cleaner {
	return macosCleaners
}
