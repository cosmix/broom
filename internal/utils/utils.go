package utils

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// UtilsRunner interface for mocking utils functions
type UtilsRunner interface {
	RunWithIndicator(command, message string) error
	RunFdOrFind(path, args, message string, sudo bool) error
}

// DefaultUtilsRunner implements UtilsRunner with actual utils functions
type DefaultUtilsRunner struct{}

func (r DefaultUtilsRunner) RunWithIndicator(command, message string) error {
	return RunWithIndicator(command, message)
}

func (r DefaultUtilsRunner) RunFdOrFind(path, args, message string, sudo bool) error {
	return RunFdOrFind(path, args, message, sudo)
}

var Runner UtilsRunner = DefaultUtilsRunner{}

// SetUtilsRunner allows injection of a custom UtilsRunner (useful for testing)
func SetUtilsRunner(r UtilsRunner) {
	Runner = r
}

// GetFreeDiskSpace returns the amount of free disk space in bytes
func GetFreeDiskSpace() uint64 {
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	return stat.Bavail * uint64(stat.Bsize)
}

// FormatBytes formats a byte size into a human-readable string
func FormatBytes(bytes uint64) string {
	if bytes == math.MaxUint64 {
		return "Size too large to calculate"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := float64(unit), 0
	bytesFloat := float64(bytes)
	for n := bytesFloat / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	if exp > 6 || math.IsInf(bytesFloat/div, 0) {
		return "Size too large to calculate"
	}
	return fmt.Sprintf("%.1f %ciB", bytesFloat/div, "KMGTPE"[exp])
}

// CalculateSpaceFreed calculates the space freed by a cleanup operation
func CalculateSpaceFreed(startSpace, endSpace uint64, sectionName string) (uint64, string) {
	const maxReasonableSpaceFreed = 1024 * 1024 * 1024 * 1024 // 1 TiB
	const debugLogFile = "/tmp/broom_debug.log"

	logDebug := func(message string) {
		f, err := os.OpenFile(debugLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		fmt.Fprintf(f, "[%s] %s: %s\n", time.Now().Format(time.RFC3339), sectionName, message)
	}

	logDebug(fmt.Sprintf("Start space: %d, End space: %d", startSpace, endSpace))

	if endSpace > startSpace {
		spaceFreed := endSpace - startSpace
		if spaceFreed > maxReasonableSpaceFreed {
			logDebug(fmt.Sprintf("Unreasonable space freed: %d bytes", spaceFreed))
			return 0, fmt.Sprintf("Space freed by %s: Calculation error (unreasonable value)", sectionName)
		}
		logDebug(fmt.Sprintf("Space freed: %d bytes", spaceFreed))
		return spaceFreed, fmt.Sprintf("Space freed by %s: %s", sectionName, FormatBytes(spaceFreed))
	}

	if startSpace > endSpace {
		logDebug(fmt.Sprintf("Negative space freed: %d bytes", startSpace-endSpace))
		return 0, fmt.Sprintf("Space freed by %s: Insignificant (possible reallocation)", sectionName)
	}

	logDebug("No space freed")
	return 0, fmt.Sprintf("Space freed by %s: Insignificant", sectionName)
}

// RunWithIndicator runs a command with a spinner indicator and a message
func RunWithIndicator(command, message string) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()

	s.Stop()
	if err != nil {
		color.Red("Error: %s", message)
		fmt.Printf("Error executing command: %v\n", err)
		return err
	}
	color.Green("Done: %s", message)
	return nil
}

// RunFdOrFind runs a command using fd if available, or find if not, with an optional message
func RunFdOrFind(path, args, message string, ignoreErrors bool) error {
	var command string
	if CommandExists("fd") {
		command = fmt.Sprintf("fd %s %s -E /snap", args, path)
	} else {
		command = fmt.Sprintf("find %s %s -not -path '/snap/*'", path, args)
	}

	if ignoreErrors {
		command += " 2>/dev/null || true"
	}

	return RunWithIndicator(command, message)
}

// CommandExists checks if a command exists in the PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// AskConfirmation asks the user for confirmation before proceeding
func AskConfirmation(message string) bool {
	fmt.Printf("Warning: %s\n", message)
	fmt.Print("Do you want to proceed? (y/n): ")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// CheckRoot checks if the program is running as root
func CheckRoot() {
	if os.Geteuid() != 0 {
		fmt.Println("Please run this program as root or with sudo.")
		os.Exit(1)
	}
}

// PrintHeader prints a header with a border around it
func PrintHeader(header string) {
	border := strings.Repeat("=", len(header)+4)
	fmt.Println("\n" + border)
	fmt.Printf("  %s  \n", header)
	fmt.Println(border)
	fmt.Println()
}

// PrintBanner prints the program banner
func PrintBanner() {
	fmt.Println("==========================")
	fmt.Println("▗▄▄▖ ▗▄▄▖  ▗▄▖  ▗▄▖ ▗▖  ▗▖")
	fmt.Println("▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▛▚▞▜▌")
	fmt.Println("▐▛▀▚▖▐▛▀▚▖▐▌ ▐▌▐▌ ▐▌▐▌  ▐▌")
	fmt.Println("▐▙▄▞▘▐▌ ▐▌▝▚▄▞▘▝▚▄▞▘▐▌  ▐▌")
	fmt.Println("==========================")
	fmt.Println()
}

// PrintCompletionBanner prints a completion banner
func PrintCompletionBanner() {
	fmt.Println("\n=======================================")
	fmt.Println("     System Cleanup Completed!         ")
	fmt.Println("=======================================")
	fmt.Println("\nPlease review the output above for any actions you may need to take manually.")
	fmt.Println("Remember to reboot your system if any critical packages were removed.")
	fmt.Println()
}
