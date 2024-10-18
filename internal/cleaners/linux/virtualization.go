package linux

import (
	"fmt"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
)

func init() {
	cleaners.RegisterCleanup("virtualbox", cleaners.Cleaner{CleanupFunc: removeOldVirtualboxImages, RequiresConfirmation: false})
}

func removeOldVirtualboxImages() error {
	if utils.CommandExists("vboxmanage") {
		err := utils.Runner.RunWithIndicator("vboxmanage list hdds | grep -oP '(?<=UUID:).*' | xargs -I {} vboxmanage closemedium disk {} --delete", "Removing old VirtualBox disk images")
		if err != nil {
			fmt.Printf("Warning: Error while removing old VirtualBox disk images: %v\n", err)
		}
		return nil
	}
	fmt.Println("VirtualBox cleanup: Skipped (not installed)")
	return nil
}
