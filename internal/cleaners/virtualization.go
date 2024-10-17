package cleaners

import (
	"fmt"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("virtualbox", Cleaner{CleanupFunc: removeOldVirtualboxImages, RequiresConfirmation: false})
}

func removeOldVirtualboxImages() error {
	if utils.CommandExists("vboxmanage") {
		err := utils.Runner.RunFdOrFind("$HOME/VirtualBox VMs", "-type f -name '*.vdi' -mtime +90 -delete", "Removing old Virtualbox disk images...", true)
		if err != nil {
			fmt.Printf("Warning: Error while removing old Virtualbox disk images: %v\n", err)
		}
		return nil
	}
	fmt.Println("Virtualbox is not installed. Skipping Virtualbox disk images cleanup.")
	return nil
}
