package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func banner() {

	banner := color.New(color.FgYellow, color.BgMagenta, color.Bold, color.Italic)
	banner.Print("------------------------------------------------------------")
	fmt.Println()
	banner.Print("--------------------     WEBTEST     -----------------------")
	fmt.Println()
	banner.Print("------------------------------------------------------------")
	fmt.Println()

	fmt.Println()
	fmt.Println("Legend:")
	printInfo("Information")
	printSuccess("No problem")
	printWarning("Minor problem")
	printFailure("Problem")
	printError("Major problem")
}

func trailer() {
	banner := color.New(color.FgGreen, color.Bold, color.Italic)

	fmt.Println()
	fmt.Println()
	banner.Print("-------------     🎉  TESTS COMPLETED  🎉    ---------------")
	fmt.Println()
	fmt.Println()
}

func printTestHeading(message string) {
	fmt.Println()
	color.Cyan("------------------------------------------------------------")
	color.Cyan("### " + message)
	color.Cyan("------------------------------------------------------------")
}

func printHeading(message string) {
	fmt.Println()
	color.HiMagenta("------------------------------------------------------------")
	color.HiMagenta("### " + message)
	color.HiMagenta("------------------------------------------------------------")
}

func printErrorHeading(message string) {
	fmt.Println()
	err := color.New(color.FgRed, color.BgYellow, color.BlinkRapid)
	err.Print(message)
	fmt.Println()
	fmt.Println()
}

func printSuccess(message string) {
	prefix := "    ✓ "
	for _, line := range strings.Split(message, "\n") {
		color.Green(prefix + line)
	}
}

func printInfo(message string) {
	prefix := "    - "
	for _, line := range strings.Split(message, "\n") {
		color.Blue(prefix + line)
	}
}

func printWarning(message string) {
	prefix := "    ! "
	for _, line := range strings.Split(message, "\n") {
		color.Yellow(prefix + line)
	}
}

func printFailure(message string) {
	prefix := "    ⨯ "
	for _, line := range strings.Split(message, "\n") {
		color.Red(prefix + line)
	}
}

func printError(message string) {
	prefix := "⚠ "
	for _, line := range strings.Split(message, "\n") {
		fmt.Print("    ")
		err := color.New(color.FgRed, color.BgYellow, color.BlinkRapid)
		err.Print(prefix + line)
		fmt.Println()
	}
}

func printTestResult(result testResult) {

	header := ""
	if result.file != "" && result.linenumber != "" {
		header = result.file + ":" + color.MagentaString(result.linenumber)
	} else if result.file != "" {
		header = result.file
	}

	faint := color.New(color.Faint)
	if header != "" {
		if result.result_type == testInfo {
			faint.Println(header)
		} else {
			fmt.Println(header)
		}
	}

	switch result.result_type {
	case testErr:
		printError(result.message)
	case testFail:
		printFailure(result.message)
	case testWarn:
		printWarning(result.message)
	case testSuccess:
		printSuccess(result.message)
	default:
		printInfo(result.message)
	}
}

func printTestResults(heading string, results []testResult) {

	printTestHeading(heading)

	if len(results) == 0 {
		printSuccess("Everything OK")
	}

	for _, result := range results {
		printTestResult(result)
	}
}
