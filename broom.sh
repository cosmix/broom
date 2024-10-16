#!/bin/bash

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;94m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to display usage information
usage() {
    echo "Usage: $0 [-x exclude_types] [-i include_types] [--all]"
    echo
    echo "Options:"
    echo "  -x exclude_types   Comma-separated list of cleanup types to exclude"
    echo "  -i include_types   Comma-separated list of cleanup types to include"
    echo "  --all              Apply all removal types"
    echo
    echo "Note: -i and -x options cannot be used together."
    echo "      --all option cannot be used with -i or -x."
    echo
    echo "Available cleanup types:"
    echo "  kernels        - Remove old kernel versions"
    echo "  packages       - Remove unnecessary packages"
    echo "  apt            - Clear APT cache"
    echo "  logs           - Remove old log files"
    echo "  languages      - Remove unused language files"
    echo "  docker         - Clean up Docker data"
    echo "  snap           - Clean up old Snap versions"
    echo "  crash          - Remove crash reports and core dumps"
    echo "  home           - Clean up home directory"
    echo "  temp           - Remove temporary files and old backups"
    echo "  journal        - Clean old systemd journal logs"
    echo "  flatpak        - Remove old Flatpak runtimes"
    echo "  cache          - Clean up user cache directories"
    echo "  deb            - Remove old .deb files"
    echo "  timeshift      - Clean up Timeshift snapshots"
    echo "  ruby           - Remove old Ruby gems"
    echo "  python         - Clean up Python cache files"
    echo "  rotated_logs   - Remove old log rotation files"
    echo "  trash          - Clean user trash folders"
    echo "  libreoffice    - Remove LibreOffice cache"
    echo "  browser        - Clear browser caches"
    echo "  package_manager - Clean package manager caches"
    echo
    echo "Example:"
    echo "  $0 -x docker,snap -i kernels,packages"
    echo "  $0 --all"
    exit 1
}

# Parse command line arguments
EXCLUDE=()
INCLUDE=()
ALL_FLAG=false

# Check for single dash argument
if [ "$1" = "-" ]; then
    echo -e "${RED}Error: Invalid argument '-'. Use --help for usage information.${NC}" 1>&2
    exit 1
fi

# Parse options
while [[ $# -gt 0 ]]; do
    case $1 in
        -x)
            IFS=',' read -ra EXCLUDE <<< "$2"
            shift 2
            ;;
        -i)
            IFS=',' read -ra INCLUDE <<< "$2"
            shift 2
            ;;
        --all)
            ALL_FLAG=true
            shift
            ;;
        --help)
            usage
            ;;
        *)
            echo "Invalid Option: $1" 1>&2
            usage
            ;;
    esac
done

# Check for incompatible options
if [ ${#EXCLUDE[@]} -gt 0 ] && [ ${#INCLUDE[@]} -gt 0 ]; then
    echo -e "${RED}Error: -i and -x options cannot be used together.${NC}" 1>&2
    usage
fi

if [ "$ALL_FLAG" = true ] && { [ ${#EXCLUDE[@]} -gt 0 ] || [ ${#INCLUDE[@]} -gt 0 ]; }; then
    echo -e "${RED}Error: --all option cannot be used with -i or -x.${NC}" 1>&2
    usage
fi

# If no arguments are provided, print usage
if [ ${#EXCLUDE[@]} -eq 0 ] && [ ${#INCLUDE[@]} -eq 0 ] && [ "$ALL_FLAG" = false ]; then
    usage
fi

# Function to check if a cleanup type should be executed
should_execute() {
    local type=$1
    if [ "$ALL_FLAG" = true ]; then
        return 0
    elif [ ${#INCLUDE[@]} -ne 0 ]; then
        IFS=',' read -ra INCLUDE_ARRAY <<< "${INCLUDE[@]}"
        for i in "${INCLUDE_ARRAY[@]}"; do
            if [[ "$i" == "$type" ]]; then
                return 0
            fi
        done
        return 1
    elif [ ${#EXCLUDE[@]} -ne 0 ]; then
        IFS=',' read -ra EXCLUDE_ARRAY <<< "${EXCLUDE[@]}"
        for e in "${EXCLUDE_ARRAY[@]}"; do
            if [[ "$e" == "$type" ]]; then
                return 1
            fi
        done
        return 0
    else
        return 0
    fi
}

activity_indicator() {
    local pid=$1
    local message="$2"
    local delay=0.1
    local frames=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)
    local num_frames=${#frames[@]}
    
    # Determine the terminal width and adjust the message length if needed
    local term_width
    term_width=$(tput cols)
    local max_message_length=$((term_width - 10))

    if [ ${#message} -gt $max_message_length ]; then
        message="${message:0:$max_message_length}..."
    fi

    # Hide the cursor
    tput civis
    trap "tput cnorm; exit" SIGINT SIGTERM

    # Display the message and start the spinner
    printf "%s" "$message"

    while kill -0 "$pid" 2>/dev/null; do
        for frame in "${frames[@]}"; do
            # Save cursor position
            tput sc
            # Print the spinner at the end of the message
            printf " [%s]" "$frame"
            # Restore the cursor position after printing the spinner
            tput rc
            sleep "$delay"
        done
    done

    # Clean up after the process finishes
    tput cnorm  # Show cursor
    # Clear the entire line by printing enough spaces
    printf "\r"
    for ((i=0; i<term_width; i++)); do
        printf " "
    done
    printf "\r"  # Move the cursor back to the start of the line
}


# Function to run a command with an activity indicator
run_with_indicator() {
    local command="$1"
    local message="$2"
    eval "$command" &>/dev/null &
    activity_indicator $! "$message"
    printf "%s ${GREEN}Done${NC}\n" "$message"
}

# Function to print section headers
print_header() {
    echo -e "\n${BLUE}=======================================${NC}"
    echo -e "${BLUE}=== $1 ===${NC}"
    echo -e "${BLUE}=======================================${NC}\n"
}

# Function to check if the script is run with root privileges
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Please run this script as root or with sudo.${NC}"
        exit 1
    fi
}

# Function to get free disk space in bytes
get_free_disk_space() {
    df --output=avail --block-size=1 / | tail -1
}

# Function to format bytes to human-readable format
format_bytes() {
    local bytes=$1
    if ((bytes < 1024)); then
        echo "${bytes} B"
    elif ((bytes < 1048576)); then
        echo "$(( (bytes + 512) / 1024 )) KiB"
    elif ((bytes < 1073741824)); then
        echo "$(( (bytes + 524288) / 1048576 )) MiB"
    elif ((bytes < 1099511627776)); then
        echo "$(( (bytes + 536870912) / 1073741824 )) GiB"
    else
        echo "$(( (bytes + 549755813888) / 1099511627776 )) TiB"
    fi
}

# Function to calculate and print space freed
calculate_space_freed() {
    local start_space=$1
    local end_space=$2
    local section_name=$3
    if (( end_space > start_space )); then
        local space_freed=$((end_space - start_space))
        echo -e "${CYAN}Space freed by $section_name: ${GREEN}$(format_bytes $space_freed)${NC}"
    else
        echo -e "No significant space freed by $section_name."
    fi
}

# Function to get the current kernel version
get_current_kernel() {
    uname -r
}

# Function to list installed kernel versions
list_installed_kernels() {
    dpkg --list | grep linux-image | awk '{ print $2 }' | sort -V
}

# Function to ask for user confirmation
ask_confirmation() {
    local message="$1"
    echo -e "${YELLOW}$message${NC}"
    read -p "Do you want to proceed? (y/n): " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]]
}

# Function to remove old kernels in chunks
remove_old_kernels() {
    if should_execute "kernels"; then
        print_header "Removing Old Kernels"
        echo -e "Warning: This will remove old kernel versions. This action is not reversible."
        
        local start_space
        start_space=$(get_free_disk_space)
        
        local current_kernel
        current_kernel=$(get_current_kernel)
        local installed_kernels
        installed_kernels=$(list_installed_kernels)
        
        echo -e "Current kernel: ${GREEN}$current_kernel${NC}"
        echo -e "${BLUE}Installed kernels:${NC}"
        echo "$installed_kernels"
        echo

        local kernels_to_remove=()
        
        while read -r kernel; do
            if [[ $kernel != *"$current_kernel"* ]] && [[ $kernel == linux-image-* ]]; then
                kernels_to_remove+=("$kernel")
            fi
        done <<< "$installed_kernels"

        local total_kernels=${#kernels_to_remove[@]}
        echo -e "Total kernels to remove: ${BLUE}$total_kernels${NC}"

        for ((i=0; i<${#kernels_to_remove[@]}; i+=20)); do
            local chunk=("${kernels_to_remove[@]:i:20}")
            echo -e "\nRemoving chunk of kernels (${#chunk[@]} kernels):"
            printf '%s\n' "${chunk[@]}"
            run_with_indicator "apt-get purge -y ${chunk[*]}" "Removing kernels..."
        done

        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Old Kernels Removal"
    fi
}

# Function to remove unnecessary packages
remove_unnecessary_packages() {
    if should_execute "packages"; then
        print_header "Removing Unnecessary Packages"
        local start_space
        start_space=$(get_free_disk_space)
        run_with_indicator "apt-get autoremove -y" "Removing unnecessary packages..."
        run_with_indicator "apt-get purge -y nano vim-tiny" "Removing non-critical packages..."
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Unnecessary Packages Removal"
    fi
}

# Function to clear APT cache
clear_apt_cache() {
    if should_execute "apt"; then
        print_header "Clearing APT Cache"
        local start_space
        start_space=$(get_free_disk_space)
        run_with_indicator "apt-get clean" "Clearing APT cache..."
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "APT Cache Clearing"
    fi
}

# Function to remove old logs
remove_old_logs() {
    if should_execute "logs"; then
        print_header "Removing Old Logs"
        echo -e "Warning: This will remove old log files. This action is not reversible."
        if ask_confirmation "Do you want to proceed with removing old logs?"; then
            local start_space
            start_space=$(get_free_disk_space)
            run_with_indicator "journalctl --vacuum-time=3d" "Clearing old journal logs..."
            run_with_indicator "find /var/log -type f -name \"*.log\" -mtime +30 -delete" "Removing old log files..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Old Logs Removal"
        else
            echo -e "${YELLOW}Skipping removal of old logs.${NC}"
        fi
    fi
}

# Function to remove unused languages
remove_unused_languages() {
    if should_execute "languages"; then
        print_header "Removing Unused Languages"
        echo -e "Warning: This will remove unused language files. This action is not reversible."
        if ask_confirmation "Do you want to proceed with removing unused languages?"; then
            local start_space
            start_space=$(get_free_disk_space)
            run_with_indicator "echo 'localepurge localepurge/nopurge multiselect en, en_US.UTF-8' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/use-dpkg-feature boolean true' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/none_selected boolean false' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/mandelete boolean true' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/dontbothernew boolean true' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/showfreedspace boolean true' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/quickndirtycalc boolean true' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "echo 'localepurge localepurge/verbose boolean false' | debconf-set-selections" "Configuring localepurge..."
            run_with_indicator "apt-get install -y localepurge" "Installing localepurge..."
            run_with_indicator "localepurge" "Removing unused language files..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Unused Languages Removal"
        else
            echo -e "${YELLOW}Skipping removal of unused languages.${NC}"
        fi
    fi
}

# Function to clean up Docker
clean_docker() {
    if should_execute "docker"; then
        if command -v docker &> /dev/null; then
            print_header "Cleaning Docker"
            echo -e "Warning: This will remove unused Docker data. This action is not reversible."
            if ask_confirmation "Do you want to proceed with Docker cleanup?"; then
                local start_space
                start_space=$(get_free_disk_space)
                run_with_indicator "docker system prune -af" "Removing unused Docker data..."
                local end_space
                end_space=$(get_free_disk_space)
                calculate_space_freed "$start_space" "$end_space" "Docker Cleanup"
            else
                echo -e "${YELLOW}Skipping Docker cleanup.${NC}"
            fi
        else
            echo -e "${YELLOW}Docker is not installed. Skipping Docker cleanup.${NC}"
        fi
    fi
}

# Function to clean up old Snap versions
clean_snap() {
    if should_execute "snap"; then
        if command -v snap &> /dev/null; then
            print_header "Cleaning Snap Packages and Caches"
            local start_space
            start_space=$(get_free_disk_space)

            run_with_indicator "snap list --all | awk '/disabled/{print \$1, \$3}' | while read snapname revision; do sudo snap remove \$snapname --revision=\$revision; done" "Removing old snap versions..."
            run_with_indicator "rm -rf /var/lib/snapd/cache/*" "Clearing snap cache..."
            run_with_indicator "snap list --all | awk '/disabled/{print \$1, \$3}' | while read snapname revision; do sudo snap remove \$snapname --revision=\$revision; done" "Removing unused snaps..."
            run_with_indicator "journalctl --vacuum-time=1d" "Cleaning up snap-related logs..."
            run_with_indicator "sudo rm -rf /var/lib/snapd/desktop/applications/* && sudo rm -rf /var/lib/snapd/desktop/icons/*" "Pruning unused snap themes and icons..."
            run_with_indicator "set -eu
            snap list --all | awk '/disabled/{print \$1, \$3}' |
                while read snapname revision; do
                    snap remove \"\$snapname\" --revision=\"\$revision\"
                done" "Removing old revisions of snaps..."
            run_with_indicator "rm -rf /home/*/.config/autostart/snap*" "Clearing snap autostart entries..."

            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Snap Cleanup"
        else
            echo -e "${YELLOW}Snap is not installed. Skipping Snap cleanup.${NC}"
        fi
    fi
}

# Function to remove old crash reports and core dumps
remove_crash_reports() {
    if should_execute "crash"; then
        print_header "Removing Crash Reports and Core Dumps"
        echo -e "Warning: This will remove crash reports and core dumps. This action is not reversible."
        if ask_confirmation "Do you want to proceed with removing crash reports and core dumps?"; then
            local start_space
            start_space=$(get_free_disk_space)
            run_with_indicator "rm -rf /var/crash/*" "Removing crash reports..."
            run_with_indicator "find /var/lib/systemd/coredump/ -type f -delete" "Removing core dumps..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Crash Reports and Core Dumps Removal"
        else
            echo -e "${YELLOW}Skipping removal of crash reports and core dumps.${NC}"
        fi
    fi
}

# Function to clean up home directory
clean_home_directory() {
    if should_execute "home"; then
        print_header "Cleaning Home Directory"
        local start_space
        start_space=$(get_free_disk_space)
        run_with_indicator "find /home -type f \( -name '*.tmp' -o -name '*.temp' -o -name '*.swp' -o -name '*~' \) -delete" "Removing temporary files in home directory..."
        run_with_indicator "rm -rf /home/*/.cache/thumbnails/*" "Clearing thumbnail cache..."
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Home Directory Cleanup"
    fi
}

# Function to remove temporary files and old backups
remove_temp_and_backups() {
    if should_execute "temp"; then
        print_header "Removing Temporary Files and Old Backups"
        local start_space
        start_space=$(get_free_disk_space)
        run_with_indicator "find /tmp -type f -atime +10 -delete" "Removing old files in /tmp..."
        run_with_indicator "find /var/tmp -type f -atime +10 -delete" "Removing old files in /var/tmp..."
        run_with_indicator "find / -name \"*~\" -type f -delete" "Removing backup files..."
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Temporary Files and Old Backups Removal"
    fi
}

# Function to clean old systemd journal logs
clean_journal_logs() {
    if should_execute "journal"; then
        print_header "Cleaning Old Systemd Journal Logs"
        echo -e "Warning: This will limit the journal size to 100MB. This action is not reversible."
        if ask_confirmation "Do you want to proceed with journal log cleanup?"; then
            local start_space
            start_space=$(get_free_disk_space)
            run_with_indicator "journalctl --vacuum-size=100M" "Limiting journal size to 100MB..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Journal Logs Cleanup"
        else
            echo -e "${YELLOW}Skipping journal log cleanup.${NC}"
        fi
    fi
}

# Function to remove old flatpak runtimes
clean_flatpak() {
    if should_execute "flatpak"; then
        if command -v flatpak &> /dev/null; then
            print_header "Cleaning Old Flatpak Runtimes"
            echo -e "Warning: This will remove unused Flatpak runtimes. This action is not reversible."
            if ask_confirmation "Do you want to proceed with Flatpak cleanup?"; then
                local start_space
                start_space=$(get_free_disk_space)
                run_with_indicator "flatpak uninstall --unused -y" "Removing unused Flatpak runtimes..."
                local end_space
                end_space=$(get_free_disk_space)
                calculate_space_freed "$start_space" "$end_space" "Flatpak Cleanup"
            else
                echo -e "${YELLOW}Skipping Flatpak cleanup.${NC}"
            fi
        else
            echo -e "${YELLOW}Flatpak is not installed. Skipping Flatpak cleanup.${NC}"
        fi
    fi
}

# Function to clean up user cache directories
clean_user_caches() {
    if should_execute "cache"; then
        print_header "Cleaning User Cache Directories"
        local start_space
        start_space=$(get_free_disk_space)
        for user_home in /home/*; do
            if [ -d "$user_home/.cache" ]; then
                run_with_indicator "rm -rf $user_home/.cache/*" "Clearing cache for $(basename "$user_home")..."
            fi
        done
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "User Cache Cleanup"
    fi
}

# Function to remove old .deb files
remove_old_deb_files() {
    if should_execute "deb"; then
        print_header "Removing Old .deb Files"
        local start_space
        start_space=$(get_free_disk_space)
        run_with_indicator "rm -f /var/cache/apt/archives/*.deb" "Removing old .deb files..."
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Old .deb Files Removal"
    fi
}

# Function to clean up Timeshift snapshots
clean_timeshift_snapshots() {
    if should_execute "timeshift"; then
        if command -v timeshift &> /dev/null; then
            print_header "Cleaning Old Timeshift Snapshots"
            echo -e "Warning: This will remove old Timeshift snapshots. This action is not reversible."
            if ask_confirmation "Do you want to proceed with Timeshift snapshots cleanup?"; then
                local start_space
                start_space=$(get_free_disk_space)
                run_with_indicator "timeshift --list | grep -oP '(?<=\s)\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}' | sort | head -n -3 | xargs -I {} timeshift --delete --snapshot '{}'" "Removing old Timeshift snapshots..."
                local end_space
                end_space=$(get_free_disk_space)
                calculate_space_freed "$start_space" "$end_space" "Timeshift Snapshots Cleanup"
            else
                echo -e "${YELLOW}Skipping Timeshift snapshots cleanup.${NC}"
            fi
        else
            echo -e "${YELLOW}Timeshift is not installed. Skipping Timeshift cleanup.${NC}"
        fi
    fi
}

# Function to remove old Ruby gems
clean_ruby_gems() {
    if should_execute "ruby"; then
        if command -v gem &> /dev/null; then
            print_header "Cleaning Old Ruby Gems"
            local start_space
            start_space=$(get_free_disk_space)
            run_with_indicator "gem cleanup" "Removing old Ruby gems..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Ruby Gems Cleanup"
        else
            echo -e "${YELLOW}Ruby is not installed. Skipping Ruby gems cleanup.${NC}"
        fi
    fi
}

# Function to clean up Python cache files
clean_python_cache() {
    if should_execute "python"; then
        print_header "Cleaning Python Cache Files"
        local start_space
        start_space=$(get_free_disk_space)
        run_with_indicator "find / -type d -name __pycache__ -exec rm -rf {} +" "Removing Python cache files..."
        run_with_indicator "find / -name '*.pyc' -delete" "Removing .pyc files..."
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Python Cache Cleanup"
    fi
}

# Function to remove old log rotation files
remove_old_rotated_logs() {
    if should_execute "rotated_logs"; then
        print_header "Removing Old Rotated Logs"
        echo -e "Warning: This will remove old rotated log files. This action is not reversible."
        if ask_confirmation "Do you want to proceed with removing old rotated logs?"; then
            local start_space
            start_space=$(get_free_disk_space)
            run_with_indicator "find /var/log -type f -name '*.gz' -delete" "Removing compressed log files..."
            run_with_indicator "find /var/log -type f -name '*.[0-9]' -delete" "Removing numbered log files..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "Old Rotated Logs Removal"
        else
            echo -e "${YELLOW}Skipping removal of old rotated logs.${NC}"
        fi
    fi
}

# Function to clean user trash folders
clean_user_trash() {
    if should_execute "trash"; then
        print_header "Cleaning User Trash Folders"
        echo -e "Warning: This will empty user trash folders. This action is not reversible."
        if ask_confirmation "Do you want to proceed with user trash cleanup?"; then
            local start_space
            start_space=$(get_free_disk_space)
            for user_home in /home/*; do
                if [ -d "$user_home/.local/share/Trash" ]; then
                    run_with_indicator "rm -rf $user_home/.local/share/Trash/*" "Emptying trash for $(basename "$user_home")..."
                fi
            done
            run_with_indicator "rm -rf /root/.local/share/Trash/*" "Emptying trash for root..."
            local end_space
            end_space=$(get_free_disk_space)
            calculate_space_freed "$start_space" "$end_space" "User Trash Cleanup"
        else
            echo -e "${YELLOW}Skipping user trash cleanup.${NC}"
        fi
    fi
}

# Function to remove LibreOffice cache
clean_libreoffice_cache() {
    if should_execute "libreoffice"; then
        print_header "Cleaning LibreOffice Cache"
        local start_space
        start_space=$(get_free_disk_space)
        for user_home in /home/*; do
            if [ -d "$user_home/.config/libreoffice" ]; then
                run_with_indicator "rm -rf $user_home/.config/libreoffice/4/user/uno_packages/cache" "Clearing LibreOffice cache for $(basename "$user_home")..."
            fi
        done
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "LibreOffice Cache Cleanup"
    fi
}

# Function to clear browser caches
clear_browser_caches() {
    if should_execute "browser"; then
        print_header "Clearing Browser Caches"
        local start_space
        start_space=$(get_free_disk_space)
        for user_home in /home/*; do
            # Chrome/Chromium
            if [ -d "$user_home/.cache/google-chrome" ]; then
                run_with_indicator "rm -rf $user_home/.cache/google-chrome/Default/Cache/*" "Clearing Chrome cache for $(basename "$user_home")..."
            fi
            if [ -d "$user_home/.cache/chromium" ]; then
                run_with_indicator "rm -rf $user_home/.cache/chromium/Default/Cache/*" "Clearing Chromium cache for $(basename "$user_home")..."
            fi
            # Firefox
            if [ -d "$user_home/.mozilla/firefox" ]; then
                run_with_indicator "find $user_home/.mozilla/firefox -type d -name 'Cache' -exec rm -rf {}/* \;" "Clearing Firefox cache for $(basename "$user_home")..."
            fi
        done
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Browser Caches Cleanup"
    fi
}

# Function to clean package manager caches
clean_package_manager_caches() {
    if should_execute "package_manager"; then
        print_header "Cleaning Package Manager Caches"
        local start_space
        start_space=$(get_free_disk_space)
        if command -v apt-get &> /dev/null; then
            run_with_indicator "apt-get clean" "Cleaning APT cache..."
        fi
        if command -v yum &> /dev/null; then
            run_with_indicator "yum clean all" "Cleaning YUM cache..."
        fi
        if command -v dnf &> /dev/null; then
            run_with_indicator "dnf clean all" "Cleaning DNF cache..."
        fi
        local end_space
        end_space=$(get_free_disk_space)
        calculate_space_freed "$start_space" "$end_space" "Package Manager Caches Cleanup"
    fi
}

# Main execution
clear
echo -e "${GREEN}==========${NC}"
echo -e "${GREEN}  Broom  ${NC}"
echo -e "${GREEN}==========${NC}\n"

check_root

print_header "Disk Space Before Cleanup"
start_space=$(get_free_disk_space)
echo -e "Free disk space before cleanup: ${GREEN}$(format_bytes "$start_space")${NC}"

remove_old_kernels
remove_unnecessary_packages
clear_apt_cache
remove_old_logs
remove_unused_languages
clean_docker
clean_snap
remove_crash_reports
clean_home_directory
remove_temp_and_backups
clean_journal_logs
clean_flatpak
clean_user_caches
remove_old_deb_files
clean_timeshift_snapshots
clean_ruby_gems
clean_python_cache
remove_old_rotated_logs
clean_user_trash
clean_libreoffice_cache
clear_browser_caches
clean_package_manager_caches

print_header "Disk Space After Cleanup"
end_space=$(get_free_disk_space)
echo -e "Free disk space after cleanup: ${GREEN}$(format_bytes "$end_space")${NC}"

if (( end_space > start_space )); then
    space_freed=$((end_space - start_space))
    echo -e "\n${GREEN}Total disk space freed: ${CYAN}$(format_bytes "$space_freed")${NC}"
else
    echo -e "\nNo significant disk space was freed during this cleanup."
    echo -e "This can happen if the system was already clean or if freed space was immediately reallocated."
fi

echo -e "\n${GREEN}=======================================${NC}"
echo -e "${GREEN}     System Cleanup Completed!         ${NC}"
echo -e "${GREEN}=======================================${NC}"
echo -e "\n${BLUE}Please review the output above for any actions you may need to take manually.${NC}"
echo -e "${BLUE}Remember to reboot your system if any critical packages were removed.${NC}"
