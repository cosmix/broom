# Broom - System Cleanup Utility

Broom is a Go-based system cleanup utility for GNU/Linux-based operating systems that helps you free up disk space by removing unnecessary files and cleaning up various system components. It has been tested on Debian/Ubuntu/Pop!_OS etc., but should work on other Linux distributions as well.

## Features

- Remove old kernel versions
- Remove unnecessary packages
- Clear APT cache
- Remove old log files
- Remove unused language files
- Clean up Docker data
- Clean up old Snap versions
- Remove crash reports and core dumps
- Clean up home directory
- Remove temporary files and old backups (excluding system directories like /run, /proc, /sys, and /dev)
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
sudo ./broom -x docker,snap -i kernels,packages
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

This program requires root privileges to perform most cleanup operations. Always use with caution and consider backing up important data before running extensive cleanup operations.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
