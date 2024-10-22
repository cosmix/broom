# 🧹 Broom - System Cleanup Utility

Broom is a Go-based system cleanup utility for GNU/Linux-based operating systems that helps you free up disk space by removing unnecessary files and cleaning up various system components. It is work in progress, and has been tested only on Debian/Ubuntu/Pop!_OS etc., but should work reasonably well on other Linux distributions.

## Features

- Remove old kernel versions
- Remove unnecessary packages
- Clear APT cache
- Remove old log files
- Clean up unused Docker data
- Clean up old Snap versions
- Remove crash reports and core dumps
- Remove temporary files and old backups (excluding system directories e.g. /run, /proc, /sys, and /dev)
- Clean old systemd journal logs
- Remove old Flatpak runtimes
- Clean up user cache directories
- Clean user trash folders
- Clean up old or large log files in user home directories
- Clean up Timeshift snapshots
- Remove old Ruby gems
- Clean up Python cache files
- Remove LibreOffice cache
- Clear browser caches (Chrome, Chromium, Firefox)
- Clean package manager caches (APT, YUM, DNF)
- Clean npm cache
- Clean Gradle cache
- Clean Composer cache
- Remove old Wine prefixes
- Clean up old Electron apps cache
- Remove old Virtualbox disk images
- Clean Kdenlive render files
- Clean Blender temporary files
- Clean Steam download cache
- Clean MySQL/MariaDB binary logs
- Clean Thunderbird cache
- Clean Dropbox cache
- Clean Maven cache
- Clean Go modules cache
- Clean Rust cargo cache
- Clean Android SDK packages
- Clean JetBrains IDE caches
- Clean R packages cache
- Clean Julia packages cache
- Clean unused Conda environments
- Clean LXC/LXD images and containers
- Clean Podman images and containers
- Clean Vagrant boxes and entries
- Clean Buildah images
- Clean Mercurial backup files and bundles
- Clean Git LFS cache
- Clean CMake build directories
- Clean Autotools generated files
- Clean ccache
- Use `fd` for faster file searching when available, with fallback to `find`

Broom asks you for confirmation whenever it's about to perform a potentially destructive operation. You can choose to include or exclude specific 'cleaners' based on your requirements. At the end of a brooming session you will be presented with a summary of the cleanup operations performed, and their characteristics.

## Usage

```
sudo broom [options]
```

Options:
- `-x`: Comma-separated list of cleanup types to exclude
- `-i`: Comma-separated list of cleanup types to include
- `--all`: Apply all removal types

Example: Execute all cleaners except docker and snap
```
sudo broom -x docker,snap
```

Example: Execute only the cache and kernels cleaner
```
sudo broom -i kernels,cache
```

Note that the `-x` and `-i` options are mutually exclusive. And `-all` is mutually exclusive with all other options.

## Building

To build the executable, use the provided Makefile:

```
make
```

This will create the `broom` executable in the current directory.

To install the executable to `/usr/local/bin/`:

```
sudo make install
```

## Note

**Important:** This program requires root privileges to perform most cleanup operations. Always use with caution and consider backing up important data before running extensive cleanup operations.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
