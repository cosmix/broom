package main

import (
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
}

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Exiting...")
		os.Exit(0)
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

	results := performCleanups(typesToRun)

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

	printCleanupSummary(results, totalSpaceFreed)

	utils.PrintCompletionBanner()
}

func performCleanups(typesToRun []string) []cleanupResult {
	results := make([]cleanupResult, 0, len(typesToRun))

	for _, cleanupType := range typesToRun {
		utils.PrintHeader(cleanupType)

		if needsConfirmation(cleanupType) {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Do you want to proceed with %s cleanup", cleanupType),
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil {
				fmt.Printf("Skipping %s cleanup\n\n", cleanupType)
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
		}

		if err != nil {
			result.result = fmt.Sprintf("Error during cleanup: %v", err)
		}

		results = append(results, result)
		fmt.Println() // Add a newline for better separation between cleanup types
	}

	// Print results after all cleanups are done
	printCleanupResults(results)

	return results
}

func printCleanupResults(results []cleanupResult) {
	for _, result := range results {
		utils.PrintHeader(result.cleanupType)
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
		fmt.Println()
	}
}

func printCleanupSummary(results []cleanupResult, totalSpaceFreed uint64) {
	fmt.Println(au.Bold("\nCleanup Summary:"))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Cleanup Type", "Status", "Space Freed", "Time Taken"})
	table.SetBorder(false)

	for _, result := range results {
		status := "Success"
		if result.err != nil {
			status = "Error"
		}
		spaceFreed := utils.FormatBytes(result.spaceFreed)
		if result.spaceFreed == 0 {
			spaceFreed = "Insignificant"
		}
		durationValue, durationUnit := formatDuration(result.duration)
		timeTaken := fmt.Sprintf("%.2f%s", durationValue, durationUnit)

		table.Append([]string{result.cleanupType, status, spaceFreed, timeTaken})
	}

	table.SetFooter([]string{"Total", "", utils.FormatBytes(totalSpaceFreed), ""})
	table.Render()
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
	confirmationTypes := []string{
		"kernels", "docker", "crash", "journal", "flatpak", "timeshift", "trash",
	}
	return contains(confirmationTypes, cleanupType)
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
