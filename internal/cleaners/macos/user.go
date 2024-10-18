package macos

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("macOSUserCaches", cleaners.Cleaner{
		CleanupFunc:          cleanMacOSUserCaches,
		RequiresConfirmation: false,
	})
	cleaners.RegisterCleanup("macOSUserLogs", cleaners.Cleaner{
		CleanupFunc:          cleanMacOSUserLogs,
		RequiresConfirmation: true,
	})
}

func cleanMacOSUserCaches() error {
	cacheDir := filepath.Join(os.Getenv("HOME"), "Library/Caches")
	err := utils.RunFdOrFind(cacheDir, "-type f -delete", fmt.Sprintf("Cleaning %s", cacheDir), false)
	if err != nil {
		return fmt.Errorf("error cleaning %s: %w", cacheDir, err)
	}
	return nil
}

func cleanMacOSUserLogs() error {
	logDir := filepath.Join(os.Getenv("HOME"), "Library/Logs")
	err := utils.RunFdOrFind(logDir, "-type f -mtime +30 -delete", fmt.Sprintf("Removing old logs from %s", logDir), true)
	if err != nil {
		return fmt.Errorf("error cleaning logs in %s: %w", logDir, err)
	}
	return nil
}
