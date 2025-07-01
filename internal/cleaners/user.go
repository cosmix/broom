package cleaners

import (
	"fmt"
	"os"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("home", Cleaner{CleanupFunc: cleanHomeDirectory, RequiresConfirmation: true})
	registerCleanup("cache", Cleaner{CleanupFunc: cleanUserCaches, RequiresConfirmation: true})
	registerCleanup("trash", Cleaner{CleanupFunc: cleanUserTrash, RequiresConfirmation: true})
	registerCleanup("user_logs", Cleaner{CleanupFunc: cleanUserHomeLogs, RequiresConfirmation: true})
}

func cleanHomeDirectory() error {
	// Target current user's home directory instead of all users for better performance
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}
	
	err := utils.Runner.RunFdOrFind(homeDir, "-type f \\( -name '*.tmp' -o -name '*.temp' -o -name '*.swp' -o -name '*~' \\) -delete", "Removing temporary files in home directory...", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing temporary files in home directory: %v\n", err)
	}
	
	// Also optimize the thumbnail cache cleanup to be user-specific
	thumbnailCmd := fmt.Sprintf("rm -rf %s/.cache/thumbnails/*", homeDir)
	return utils.Runner.RunWithIndicator(thumbnailCmd, "Clearing thumbnail cache...")
}

func cleanUserCaches() error {
	// Target current user's home directory for better performance
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}
	
	err := utils.Runner.RunFdOrFind(homeDir, "-type d -name '.cache' -exec rm -rf {}/* \\;", "Clearing user caches...", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing user caches: %v\n", err)
	}
	return nil
}

func cleanUserTrash() error {
	// Target current user's home directory for better performance
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}
	
	err := utils.Runner.RunFdOrFind(homeDir, "-type d -name 'Trash' -exec rm -rf {}/* \\;", "Emptying user trash folders...", true)
	if err != nil {
		fmt.Printf("Warning: Error while emptying user trash folders: %v\n", err)
	}
	
	// Also clean the user-specific trash location
	trashCmd := fmt.Sprintf("rm -rf %s/.local/share/Trash/*", homeDir)
	return utils.Runner.RunWithIndicator(trashCmd, "Emptying user trash...")
}

func cleanUserHomeLogs() error {
	// Target current user's home directory for better performance
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}
	
	err := utils.Runner.RunFdOrFind(homeDir, "-type f -name '*.log' -size +10M -delete", "Removing large log files in user home directories...", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing large log files in user home directories: %v\n", err)
	}
	return nil
}
