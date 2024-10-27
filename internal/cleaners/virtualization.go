package cleaners

import (
	"fmt"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("virtualbox", Cleaner{CleanupFunc: removeOldVirtualboxImages, RequiresConfirmation: true})
	registerCleanup("lxc_lxd", Cleaner{CleanupFunc: cleanLXCLXD, RequiresConfirmation: true})
	registerCleanup("podman", Cleaner{CleanupFunc: cleanPodman, RequiresConfirmation: true})
	registerCleanup("vagrant", Cleaner{CleanupFunc: cleanVagrant, RequiresConfirmation: true})
	registerCleanup("buildah", Cleaner{CleanupFunc: cleanBuildah, RequiresConfirmation: true})
}

func removeOldVirtualboxImages() error {
	return removeOldVirtualboxImagesWithCheck(utils.CommandExists)
}

func removeOldVirtualboxImagesWithCheck(commandExists utils.CommandExistsFunc) error {
	if commandExists("vboxmanage") {
		err := utils.Runner.RunFdOrFind("$HOME/VirtualBox VMs", "-type f -name '*.vdi' -mtime +90 -delete", "Removing old Virtualbox disk images...", true)
		if err != nil {
			fmt.Printf("Warning: Error while removing old Virtualbox disk images: %v\n", err)
		}
		return nil
	}
	fmt.Println("Virtualbox is not installed. Skipping Virtualbox disk images cleanup.")
	return nil
}

func cleanLXCLXD() error {
	return cleanLXCLXDWithCheck(utils.CommandExists)
}

func cleanLXCLXDWithCheck(commandExists utils.CommandExistsFunc) error {
	if commandExists("lxc") {
		err := utils.Runner.RunWithIndicator("lxc image list --format csv | cut -d',' -f1 | xargs -I {} lxc image delete {}", "Removing unused LXC/LXD images...")
		if err != nil {
			fmt.Printf("Warning: Error while removing unused LXC/LXD images: %v\n", err)
		}
		err = utils.Runner.RunWithIndicator("lxc list --format csv | cut -d',' -f1 | xargs -I {} lxc delete --force {}", "Removing unused LXC/LXD containers...")
		if err != nil {
			fmt.Printf("Warning: Error while removing unused LXC/LXD containers: %v\n", err)
		}
		return nil
	}
	fmt.Println("LXC/LXD is not installed. Skipping LXC/LXD cleanup.")
	return nil
}

func cleanPodman() error {
	return cleanPodmanWithCheck(utils.CommandExists)
}

func cleanPodmanWithCheck(commandExists utils.CommandExistsFunc) error {
	if commandExists("podman") {
		err := utils.Runner.RunWithIndicator("podman image prune -af", "Removing unused Podman images...")
		if err != nil {
			fmt.Printf("Warning: Error while removing unused Podman images: %v\n", err)
		}
		err = utils.Runner.RunWithIndicator("podman container prune -f", "Removing unused Podman containers...")
		if err != nil {
			fmt.Printf("Warning: Error while removing unused Podman containers: %v\n", err)
		}
		return nil
	}
	fmt.Println("Podman is not installed. Skipping Podman cleanup.")
	return nil
}

func cleanVagrant() error {
	return cleanVagrantWithCheck(utils.CommandExists)
}

func cleanVagrantWithCheck(commandExists utils.CommandExistsFunc) error {
	if commandExists("vagrant") {
		err := utils.Runner.RunWithIndicator("vagrant global-status --prune", "Pruning invalid Vagrant entries...")
		if err != nil {
			fmt.Printf("Warning: Error while pruning invalid Vagrant entries: %v\n", err)
		}
		err = utils.Runner.RunWithIndicator("rm -rf ~/.vagrant.d/boxes/*", "Removing Vagrant box cache...")
		if err != nil {
			fmt.Printf("Warning: Error while removing Vagrant box cache: %v\n", err)
		}
		return nil
	}
	fmt.Println("Vagrant is not installed. Skipping Vagrant cleanup.")
	return nil
}

func cleanBuildah() error {
	return cleanBuildahWithCheck(utils.CommandExists)
}

func cleanBuildahWithCheck(commandExists utils.CommandExistsFunc) error {
	if commandExists("buildah") {
		err := utils.Runner.RunWithIndicator("buildah rmi --all", "Removing dangling Buildah images...")
		if err != nil {
			fmt.Printf("Warning: Error while removing dangling Buildah images: %v\n", err)
		}
		return nil
	}
	fmt.Println("Buildah is not installed. Skipping Buildah cleanup.")
	return nil
}
