package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cosmix/broom/internal/cleaners"
	"github.com/cosmix/broom/internal/utils"
	"github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
)

var (
	au = aurora.NewAurora(true)
)

type cleanupResult struct {
	cleanupType string
	result      string
	err         error
	spaceFreed  uint64
	duration    time.Duration
	skipped     bool
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Exiting...")
		cancel()
	}()

	utils.CheckRoot()

	excludeTypes := flag.String("x", "", "Comma-separated list of cleanup types to exclude")
	includeTypes := flag.String("i", "", "Comma-separated list of cleanup types to include")
	allFlag := flag.Bool("all", false, "Apply all removal types")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-x exclude_types] [-i include_types] [--all]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nAvailable cleanup types:\n")
		for _, t := range cleaners.GetAllCleanupTypes() {
			fmt.Fprintf(os.Stderr, "  %s\n", t)
		}
	}

	flag.Parse()

	typesToRun, err := parseFlags(*excludeTypes, *includeTypes, *allFlag)
	if err != nil {
		fmt.Println(au.Red(fmt.Sprintf("Error: %s", err)))
		flag.Usage()
		os.Exit(1)
	}

	utils.PrintBanner()

	startSpace := utils.GetFreeDiskSpace()
	fmt.Println(au.Blue(fmt.Sprintf("Free disk space before cleanup: %s", utils.FormatBytes(startSpace))))

	results := performCleanups(ctx, typesToRun)

	endSpace := utils.GetFreeDiskSpace()
	fmt.Println(au.Blue(fmt.Sprintf("\nFree disk space after cleanup: %s", utils.FormatBytes(endSpace))))

	var totalSpaceFreed uint64
	for _, result := range results {
		totalSpaceFreed += result.spaceFreed
	}

	if totalSpaceFreed > 0 {
		fmt.Println(au.Green(fmt.Sprintf("\nTotal disk space freed: %s", utils.FormatBytes(totalSpaceFreed))))
	} else {
		fmt.Println(au.Blue("\nInsignificant disk space freed."))
		fmt.Println(au.Blue("This can happen if the system was already clean or if freed space was immediately reallocated."))
	}

	printCleanupSummary(results, totalSpaceFreed, startSpace)

	utils.PrintCompletionBanner()
}

func performCleanups(ctx context.Context, typesToRun []string) []cleanupResult {
	results := make([]cleanupResult, 0, len(typesToRun))

	for _, cleanupType := range typesToRun {
		select {
		case <-ctx.Done():
			return results
		default:
			utils.PrintHeader(cleanupType)

			if needsConfirmation(cleanupType) {
				prompt := promptui.Prompt{
					Label:     fmt.Sprintf("Do you want to proceed with %s cleanup", cleanupType),
					IsConfirm: true,
				}

				result, err := prompt.Run()
				if err != nil || strings.ToLower(result) != "y" {
					fmt.Printf("Skipping %s cleanup\n\n", cleanupType)
					results = append(results, cleanupResult{
						cleanupType: cleanupType,
						result:      "Skipped",
						skipped:     true,
					})
					continue
				}
			}

			startTime := time.Now()
			spaceFreed, err := cleaners.PerformCleanup(cleanupType)
			duration := time.Since(startTime)

			result := cleanupResult{
				cleanupType: cleanupType,
				result:      "Cleanup completed successfully",
				err:         err,
				spaceFreed:  spaceFreed,
				duration:    duration,
				skipped:     false,
			}

			if err != nil {
				result.result = fmt.Sprintf("Error during cleanup: %v", err)
			}

			results = append(results, result)

			// Print result immediately after each cleanup
			if result.err != nil {
				fmt.Println(au.Red(result.result))
			} else {
				fmt.Println(au.Green(result.result))
			}
			spaceFreedStr := utils.FormatBytes(result.spaceFreed)
			if result.spaceFreed == 0 {
				spaceFreedStr = "Insignificant"
			}
			fmt.Println(au.Blue(fmt.Sprintf("Space freed: %s", spaceFreedStr)))
			durationValue, durationUnit := formatDuration(result.duration)
			fmt.Printf(au.Blue("Time taken: %.2f%s\n").String(), durationValue, durationUnit)
			fmt.Println() // Add a newline for better separation between cleanup types
		}
	}

	return results
}

func printCleanupSummary(results []cleanupResult, totalSpaceFreed, startSpace uint64) {
	fmt.Println(au.Bold("\nCleanup Summary:"))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Cleanup Type", "Status", "Space Freed", "Time Taken"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})

	maxSpaceFreed := uint64(0)
	maxDuration := time.Duration(0)

	for _, result := range results {
		if result.spaceFreed > maxSpaceFreed {
			maxSpaceFreed = result.spaceFreed
		}
		if result.duration > maxDuration {
			maxDuration = result.duration
		}
	}

	for _, result := range results {
		status := getColoredStatus(result.skipped, result.err)
		spaceFreed := getColoredSpaceFreed(result.spaceFreed, maxSpaceFreed, startSpace)
		timeTaken := getColoredDuration(result.duration, maxDuration)

		table.Append([]string{result.cleanupType, status, spaceFreed, timeTaken})
	}

	table.SetFooter([]string{"Total", "", utils.FormatBytes(totalSpaceFreed), ""})
	table.Render()
}

func getColoredStatus(skipped bool, err error) string {
	if skipped {
		return fmt.Sprintf("\x1b[38;2;255;165;0m%s\x1b[0m", "Skipped") // Orange
	} else if err != nil {
		return fmt.Sprintf("\x1b[31m%s\x1b[0m", "Error") // Red
	}
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", "Success") // Green
}

func getColoredSpaceFreed(spaceFreed, maxSpaceFreed, startSpace uint64) string {
	if spaceFreed == 0 {
		return fmt.Sprintf("\x1b[32m%s\x1b[0m", "Insignificant") // Green
	}

	threshold := startSpace / 2 // 50% of the disk space previously available
	if maxSpaceFreed > threshold {
		maxSpaceFreed = threshold
	}

	ratio := float64(spaceFreed) / float64(maxSpaceFreed)
	r, g, b := getHeatmapColor(ratio)

	// Format the space freed and ensure the entire string is colored
	formattedSpace := utils.FormatBytes(spaceFreed)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, formattedSpace)
}

func getColoredDuration(duration, maxDuration time.Duration) string {
	durationValue, durationUnit := formatDuration(duration)
	ratio := float64(duration) / float64(maxDuration)
	r, g, b := getHeatmapColor(ratio)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%.2f%s\x1b[0m", r, g, b, durationValue, durationUnit)
}

func getHeatmapColor(ratio float64) (r, g, b uint8) {
	if ratio < 0.5 {
		return 0, 255, uint8(255 * (1 - 2*ratio))
	}
	return uint8(255 * (2*ratio - 1)), uint8(255 * (2 - 2*ratio)), 0
}

func parseFlags(excludeTypes, includeTypes string, allFlag bool) ([]string, error) {
	switch {
	case excludeTypes != "" && includeTypes != "":
		return nil, fmt.Errorf("-i and -x options cannot be used together")
	case allFlag && (excludeTypes != "" || includeTypes != ""):
		return nil, fmt.Errorf("--all option cannot be used with -i or -x")
	case !allFlag && excludeTypes == "" && includeTypes == "":
		return nil, fmt.Errorf("no cleanup types specified")
	}

	cleanupTypes := cleaners.GetAllCleanupTypes()
	var typesToRun []string

	if allFlag {
		typesToRun = cleanupTypes
	} else if includeTypes != "" {
		for _, t := range strings.Split(includeTypes, ",") {
			if !contains(cleanupTypes, t) {
				return nil, fmt.Errorf("invalid cleanup type: %s", t)
			}
			typesToRun = append(typesToRun, t)
		}
	} else {
		excludeMap := make(map[string]bool)
		for _, t := range strings.Split(excludeTypes, ",") {
			if !contains(cleanupTypes, t) {
				return nil, fmt.Errorf("invalid cleanup type: %s", t)
			}
			excludeMap[t] = true
		}
		for _, t := range cleanupTypes {
			if !excludeMap[t] {
				typesToRun = append(typesToRun, t)
			}
		}
	}

	return typesToRun, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func needsConfirmation(cleanupType string) bool {
	cleaner, ok := cleaners.GetCleaner(cleanupType)
	if !ok {
		return false
	}
	return cleaner.RequiresConfirmation
}

func formatDuration(d time.Duration) (float64, string) {
	switch {
	case d.Seconds() < 1:
		return d.Seconds() * 1000, "ms"
	case d.Minutes() < 1:
		return d.Seconds(), "s"
	default:
		return d.Minutes(), "m"
	}
}
