// Package cleaners provides application-specific cleanup functions for broom.
package cleaners

import (
	"fmt"
	"strings"

	"github.com/cosmix/broom/internal/utils"
)

func init() {
	registerCleanup("docker", Cleaner{CleanupFunc: cleanDocker(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("snap", Cleaner{CleanupFunc: cleanSnap(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("flatpak", Cleaner{CleanupFunc: cleanFlatpak(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("timeshift", Cleaner{CleanupFunc: cleanTimeshiftSnapshots(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("ruby", Cleaner{CleanupFunc: cleanRubyGems(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("python", Cleaner{CleanupFunc: cleanPythonCache, RequiresConfirmation: false})
	registerCleanup("libreoffice", Cleaner{CleanupFunc: cleanLibreOfficeCache, RequiresConfirmation: false})
	registerCleanup("browser", Cleaner{CleanupFunc: clearBrowserCaches, RequiresConfirmation: false})
	registerCleanup("package_manager", Cleaner{CleanupFunc: cleanPackageManagerCaches(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("npm", Cleaner{CleanupFunc: cleanNpmCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("yarn", Cleaner{CleanupFunc: cleanYarnCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("pnpm", Cleaner{CleanupFunc: cleanPnpmCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("deno", Cleaner{CleanupFunc: cleanDenoCache, RequiresConfirmation: false})
	registerCleanup("bun", Cleaner{CleanupFunc: cleanBunCache, RequiresConfirmation: false})
	registerCleanup("pip", Cleaner{CleanupFunc: cleanPipCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("poetry", Cleaner{CleanupFunc: cleanPoetryCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("pipenv", Cleaner{CleanupFunc: cleanPipenvCache, RequiresConfirmation: false})
	registerCleanup("uv", Cleaner{CleanupFunc: cleanUvCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("gradle", Cleaner{CleanupFunc: cleanGradleCache, RequiresConfirmation: false})
	registerCleanup("composer", Cleaner{CleanupFunc: cleanComposerCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("wine", Cleaner{CleanupFunc: removeOldWinePrefixes(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("electron", Cleaner{CleanupFunc: cleanElectronCache, RequiresConfirmation: false})
	registerCleanup("kdenlive", Cleaner{CleanupFunc: cleanKdenliveRenderFiles(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("blender", Cleaner{CleanupFunc: cleanBlenderTempFiles(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("steam", Cleaner{CleanupFunc: cleanSteamDownloadCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("mysql_mariadb", Cleaner{CleanupFunc: cleanMySQLMariaDBBinlogs(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("thunderbird", Cleaner{CleanupFunc: cleanThunderbirdCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("dropbox", Cleaner{CleanupFunc: cleanDropboxCache, RequiresConfirmation: true})
	registerCleanup("maven", Cleaner{CleanupFunc: cleanMavenCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("go", Cleaner{CleanupFunc: cleanGoCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("rust", Cleaner{CleanupFunc: cleanRustCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("android", Cleaner{CleanupFunc: cleanAndroidSDK(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("jetbrains", Cleaner{CleanupFunc: cleanJetBrainsIDECaches(), RequiresConfirmation: true})
	registerCleanup("r_packages", Cleaner{CleanupFunc: cleanRPackagesCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("julia_packages", Cleaner{CleanupFunc: cleanJuliaPackagesCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("conda", Cleaner{CleanupFunc: cleanUnusedCondaEnvironments(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("mercurial", Cleaner{CleanupFunc: cleanMercurialBackups(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("git_lfs", Cleaner{CleanupFunc: cleanGitLFSCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("cmake", Cleaner{CleanupFunc: cleanCMakeBuildDirs, RequiresConfirmation: true})
	registerCleanup("autotools", Cleaner{CleanupFunc: cleanAutotoolsFiles, RequiresConfirmation: true})
	registerCleanup("ccache", Cleaner{CleanupFunc: cleanCCache(utils.CommandExists), RequiresConfirmation: true})
	// Phase 2: Modern DevOps Tools
	registerCleanup("kubectl", Cleaner{CleanupFunc: cleanKubectlCache, RequiresConfirmation: false})
	registerCleanup("helm", Cleaner{CleanupFunc: cleanHelmCache, RequiresConfirmation: false})
	registerCleanup("minikube", Cleaner{CleanupFunc: cleanMinikubeCache(utils.CommandExists), RequiresConfirmation: false})
	registerCleanup("terraform", Cleaner{CleanupFunc: cleanTerraformCache, RequiresConfirmation: false})
	registerCleanup("ansible", Cleaner{CleanupFunc: cleanAnsibleTemp, RequiresConfirmation: false})
	registerCleanup("containerd", Cleaner{CleanupFunc: cleanContainerdCache(utils.CommandExists), RequiresConfirmation: true})
	registerCleanup("podman_system", Cleaner{CleanupFunc: cleanPodmanSystem(utils.CommandExists), RequiresConfirmation: true})
}

func cleanDocker(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("docker") {
			return utils.Runner.RunWithIndicator("docker system prune -af", "Removing unused Docker data")
		}
		fmt.Println("Docker cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanSnap(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("snap") {
			err := utils.Runner.RunWithIndicator("snap list --all | awk '/disabled/{print $1, $3}' | while read snapname revision; do sudo snap remove $snapname --revision=$revision; done", "Removing old snap versions")
			if err != nil {
				return err
			}
			return utils.Runner.RunWithIndicator("rm -rf /var/lib/snapd/cache/*", "Clearing snap cache")
		}
		fmt.Println("Snap cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanFlatpak(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("flatpak") {
			return utils.Runner.RunWithIndicator("flatpak uninstall --unused -y", "Removing unused Flatpak runtimes")
		}
		fmt.Println("Flatpak cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanTimeshiftSnapshots(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("timeshift") {
			return utils.Runner.RunWithIndicator("timeshift --list | grep -oP '(?<=\\s)\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}' | sort | head -n -3 | xargs -I {} timeshift --delete --snapshot '{}'", "Removing old Timeshift snapshots")
		}
		fmt.Println("Timeshift cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanRubyGems(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("gem") {
			return utils.Runner.RunWithIndicator("gem cleanup", "Removing old Ruby gems")
		}
		fmt.Println("Ruby gems cleanup: Skipped (not installed)")
		return nil
	}
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

func cleanPackageManagerCaches(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("apt-get") {
			err := utils.Runner.RunWithIndicator("apt-get clean", "Cleaning APT cache")
			if err != nil {
				return err
			}
		}
		if commandExists("yum") {
			err := utils.Runner.RunWithIndicator("yum clean all", "Cleaning YUM cache")
			if err != nil {
				return err
			}
		}
		if commandExists("dnf") {
			return utils.Runner.RunWithIndicator("dnf clean all", "Cleaning DNF cache")
		}
		return nil
	}
}

func cleanNpmCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("npm") {
			return utils.Runner.RunWithIndicator("npm cache clean --force", "Cleaning npm cache")
		}
		fmt.Println("npm cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanYarnCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("yarn") {
			return utils.Runner.RunWithIndicator("yarn cache clean", "Cleaning yarn cache")
		}
		fmt.Println("yarn cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanPnpmCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("pnpm") {
			return utils.Runner.RunWithIndicator("pnpm store prune", "Cleaning pnpm store")
		}
		fmt.Println("pnpm cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanDenoCache() error {
	cacheDir := "$HOME/.cache/deno"
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", cacheDir), "Cleaning Deno cache")
}

func cleanBunCache() error {
	cacheDir := "$HOME/.bun/install/cache"
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", cacheDir), "Cleaning Bun cache")
}

func cleanPipCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("pip") {
			return utils.Runner.RunWithIndicator("pip cache purge", "Cleaning pip cache")
		}
		fmt.Println("pip cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanPoetryCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("poetry") {
			return utils.Runner.RunWithIndicator("poetry cache clear . --all", "Cleaning poetry cache")
		}
		fmt.Println("poetry cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanPipenvCache() error {
	cacheDir := "$HOME/.cache/pipenv"
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", cacheDir), "Cleaning pipenv cache")
}

func cleanUvCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("uv") {
			return utils.Runner.RunWithIndicator("uv cache clean", "Cleaning uv cache")
		}
		fmt.Println("uv cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanGradleCache() error {
	return utils.Runner.RunWithIndicator("rm -rf $HOME/.gradle/caches", "Cleaning Gradle cache")
}

func cleanComposerCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("composer") {
			return utils.Runner.RunWithIndicator("composer clear-cache", "Cleaning Composer cache")
		}
		fmt.Println("Composer cache cleanup: Skipped (not installed)")
		return nil
	}
}

func removeOldWinePrefixes(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("wine") {
			err := utils.Runner.RunFdOrFind("$HOME", "-maxdepth 0 -type d -name '.wine*' -mtime +90 -exec rm -rf {} +", "Removing old Wine prefixes", true)
			if err != nil {
				fmt.Printf("Warning: Error while removing old Wine prefixes: %v\n", err)
			}
			return nil
		}
		fmt.Println("Wine prefixes cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanElectronCache() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -path '*/.config/*electron*' -exec rm -rf {}/* \\;", "Clearing Electron cache", true)
	if err != nil {
		fmt.Printf("Warning: Error while clearing Electron cache: %v\n", err)
	}
	return nil
}

func cleanKdenliveRenderFiles(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("kdenlive") {
			return utils.Runner.RunFdOrFind("$HOME", "-type f -path '*/kdenlive/render/*' -delete", "Removing Kdenlive render files", true)
		}
		fmt.Println("Kdenlive cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanBlenderTempFiles(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("blender") {
			return utils.Runner.RunFdOrFind("$HOME", "-type f -path '*/blender_*_autosave.blend' -delete", "Removing Blender temporary files", true)
		}
		fmt.Println("Blender cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanSteamDownloadCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("steam") {
			steamPath := "$HOME/.steam/steam/steamapps/downloading"
			return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", steamPath), "Clearing Steam download cache")
		}
		fmt.Println("Steam cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanMySQLMariaDBBinlogs(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("mysql") || commandExists("mariadb") {
			cmd := `mysql -e "PURGE BINARY LOGS BEFORE DATE(NOW() - INTERVAL 7 DAY);"`
			err := utils.Runner.RunWithIndicator(cmd, "Removing old MySQL/MariaDB binary logs")
			if err != nil {
				fmt.Println("Note: This command may require database admin privileges.")
			}
			return err
		}
		fmt.Println("MySQL/MariaDB cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanThunderbirdCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("thunderbird") {
			return utils.Runner.RunFdOrFind("$HOME/.thunderbird", "-type d -name 'Cache' -exec rm -rf {}/* \\;", "Clearing Thunderbird cache", true)
		}
		fmt.Println("Thunderbird cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanDropboxCache() error {
	dropboxCachePath := "$HOME/.dropbox/cache"
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", dropboxCachePath), "Clearing Dropbox cache")
}

func cleanMavenCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("mvn") {
			return utils.Runner.RunWithIndicator("rm -rf ~/.m2/repository", "Cleaning Maven local repository cache...")
		}
		fmt.Println("Maven cache cleanup: Skipped (Maven not installed)")
		return nil
	}
}

func cleanGoCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("go") {
			return utils.Runner.RunWithIndicator("go clean -modcache", "Cleaning old Go modules cache...")
		}
		fmt.Println("Go cache cleanup: Skipped (Go not installed)")
		return nil
	}
}

func cleanRustCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("cargo") {
			err := utils.Runner.RunWithIndicator("rm -rf ~/.cargo/registry", "Cleaning Rust cargo registry...")
			if err != nil {
				return fmt.Errorf("failed to clean Rust cargo registry: %v", err)
			}
			err = utils.Runner.RunWithIndicator("rm -rf ~/.cargo/git", "Cleaning Rust cargo git cache...")
			if err != nil {
				return fmt.Errorf("failed to clean Rust cargo git cache: %v", err)
			}
			return nil
		}
		fmt.Println("Rust cache cleanup: Skipped (Rust not installed)")
		return nil
	}
}

func cleanAndroidSDK(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("sdkmanager") {
			fmt.Println("Android SDK cleanup: Skipped (sdkmanager not installed)")
			return nil
		}

		output, err := utils.Runner.RunWithOutput("sdkmanager --list_installed")
		if err != nil {
			return fmt.Errorf("failed to list installed Android SDK packages: %v", err)
		}

		installedPackages := strings.Split(output, "\n")
		for _, pkg := range installedPackages {
			if strings.Contains(pkg, "system-images") || strings.Contains(pkg, "emulator") {
				packageName := strings.Fields(pkg)[0]
				err := utils.Runner.RunWithIndicator(fmt.Sprintf("sdkmanager --uninstall %s", packageName), fmt.Sprintf("Removing Android SDK package: %s", packageName))
				if err != nil {
					fmt.Printf("Warning: Failed to remove Android SDK package %s: %v\n", packageName, err)
				}
			}
		}

		return nil
	}
}

func cleanJetBrainsIDECaches() func() error {
	return func() error {
		jetbrainsDir := "~/.local/share/JetBrains"
		err := utils.Runner.RunWithIndicator(fmt.Sprintf("find %s -type d -name '.caches' -exec rm -rf {} +", jetbrainsDir), "Cleaning JetBrains IDE caches")
		if err != nil {
			return fmt.Errorf("failed to clean JetBrains IDE caches: %v", err)
		}
		return nil
	}
}

func cleanRPackagesCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("R") {
			fmt.Println("R packages cache cleanup: Skipped (R not installed)")
			return nil
		}

		cmd := "R -e \"remove.packages(installed.packages()[,1])\""
		err := utils.Runner.RunWithIndicator(cmd, "Cleaning R packages cache")
		if err != nil {
			return fmt.Errorf("failed to clean R packages cache: %v", err)
		}
		return nil
	}
}

func cleanJuliaPackagesCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("julia") {
			fmt.Println("Julia packages cache cleanup: Skipped (Julia not installed)")
			return nil
		}

		cmd := "julia -e 'using Pkg; Pkg.gc()'"
		err := utils.Runner.RunWithIndicator(cmd, "Cleaning Julia packages cache")
		if err != nil {
			return fmt.Errorf("failed to clean Julia packages cache: %v", err)
		}
		return nil
	}
}

func cleanUnusedCondaEnvironments(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("conda") {
			fmt.Println("Conda environments cleanup: Skipped (conda not installed)")
			return nil
		}

		output, err := utils.Runner.RunWithOutput("conda env list --json")
		if err != nil {
			return fmt.Errorf("failed to list Conda environments: %v", err)
		}

		// Parse JSON output to get environment names
		// For simplicity, we'll just use string manipulation here
		envs := strings.Split(output, "\n")
		for _, env := range envs {
			if strings.Contains(env, "envs") {
				envName := strings.Trim(strings.Split(env, ":")[0], " \t\"")
				if envName != "base" { // Don't remove the base environment
					err := utils.Runner.RunWithIndicator(fmt.Sprintf("conda env remove --name %s", envName), fmt.Sprintf("Removing Conda environment: %s", envName))
					if err != nil {
						fmt.Printf("Warning: Failed to remove Conda environment %s: %v\n", envName, err)
					}
				}
			}
		}

		return nil
	}
}
func cleanMercurialBackups(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("hg") {
			fmt.Println("Mercurial cleanup: Skipped (not installed)")
			return nil
		}

		err := utils.Runner.RunFdOrFind("/home", "-type f -name '*.hg*.bak' -delete", "Removing Mercurial backup files", true)
		if err != nil {
			fmt.Printf("Warning: Error while removing Mercurial backup files: %v\n", err)
		}

		bundlesPath := "$HOME/.hg/bundle-backup"
		err = utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", bundlesPath), "Removing Mercurial bundle backups")
		if err != nil {
			fmt.Printf("Warning: Error while removing Mercurial bundle backups: %v\n", err)
		}

		return nil
	}
}

func cleanGitLFSCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("git-lfs") {
			fmt.Println("Git LFS cleanup: Skipped (not installed)")
			return nil
		}

		err := utils.Runner.RunWithIndicator("git lfs prune", "Cleaning Git LFS cache")
		if err != nil {
			return fmt.Errorf("failed to clean Git LFS cache: %v", err)
		}

		return nil
	}
}

func cleanCMakeBuildDirs() error {
	err := utils.Runner.RunFdOrFind("/home", "-type d -name 'build' -exec test -f '{}/CMakeCache.txt' \\; -exec rm -rf {} \\;", "Removing old CMake build directories", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing CMake build directories: %v\n", err)
	}

	err = utils.Runner.RunFdOrFind("/home", "-type d -name 'CMakeFiles' -exec rm -rf {} \\;", "Removing CMakeFiles directories", true)
	if err != nil {
		fmt.Printf("Warning: Error while removing CMakeFiles directories: %v\n", err)
	}

	return nil
}

func cleanAutotoolsFiles() error {
	patterns := []string{
		"autom4te.cache",
		"config.status",
		"config.log",
		"configure~",
		"Makefile.in",
		"aclocal.m4",
		".deps",
		".libs",
	}

	for _, pattern := range patterns {
		err := utils.Runner.RunFdOrFind("/home", fmt.Sprintf("-type d,f -name '%s' -exec rm -rf {} \\;", pattern), fmt.Sprintf("Removing Autotools generated %s", pattern), true)
		if err != nil {
			fmt.Printf("Warning: Error while removing Autotools %s: %v\n", pattern, err)
		}
	}

	return nil
}

func cleanCCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if !commandExists("ccache") {
			fmt.Println("ccache cleanup: Skipped (not installed)")
			return nil
		}

		err := utils.Runner.RunWithIndicator("ccache -C", "Clearing ccache")
		if err != nil {
			return fmt.Errorf("failed to clear ccache: %v", err)
		}

		return nil
	}
}

// Phase 2: Modern DevOps Tools

func cleanKubectlCache() error {
	kubeCacheDir := "$HOME/.kube/cache"
	err := utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", kubeCacheDir), "Cleaning kubectl cache")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning kubectl cache: %v\n", err)
	}

	kubeHTTPCacheDir := "$HOME/.kube/http-cache"
	err = utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", kubeHTTPCacheDir), "Cleaning kubectl HTTP cache")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning kubectl HTTP cache: %v\n", err)
	}

	return nil
}

func cleanHelmCache() error {
	helmCacheDir := "$HOME/.cache/helm"
	err := utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", helmCacheDir), "Cleaning Helm cache")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning Helm cache: %v\n", err)
	}

	helmDataDir := "$HOME/.local/share/helm"
	err = utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", helmDataDir), "Cleaning Helm data")
	if err != nil {
		fmt.Printf("Warning: Error while cleaning Helm data: %v\n", err)
	}

	return nil
}

func cleanMinikubeCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("minikube") {
			minikubeCacheDir := "$HOME/.minikube/cache"
			return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", minikubeCacheDir), "Cleaning minikube cache")
		}
		fmt.Println("minikube cache cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanTerraformCache() error {
	terraformCacheDir := "$HOME/.terraform.d/plugin-cache"
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", terraformCacheDir), "Cleaning Terraform plugin cache")
}

func cleanAnsibleTemp() error {
	ansibleTempDir := "$HOME/.ansible/tmp"
	return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", ansibleTempDir), "Cleaning Ansible temporary files")
}

func cleanContainerdCache(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("containerd") {
			containerdPath := "/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs"
			return utils.Runner.RunWithIndicator(fmt.Sprintf("rm -rf %s/*", containerdPath), "Cleaning containerd cache")
		}
		fmt.Println("containerd cleanup: Skipped (not installed)")
		return nil
	}
}

func cleanPodmanSystem(commandExists utils.CommandExistsFunc) func() error {
	return func() error {
		if commandExists("podman") {
			return utils.Runner.RunWithIndicator("podman system prune -af", "Cleaning podman system")
		}
		fmt.Println("podman system cleanup: Skipped (not installed)")
		return nil
	}
}
