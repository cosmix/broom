# The Broom

This bash script provides a comprehensive system cleanup solution for Linux systems, particularly those based on Debian/Ubuntu. It performs various cleanup operations to free up disk space and optimize system performance.

## Features

- Removes old kernels
- Removes unnecessary packages
- Clears APT cache
- Removes old logs
- Removes unused languages
- Cleans Docker data (if installed)
- Cleans Snap packages and caches (if installed)
- Removes crash reports and core dumps
- Cleans home directory temporary files
- Removes temporary files and old backups
- Cleans old systemd journal logs
- Removes old Flatpak runtimes (if installed)
- Cleans user cache directories
- Removes old .deb files
- Cleans up Timeshift snapshots (if installed)
- Removes old Ruby gems (if Ruby is installed)
- Cleans Python cache files
- Removes old rotated logs
- Cleans user trash folders
- Removes LibreOffice cache
- Clears browser caches (Chrome, Chromium, Firefox)
- Cleans package manager caches (APT, YUM, DNF)

## Usage

1. Make sure you have root privileges or run the script with sudo.
2. Make the script executable:
   ```
   chmod +x cleanup.sh
   ```
3. Run the script:
   ```
   sudo ./cleanup.sh
   ```

## Notes

- The script provides visual feedback on the progress of each operation.
- It calculates and displays the amount of disk space freed after each operation and at the end of the entire cleanup process.
- Some operations are only performed if the relevant software (e.g., Docker, Snap, Flatpak) is installed on the system.
- The script is designed to be safe, but it's always a good idea to review the operations and ensure they align with your system's needs.
- It's recommended to reboot the system after running this script, especially if critical packages were removed.

## Caution

While this script is designed to be safe, it's always a good practice to backup important data before performing system-wide cleanup operations.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
