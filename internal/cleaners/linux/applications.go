package linux

import (
	"fmt"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("docker", cleaners.Cleaner{CleanupFunc: cleanDocker, RequiresConfirmation: true})
	cleaners.RegisterCleanup("snap", cleaners.Cleaner{CleanupFunc: cleanSnap, RequiresConfirmation: false})
	cleaners.RegisterCleanup("flatpak", cleaners.Cleaner{CleanupFunc: cleanFlatpak, RequiresConfirmation: false})
	cleaners.RegisterCleanup("timeshift", cleaners.Cleaner{CleanupFunc: cleanTimeshiftSnapshots, RequiresConfirmation: true})
	cleaners.RegisterCleanup("libreoffice", cleaners.Cleaner{CleanupFunc: cleanLibreOfficeCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("package_manager", cleaners.Cleaner{CleanupFunc: cleanPackageManagerCaches, RequiresConfirmation: false})
	cleaners.RegisterCleanup("wine", cleaners.Cleaner{CleanupFunc: removeOldWinePrefixes, RequiresConfirmation: false})
}

func cleanDocker() error {
	if utils.CommandExists("docker") {
		return utils.Runner.RunWithIndicator("docker system prune -af", "Removing unused Docker data")
	}
	fmt.Println("Docker cleanup: Skipped (not installed)")
	return nil
}

func cleanSnap() error {
	if utils.CommandExists("snap") {
		err := utils.Runner.RunWithIndicator("snap list --all | awk '/disabled/{print $1, $3}' | while read snapname revision; do sudo snap remove $snapname --revision=$revision; done", "Removing old snap versions")
		if err != nil {
			return err
		}
		return utils.Runner.RunWithIndicator("rm -rf /var/lib/snapd/cache/*", "Clearing snap cache")
	}
	fmt.Println("Snap cleanup: Skipped (not installed)")
	return nil
}

func cleanFlatpak() error {
	if utils.CommandExists("flatpak") {
		return utils.Runner.RunWithIndicator("flatpak uninstall --unused -y", "Removing unused Flatpak runtimes")
	}
	fmt.Println("Flatpak cleanup: Skipped (not installed)")
	return nil
}

func cleanTimeshiftSnapshots() error {
	if utils.CommandExists("timeshift") {
		return utils.Runner.RunWithIndicator("timeshift --list | grep -oP '(?<=\\s)\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}' | sort | head -n -3 | xargs -I {} timeshift --delete --snapshot '{}'", "Removing old Timeshift snapshots")
	}
	fmt.Println("Timeshift cleanup: Skipped (not installed)")
	return nil
}

func cleanLibreOfficeCache() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -path '*/.config/libreoffice/4/user/uno_packages/cache' -exec rm -rf {}/* \\;", "Clearing LibreOffice cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing LibreOffice cache: %v\n", err)
	}
	return nil
}

func cleanPackageManagerCaches() error {
	if utils.CommandExists("apt-get") {
		err := utils.Runner.RunWithIndicator("apt-get clean", "Cleaning APT cache")
		if err != nil {
			return err
		}
	}
	if utils.CommandExists("yum") {
		err := utils.Runner.RunWithIndicator("yum clean all", "Cleaning YUM cache")
		if err != nil {
			return err
		}
	}
	if utils.CommandExists("dnf") {
		return utils.Runner.RunWithIndicator("dnf clean all", "Cleaning DNF cache")
	}
	return nil
}

func removeOldWinePrefixes() error {
	if utils.CommandExists("wine") {
		err := utils.Runner.RunFdOrFind("$HOME", "-maxdepth 0 -type d -name '.wine*' -mtime +90 -exec rm -rf {} +", "Removing old Wine prefixes", true)
		if err != nil {
			fmt.Printf("Warning: Error while removing old Wine prefixes: %v\n", err)
		}
		return nil
	}
	fmt.Println("Wine prefixes cleanup: Skipped (not installed)")
	return nil
}
