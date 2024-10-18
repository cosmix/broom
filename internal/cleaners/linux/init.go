package linux

import "github.com/cosmix/broom/internal/cleaners"

var linuxCleaners map[string]cleaners.Cleaner

func Init() {
	// Initialize Linux-specific cleaners
	linuxCleaners = map[string]cleaners.Cleaner{
		"old_kernels":   {CleanupFunc: removeOldKernels, RequiresConfirmation: true},
		"apt":           {CleanupFunc: clearApt, RequiresConfirmation: false},
		"old_logs":      {CleanupFunc: removeOldLogs, RequiresConfirmation: false},
		"crash_reports": {CleanupFunc: removeCrashReports, RequiresConfirmation: false},
		"temp":          {CleanupFunc: removeTemp, RequiresConfirmation: false},
		"journal_logs":  {CleanupFunc: cleanJournalLogs, RequiresConfirmation: false},
	}
}

func GetCleaners() map[string]cleaners.Cleaner {
	return linuxCleaners
}
