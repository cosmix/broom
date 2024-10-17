# Broom - System Cleanup Utility

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
- Clear browser caches
- Clean package manager caches
- Clean npm cache
- Clean Gradle cache
- Clean Composer cache
- Remove old Wine prefixes
- Clean up old Electron apps cache
- Remove old Virtualbox disk images
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

Example:
```
sudo ./broom -x docker,snap -i snap,temp
```

Note that the `-x` and `-i` options are mutually exclusive.

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
