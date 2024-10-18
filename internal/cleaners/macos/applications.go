package macos

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("macOSTrash", cleaners.Cleaner{
		CleanupFunc:          emptyMacOSTrash,
		RequiresConfirmation: true,
	})
}

func emptyMacOSTrash() error {
	trashDir := filepath.Join(os.Getenv("HOME"), ".Trash")
	err := utils.RunFdOrFind(trashDir, "-type f -delete", "Emptying Trash", false)
	if err != nil {
		return fmt.Errorf("error emptying Trash: %w", err)
	}
	return nil
}
