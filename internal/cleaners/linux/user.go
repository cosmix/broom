package linux

import (
	"fmt"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("home", cleaners.Cleaner{CleanupFunc: cleanHomeDirectory, RequiresConfirmation: true})
	cleaners.RegisterCleanup("cache", cleaners.Cleaner{CleanupFunc: cleanUserCaches, RequiresConfirmation: true})
	cleaners.RegisterCleanup("trash", cleaners.Cleaner{CleanupFunc: cleanUserTrash, RequiresConfirmation: true})
	cleaners.RegisterCleanup("user_logs", cleaners.Cleaner{CleanupFunc: cleanUserHomeLogs, RequiresConfirmation: true})
}

func cleanHomeDirectory() error {
	err := utils.Runner.RunFdOrFind("/home", "-type f \\( -name '*.tmp' -o -name '*.temp' -o -name '*.swp' -o -name '*~' \\) -delete", "Removing temporary files in home directory...", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing temporary files in home directory: %v\n", err)
	}
	return utils.Runner.RunWithIndicator("rm -rf /home/*/.cache/thumbnails/*", "Clearing thumbnail cache...")
}

func cleanUserCaches() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -name '.cache' -exec rm -rf {}/* \\;", "Clearing user caches...", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing user caches: %v\n", err)
	}
	return nil
}

func cleanUserTrash() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -name 'Trash' -exec rm -rf {}/* \\;", "Emptying user trash folders...", true)
	if err != nil {
		fmt.Printf("Warning: Error while emptying user trash folders: %v\n", err)
	}
	return utils.Runner.RunWithIndicator("rm -rf /root/.local/share/Trash/*", "Emptying trash for root...")
}

func cleanUserHomeLogs() error {
	err := utils.Runner.RunFdOrFind("/home", "-type f -name '*.log' -size +10M -delete", "Removing large log files in user home directories...", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing large log files in user home directories: %v\n", err)
	}
	return nil
}