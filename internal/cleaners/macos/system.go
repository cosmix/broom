package macos

import (
	"fmt"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("macOSSystemCache", cleaners.Cleaner{CleanupFunc: cleanSystemCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("macOSSystemLogs", cleaners.Cleaner{CleanupFunc: cleanSystemLogs, RequiresConfirmation: false})
	cleaners.RegisterCleanup("macOSXcodeCache", cleaners.Cleaner{CleanupFunc: cleanXcodeCache, RequiresConfirmation: false})
}

func cleanSystemCache() error {
	err := utils.Runner.RunWithIndicator("sudo rm -rf /Library/Caches/*", "Cleaning system cache")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning system cache: %v\n", err)
	}
	return nil
}

func cleanSystemLogs() error {
	err := utils.Runner.RunWithIndicator("sudo rm -rf /var/log/*", "Cleaning system logs")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning system logs: %v\n", err)
	}
	return nil
}

func cleanXcodeCache() error {
	err := utils.Runner.RunWithIndicator("rm -rf ~/Library/Developer/Xcode/DerivedData", "Cleaning Xcode derived data")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning Xcode derived data: %v\n", err)
	}
	return nil
}
