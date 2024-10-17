package cleaners

import (
	"fmt"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("docker", Cleaner{CleanupFunc: cleanDocker, RequiresConfirmation: true})
	registerCleanup("snap", Cleaner{CleanupFunc: cleanSnap, RequiresConfirmation: false})
	registerCleanup("flatpak", Cleaner{CleanupFunc: cleanFlatpak, RequiresConfirmation: false})
	registerCleanup("timeshift", Cleaner{CleanupFunc: cleanTimeshiftSnapshots, RequiresConfirmation: true})
	registerCleanup("ruby", Cleaner{CleanupFunc: cleanRubyGems, RequiresConfirmation: false})
	registerCleanup("python", Cleaner{CleanupFunc: cleanPythonCache, RequiresConfirmation: false})
	registerCleanup("libreoffice", Cleaner{CleanupFunc: cleanLibreOfficeCache, RequiresConfirmation: false})
	registerCleanup("browser", Cleaner{CleanupFunc: clearBrowserCaches, RequiresConfirmation: false})
	registerCleanup("package_manager", Cleaner{CleanupFunc: cleanPackageManagerCaches, RequiresConfirmation: false})
	registerCleanup("npm", Cleaner{CleanupFunc: cleanNpmCache, RequiresConfirmation: false})
	registerCleanup("gradle", Cleaner{CleanupFunc: cleanGradleCache, RequiresConfirmation: false})
	registerCleanup("composer", Cleaner{CleanupFunc: cleanComposerCache, RequiresConfirmation: false})
	registerCleanup("wine", Cleaner{CleanupFunc: removeOldWinePrefixes, RequiresConfirmation: false})
	registerCleanup("electron", Cleaner{CleanupFunc: cleanElectronCache, RequiresConfirmation: false})
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

func cleanRubyGems() error {
	if utils.CommandExists("gem") {
		return utils.Runner.RunWithIndicator("gem cleanup", "Removing old Ruby gems")
	}
	fmt.Println("Ruby gems cleanup: Skipped (not installed)")
	return nil
}

func cleanPythonCache() error {
	err := utils.Runner.RunFdOrFind("/home /tmp", "-type d -name __pycache__ -exec rm -rf {} +", "Removing Python cache files", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing Python cache files: %v\n", err)
	}
	err = utils.Runner.RunFdOrFind("/home /tmp", "-name '*.pyc' -delete", "Removing .pyc files", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing .pyc files: %v\n", err)
	}
	return nil
}

func cleanLibreOfficeCache() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -path '*/.config/libreoffice/4/user/uno_packages/cache' -exec rm -rf {}/* \\;", "Clearing LibreOffice cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing LibreOffice cache: %v\n", err)
	}
	return nil
}

func clearBrowserCaches() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -path '*/.cache/google-chrome/Default/Cache' -exec rm -rf {}/* \\;", "Clearing Chrome cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing Chrome cache: %v\n", err)
	}
	err = utils.Runner.RunFdOrFind("/home", "-type d -path '*/.cache/chromium/Default/Cache' -exec rm -rf {}/* \\;", "Clearing Chromium cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing Chromium cache: %v\n", err)
	}
	err = utils.Runner.RunFdOrFind("/home", "-type d -path '*/.mozilla/firefox/*/Cache' -exec rm -rf {}/* \\;", "Clearing Firefox cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing Firefox cache: %v\n", err)
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

func cleanNpmCache() error {
	if utils.CommandExists("npm") {
		return utils.Runner.RunWithIndicator("npm cache clean --force", "Cleaning npm cache")
	}
	fmt.Println("npm cache cleanup: Skipped (not installed)")
	return nil
}

func cleanGradleCache() error {
	return utils.Runner.RunWithIndicator("rm -rf $HOME/.gradle/caches", "Cleaning Gradle cache")
}

func cleanComposerCache() error {
	if utils.CommandExists("composer") {
		return utils.Runner.RunWithIndicator("composer clear-cache", "Cleaning Composer cache")
	}
	fmt.Println("Composer cache cleanup: Skipped (not installed)")
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

func cleanElectronCache() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -path '*/.config/*electron*' -exec rm -rf {}/* \\;", "Clearing Electron cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing Electron cache: %v\n", err)
	}
	return nil
}
