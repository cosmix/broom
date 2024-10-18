package common

import (
	"fmt"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("python", cleaners.Cleaner{CleanupFunc: cleanPythonCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("browser", cleaners.Cleaner{CleanupFunc: clearBrowserCaches, RequiresConfirmation: false})
	cleaners.RegisterCleanup("npm", cleaners.Cleaner{CleanupFunc: cleanNpmCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("gradle", cleaners.Cleaner{CleanupFunc: cleanGradleCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("composer", cleaners.Cleaner{CleanupFunc: cleanComposerCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("electron", cleaners.Cleaner{CleanupFunc: cleanElectronCache, RequiresConfirmation: false})
	cleaners.RegisterCleanup("ruby", cleaners.Cleaner{CleanupFunc: cleanRubyGems, RequiresConfirmation: false})
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

func cleanElectronCache() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -path '*/.config/*electron*' -exec rm -rf {}/* \\;", "Clearing Electron cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing Electron cache: %v\n", err)
	}
	return nil
}

func cleanRubyGems() error {
	if utils.CommandExists("gem") {
		return utils.Runner.RunWithIndicator("gem cleanup", "Removing old Ruby gems")
	}
	fmt.Println("Ruby gems cleanup: Skipped (not installed)")
	return nil
}
