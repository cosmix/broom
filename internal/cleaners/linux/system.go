package linux

import (
	"fmt"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("old_kernels", cleaners.Cleaner{CleanupFunc: removeOldKernels, RequiresConfirmation: true})
	cleaners.RegisterCleanup("apt", cleaners.Cleaner{CleanupFunc: clearApt, RequiresConfirmation: false})
	cleaners.RegisterCleanup("old_logs", cleaners.Cleaner{CleanupFunc: removeOldLogs, RequiresConfirmation: false})
	cleaners.RegisterCleanup("crash_reports", cleaners.Cleaner{CleanupFunc: removeCrashReports, RequiresConfirmation: false})
	cleaners.RegisterCleanup("temp", cleaners.Cleaner{CleanupFunc: removeTemp, RequiresConfirmation: false})
	cleaners.RegisterCleanup("journal_logs", cleaners.Cleaner{CleanupFunc: cleanJournalLogs, RequiresConfirmation: false})
}

func removeOldKernels() error {
	return utils.Runner.RunWithIndicator("dpkg -l 'linux-*' | sed '/^ii/!d;/'\"`uname -r | sed \"s/\\(.*\\)-\\([^0-9]\\+\\)/\\1/\"`\"'/d;s/^[^ ]* [^ ]* \\([^ ]*\\).*/\\1/;/[0-9]/!d' | xargs sudo apt-get -y purge", "Removing old kernels")
}

func clearApt() error {
	return utils.Runner.RunWithIndicator("sudo apt-get clean", "Clearing APT cache")
}

func removeOldLogs() error {
	return utils.Runner.RunFdOrFind("/var/log", "-type f -name '*.log' -mtime +30 -delete", "Removing old log files", true)
}

func removeCrashReports() error {
	return utils.Runner.RunFdOrFind("/var/crash", "-type f -delete", "Removing crash reports", true)
}

func removeTemp() error {
	err := utils.Runner.RunFdOrFind("/tmp", "-type f -atime +10 -delete", "Removing old temporary files", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing old temporary files: %v\n", err)
	}
	err = utils.Runner.RunFdOrFind("/var/tmp", "-type f -atime +10 -delete", "Removing old files in /var/tmp", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing old files in /var/tmp: %v\n", err)
	}
	return nil
}

func cleanJournalLogs() error {
	return utils.Runner.RunWithIndicator("sudo journalctl --vacuum-time=30d", "Cleaning old journal logs")
}
