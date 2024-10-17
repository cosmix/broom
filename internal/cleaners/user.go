package cleaners

import (
	"fmt"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("home", Cleaner{CleanupFunc: cleanHomeDirectory, RequiresConfirmation: true})
	registerCleanup("cache", Cleaner{CleanupFunc: cleanUserCaches, RequiresConfirmation: true})
	registerCleanup("trash", Cleaner{CleanupFunc: cleanUserTrash, RequiresConfirmation: true})
	registerCleanup("user_logs", Cleaner{CleanupFunc: cleanUserHomeLogs, RequiresConfirmation: true})
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
