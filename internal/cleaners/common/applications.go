package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Warning: Unable to get user home directory: %v\n", err)
		return ""
	}
	return home
}

func cleanPythonCache() error {
	homeDir := getHomeDir()
	if homeDir == "" {
		return fmt.Errorf("unable to determine home directory")
	}

	searchPaths := fmt.Sprintf("%s /tmp", homeDir)
	err := utils.Runner.RunFdOrFind(searchPaths, "-type d -name __pycache__ -exec rm -rf {} + -o -name '*.pyc' -delete", "Removing Python cache files and .pyc files", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing Python cache files and .pyc files: %v\n", err)
	}
	return nil
}

func clearBrowserCaches() error {
	homeDir := getHomeDir()
	if homeDir == "" {
		return fmt.Errorf("unable to determine home directory")
	}

	browserPaths := map[string]string{
		"chrome":   filepath.Join(homeDir, ".cache", "google-chrome", "Default", "Cache"),
		"chromium": filepath.Join(homeDir, ".cache", "chromium", "Default", "Cache"),
		"firefox":  filepath.Join(homeDir, ".mozilla", "firefox"),
	}

	if runtime.GOOS == "darwin" {
		browserPaths = map[string]string{
			"chrome":   filepath.Join(homeDir, "Library", "Caches", "Google", "Chrome"),
			"chromium": filepath.Join(homeDir, "Library", "Caches", "Chromium"),
			"firefox":  filepath.Join(homeDir, "Library", "Caches", "Firefox"),
		}
	}

	for browser, path := range browserPaths {
		if browser == "firefox" {
			err := utils.Runner.RunFdOrFind(path, "-type d -name Cache -exec rm -rf {}/* \\;", fmt.Sprintf("Clearing %s cache", browser), true)
			if err != nil {
				fmt.Printf("Warning: Error while clearing %s cache: %v\n", browser, err)
			}
		} else {
			err := utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", path), fmt.Sprintf("Clearing %s cache", browser))
			if err != nil {
				fmt.Printf("Warning: Error while clearing %s cache: %v\n", browser, err)
			}
		}
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
	homeDir := getHomeDir()
	if homeDir == "" {
		return fmt.Errorf("unable to determine home directory")
	}
	gradleCachePath := filepath.Join(homeDir, ".gradle", "caches")
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s", gradleCachePath), "Cleaning Gradle cache")
}

func cleanComposerCache() error {
	if utils.CommandExists("composer") {
		return utils.Runner.RunWithIndicator("composer clear-cache", "Cleaning Composer cache")
	}
	fmt.Println("Composer cache cleanup: Skipped (not installed)")
	return nil
}

func cleanElectronCache() error {
	homeDir := getHomeDir()
	if homeDir == "" {
		return fmt.Errorf("unable to determine home directory")
	}

	electronPath := filepath.Join(homeDir, ".config")
	if runtime.GOOS == "darwin" {
		electronPath = filepath.Join(homeDir, "Library", "Application Support")
	}

	err := utils.Runner.RunFdOrFind(electronPath, "-type d -path '*electron*' -exec rm -rf {}/* \\;", "Clearing Electron cache", true)
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
