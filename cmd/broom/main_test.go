package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestPrintCleanupSummary(t *testing.T) {
	tests := []struct {
		name            string
		results         []cleanupResult
		totalSpaceFreed uint64
		expectedOutput  *regexp.Regexp
	}{
		{
			name: "Single successful cleanup",
			results: []cleanupResult{
				{
					cleanupType: "testType1",
					result:      "Cleanup completed successfully",
					err:         nil,
					spaceFreed:  1024,
					duration:    2 * time.Second,
				},
			},
			totalSpaceFreed: 1024,
			expectedOutput: regexp.MustCompile(`(?s)Cleanup Summary:.*CLEANUP TYPE\s+\|\s+STATUS\s+\|\s+SPACE FREED\s+\|\s+TIME TAKEN.*` +
				`-+\s+` +
				`testType1\s+\|\s+Success\s+\|\s+1\.0 KiB\s+\|\s+2\.00s.*` +
				`TOTAL\s+\|\s+1\.0 KIB\s+\|`),
		},
		{
			name: "Multiple cleanups with errors",
			results: []cleanupResult{
				{
					cleanupType: "testType1",
					result:      "Cleanup completed successfully",
					err:         nil,
					spaceFreed:  1024,
					duration:    2 * time.Second,
				},
				{
					cleanupType: "testType2",
					result:      "Error during cleanup",
					err:         errors.New("error"),
					spaceFreed:  0,
					duration:    1 * time.Second,
				},
			},
			totalSpaceFreed: 1024,
			expectedOutput: regexp.MustCompile(`(?s)Cleanup Summary:.*CLEANUP TYPE\s+\|\s+STATUS\s+\|\s+SPACE FREED\s+\|\s+TIME TAKEN.*` +
				`-+\s+` +
				`testType1\s+\|\s+Success\s+\|\s+1\.0 KiB\s+\|\s+2\.00s.*` +
				`testType2\s+\|\s+Error\s+\|\s+Insignificant\s+\|\s+1\.00s.*` +
				`TOTAL\s+\|\s+1\.0 KIB\s+\|`),
		},
		{
			name: "No space freed",
			results: []cleanupResult{
				{
					cleanupType: "testType1",
					result:      "Cleanup completed successfully",
					err:         nil,
					spaceFreed:  0,
					duration:    2 * time.Second,
				},
			},
			totalSpaceFreed: 0,
			expectedOutput: regexp.MustCompile(`(?s)Cleanup Summary:.*CLEANUP TYPE\s+\|\s+STATUS\s+\|\s+SPACE FREED\s+\|\s+TIME TAKEN.*` +
				`-+\s+` +
				`testType1\s+\|\s+Success\s+\|\s+Insignificant\s+\|\s+2\.00s.*` +
				`TOTAL\s+\|\s+0 B\s+\|`),
		},
		{
			name: "All cleanups failed",
			results: []cleanupResult{
				{
					cleanupType: "testType1",
					result:      "Error during cleanup",
					err:         errors.New("error"),
					spaceFreed:  0,
					duration:    2 * time.Second,
				},
				{
					cleanupType: "testType2",
					result:      "Error during cleanup",
					err:         errors.New("error"),
					spaceFreed:  0,
					duration:    1 * time.Second,
				},
			},
			totalSpaceFreed: 0,
			expectedOutput: regexp.MustCompile(`(?s)Cleanup Summary:.*CLEANUP TYPE\s+\|\s+STATUS\s+\|\s+SPACE FREED\s+\|\s+TIME TAKEN.*` +
				`-+\s+` +
				`testType1\s+\|\s+Error\s+\|\s+Insignificant\s+\|\s+2\.00s.*` +
				`testType2\s+\|\s+Error\s+\|\s+Insignificant\s+\|\s+1\.00s.*` +
				`TOTAL\s+\|\s+0 B\s+\|`),
		},
		{
			name: "Mixed results with significant space freed",
			results: []cleanupResult{
				{
					cleanupType: "testType1",
					result:      "Cleanup completed successfully",
					err:         nil,
					spaceFreed:  2048,
					duration:    2 * time.Second,
				},
				{
					cleanupType: "testType2",
					result:      "Error during cleanup",
					err:         errors.New("error"),
					spaceFreed:  0,
					duration:    1 * time.Second,
				},
				{
					cleanupType: "testType3",
					result:      "Cleanup completed successfully",
					err:         nil,
					spaceFreed:  4096,
					duration:    3 * time.Second,
				},
			},
			totalSpaceFreed: 6144,
			expectedOutput: regexp.MustCompile(`(?s)Cleanup Summary:.*CLEANUP TYPE\s+\|\s+STATUS\s+\|\s+SPACE FREED\s+\|\s+TIME TAKEN.*` +
				`-+\s+` +
				`testType1\s+\|\s+Success\s+\|\s+2\.0 KiB\s+\|\s+2\.00s.*` +
				`testType2\s+\|\s+Error\s+\|\s+Insignificant\s+\|\s+1\.00s.*` +
				`testType3\s+\|\s+Success\s+\|\s+4\.0 KiB\s+\|\s+3\.00s.*` +
				`TOTAL\s+\|\s+6\.0 KIB\s+\|`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printCleanupSummary(tt.results, tt.totalSpaceFreed)

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout

			if !tt.expectedOutput.MatchString(buf.String()) {
				t.Errorf("output does not match expected pattern.\nGot:\n%s\nExpected to match:\n%s", buf.String(), tt.expectedOutput.String())
			}
		})
	}
}
