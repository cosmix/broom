package cleaners

import (
	"fmt"
	"strings"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("kernels", Cleaner{CleanupFunc: removeOldKernels, RequiresConfirmation: false})
	registerCleanup("apt", Cleaner{CleanupFunc: clearApt, RequiresConfirmation: false})
	registerCleanup("logs", Cleaner{CleanupFunc: removeOldLogs, RequiresConfirmation: true})
	registerCleanup("crash", Cleaner{CleanupFunc: removeCrashReports, RequiresConfirmation: true})
	registerCleanup("temp", Cleaner{CleanupFunc: removeTemp, RequiresConfirmation: false})
	registerCleanup("journal", Cleaner{CleanupFunc: cleanJournalLogs, RequiresConfirmation: true})
}

func removeOldKernels() error {
	// First, get the current kernel version
	currentKernel, err := utils.Runner.RunWithOutput("uname -r")
	if err != nil {
		return fmt.Errorf("failed to get current kernel version: %v", err)
	}
	currentKernel = strings.TrimSpace(currentKernel)
	
	// Get list of old kernels to remove (excluding current)
	cmd := fmt.Sprintf("dpkg --list | grep linux-image | grep -v %s | awk '{ print $2 }' | grep -E 'linux-image-[0-9]' | head -n -1", currentKernel)
	oldKernels, err := utils.Runner.RunWithOutput(cmd)
	if err != nil || strings.TrimSpace(oldKernels) == "" {
		fmt.Println("No old kernels to remove")
		return nil
	}
	
	// Remove old kernels
	kernelList := strings.TrimSpace(oldKernels)
	return utils.Runner.RunWithIndicator(fmt.Sprintf("apt-get -y purge %s", kernelList), "Removing old kernels...")
}

func clearApt() error {
	err := utils.Runner.RunWithIndicator("apt-get autoremove -y", "Removing unnecessary packages...")
	if err != nil {
		return err
	}
	err = utils.Runner.RunWithIndicator("apt-get purge -y nano vim-tiny", "Removing non-critical packages...")
	if err != nil {
		return err
	}
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
