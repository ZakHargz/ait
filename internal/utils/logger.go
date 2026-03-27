package utils

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	// Colors
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()

	// Color disabled flag
	noColor = false
)

// DisableColor disables colored output
func DisableColor() {
	noColor = true
	color.NoColor = true
}

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Println("✓ " + msg)
	} else {
		fmt.Println(Success("✓") + " " + msg)
	}
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Fprintln(os.Stderr, "✗ "+msg)
	} else {
		fmt.Fprintln(os.Stderr, Error("✗")+" "+msg)
	}
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Println("⚠ " + msg)
	} else {
		fmt.Println(Warning("⚠") + " " + msg)
	}
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Println("ℹ " + msg)
	} else {
		fmt.Println(Info("ℹ") + " " + msg)
	}
}

// Verbose prints if verbose mode is enabled
func Verbose(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Printf("  "+format+"\n", args...)
	}
}
