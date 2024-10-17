package utils

import (
	"io"
	"os"
	"testing"
)

func TestGetFreeDiskSpace(t *testing.T) {
	space := GetFreeDiskSpace()
	if space == 0 {
		t.Error("GetFreeDiskSpace returned 0, expected non-zero value")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1099511627776, "1.0 TiB"},
		{1125899906842624, "1.0 PiB"},
		{1152921504606846976, "1.0 EiB"},
		{18446744073709551615, "Size too large to calculate"},
	}

	for _, test := range tests {
		result := FormatBytes(test.input)
		if result != test.expected {
			t.Errorf("FormatBytes(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestCalculateSpaceFreed(t *testing.T) {
	tests := []struct {
		startSpace      uint64
		endSpace        uint64
		sectionName     string
		expectedSpace   uint64
		expectedMessage string
	}{
		{1000, 2000, "test1", 1000, "Space freed by test1: 1000 B"},
		{2000, 1000, "test2", 0, "Space freed by test2: Insignificant (possible reallocation)"},
		{1048576, 2097152, "test3", 1048576, "Space freed by test3: 1.0 MiB"},
	}

	for _, test := range tests {
		spaceFreed, message := CalculateSpaceFreed(test.startSpace, test.endSpace, test.sectionName)
		if spaceFreed != test.expectedSpace {
			t.Errorf("CalculateSpaceFreed(%d, %d, %s) returned space %d; want %d",
				test.startSpace, test.endSpace, test.sectionName, spaceFreed, test.expectedSpace)
		}
		if message != test.expectedMessage {
			t.Errorf("CalculateSpaceFreed(%d, %d, %s) returned message '%s'; want '%s'",
				test.startSpace, test.endSpace, test.sectionName, message, test.expectedMessage)
		}
	}
}

func TestRunWithIndicator(t *testing.T) {
	// Test with a valid command expected to succeed
	err := RunWithIndicator("echo 'test'", "Testing RunWithIndicator success")
	if err != nil {
		t.Errorf("RunWithIndicator returned an error for valid command: %v", err)
	}

	// Test with a valid command expected to fail
	err = RunWithIndicator("ls /nonexistent_directory", "Testing RunWithIndicator failure")
	if err == nil {
		t.Error("RunWithIndicator should have returned an error for failing command")
	}

	// Test with an invalid/nonexistent command expected to fail
	err = RunWithIndicator("sldfkj", "Testing RunWithIndicator failure")
	if err == nil {
		t.Error("RunWithIndicator should have returned an error for a nonexistent command")
	}
}

func TestRunFdOrFind(t *testing.T) {
	// Test with a path expected to exist
	err := RunFdOrFind("/tmp", "-type d", "Testing RunFdOrFind success", true)
	if err != nil {
		t.Errorf("RunFdOrFind returned an error for valid path: %v", err)
	}

	// Test with a path expected not to exist
	err = RunFdOrFind("/nonexistent_path", "-type d", "Testing RunFdOrFind failure", false)
	if err == nil {
		t.Error("RunFdOrFind should have returned an error for non-existent path")
	}
}

func TestCommandExists(t *testing.T) {
	if !CommandExists("ls") {
		t.Error("CommandExists returned false for 'ls', expected true")
	}

	if CommandExists("non_existent_command") {
		t.Error("CommandExists returned true for 'non_existent_command', expected false")
	}
}

func TestAskConfirmation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"y", true},
		{"yes", true},
		{"n", false},
		{"no", false},
		{"", false},
	}

	for _, test := range tests {
		result := askConfirmationMock(test.input)
		if result != test.expected {
			t.Errorf("AskConfirmation with input '%s' = %v; want %v", test.input, result, test.expected)
		}
	}
}

func askConfirmationMock(input string) bool {
	switch input {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func TestCheckRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping TestCheckRoot when not running as root")
	}

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("os.Exit called")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("CheckRoot did not call os.Exit")
		}
		if exitCode != 1 {
			t.Errorf("CheckRoot called os.Exit with %d, want 1", exitCode)
		}
	}()

	CheckRoot()
}

func TestPrintHeader(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintHeader("Test Header")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	expected := "\n===============\n  Test Header  \n===============\n\n"
	if string(out) != expected {
		t.Errorf("PrintHeader output = %s; want %s", string(out), expected)
	}
}

func TestPrintBanner(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintBanner()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	if len(string(out)) == 0 {
		t.Error("PrintBanner output is empty")
	}
}

func TestPrintCompletionBanner(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintCompletionBanner()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	if len(string(out)) == 0 {
		t.Error("PrintCompletionBanner output is empty")
	}
}

var osExit = os.Exit
