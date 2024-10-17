package cleaners

import (
	"fmt"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("kernels", Cleaner{CleanupFunc: removeOldKernels, RequiresConfirmation: true})
	registerCleanup("packages", Cleaner{CleanupFunc: removeUnnecessaryPackages, RequiresConfirmation: true})
	registerCleanup("apt", Cleaner{CleanupFunc: clearAptCache, RequiresConfirmation: true})
	registerCleanup("logs", Cleaner{CleanupFunc: removeOldLogs, RequiresConfirmation: true})
	registerCleanup("crash", Cleaner{CleanupFunc: removeCrashReports, RequiresConfirmation: true})
	registerCleanup("temp", Cleaner{CleanupFunc: removeTemp, RequiresConfirmation: true})
	registerCleanup("journal", Cleaner{CleanupFunc: cleanJournalLogs, RequiresConfirmation: true})
}

func removeOldKernels() error {
	return utils.Runner.RunWithIndicator("dpkg --list | grep linux-image | awk '{ print $2 }' | sort -V | sed -n '/'`uname -r`'/q;p' | xargs sudo apt-get -y purge", "Removing old kernels...")
}

func removeUnnecessaryPackages() error {
	err := utils.Runner.RunWithIndicator("apt-get autoremove -y", "Removing unnecessary packages...")
	if err != nil {
		return err
	}
	return utils.Runner.RunWithIndicator("apt-get purge -y nano vim-tiny", "Removing non-critical packages...")
}

func clearAptCache() error {
	return utils.Runner.RunWithIndicator("apt-get clean", "Clearing APT cache...")
}

func removeOldLogs() error {
	err := utils.Runner.RunWithIndicator("journalctl --vacuum-time=3d", "Clearing old journal logs...")
	if err != nil {
		return err
	}
	return utils.Runner.RunFdOrFind("/var/log", "-type f -name \"*.log\" -mtime +30 -delete", "Removing old log files...", false)
}

func removeCrashReports() error {
	err := utils.Runner.RunWithIndicator("rm -rf /var/crash/*", "Removing crash reports...")
	if err != nil {
		return err
	}
	return utils.Runner.RunFdOrFind("/var/lib/systemd/coredump", "-type f -delete", "Removing core dumps...", false)
}

func removeTemp() error {
	commands := []struct {
		path string
		args string
		msg  string
	}{
		{"/tmp", "-type f -atime +10 -delete", "Removing old files in /tmp..."},
		{"/var/tmp", "-type f -atime +10 -delete", "Removing old files in /var/tmp..."},
	}

	for _, command := range commands {
		err := utils.Runner.RunFdOrFind(command.path, command.args, command.msg, true)
		if err != nil {
			fmt.Printf("Warning: %s: %v\n", command.msg, err)
		}
	}
	return nil
}

func cleanJournalLogs() error {
	return utils.Runner.RunWithIndicator("journalctl --vacuum-size=100M", "Limiting journal size to 100MB...")
}
