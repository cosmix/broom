package common

import "github.com/cosmix/broom/internal/cleaners"

var commonCleaners map[string]cleaners.Cleaner

func Init() {
	// Initialize common cleaners
	commonCleaners = map[string]cleaners.Cleaner{
		"python":   {CleanupFunc: cleanPythonCache, RequiresConfirmation: false},
		"browser":  {CleanupFunc: clearBrowserCaches, RequiresConfirmation: false},
		"npm":      {CleanupFunc: cleanNpmCache, RequiresConfirmation: false},
		"gradle":   {CleanupFunc: cleanGradleCache, RequiresConfirmation: false},
		"composer": {CleanupFunc: cleanComposerCache, RequiresConfirmation: false},
		"electron": {CleanupFunc: cleanElectronCache, RequiresConfirmation: false},
	}
}

func GetCleaners() map[string]cleaners.Cleaner {
	return commonCleaners
}
