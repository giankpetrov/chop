package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AgusRdz/chop/filters"
	"github.com/AgusRdz/chop/tee"
	"github.com/AgusRdz/chop/tracking"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: chop <command> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--version", "version":
		fmt.Printf("chop %s\n", version)
		return
	case "gain":
		runGain(os.Args[2:])
		return
	case "capture":
		runCapture(os.Args[2:])
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	filter := filters.Get(command, args)

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "chop: failed to run %s: %v\n", command, err)
			os.Exit(1)
		}
	}

	raw := string(output)
	fullCmd := command
	if len(args) > 0 {
		fullCmd = command + " " + strings.Join(args, " ")
	}

	var finalOutput string
	if filter != nil {
		filtered, ferr := filter(raw)
		if ferr != nil {
			finalOutput = raw
		} else {
			finalOutput = filtered
		}
	} else {
		// Auto-detect compression for unrecognized commands
		autoFiltered, aerr := filters.AutoDetect(raw)
		if aerr != nil || autoFiltered == raw {
			finalOutput = raw
		} else {
			finalOutput = autoFiltered
		}
	}

	fmt.Print(finalOutput)
	trackSilent(fullCmd, raw, finalOutput)

	// Tee: save raw output for LLM re-read
	rawTokens := tracking.CountTokens(raw)
	filteredTokens := tracking.CountTokens(finalOutput)
	savingsPct := 0.0
	if rawTokens > 0 {
		savingsPct = 100.0 - (float64(filteredTokens)/float64(rawTokens)*100.0)
	}
	if path := tee.Save(fullCmd, raw, exitCode, savingsPct); path != "" {
		fmt.Fprintf(os.Stderr, "[full output: %s]\n", path)
	}

	os.Exit(exitCode)
}

func trackSilent(command, raw, filtered string) {
	rawTokens := tracking.CountTokens(raw)
	filteredTokens := tracking.CountTokens(filtered)
	_ = tracking.Track(command, rawTokens, filteredTokens)
}

func runCapture(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: chop capture <command> [args...]")
		os.Exit(1)
	}

	command := args[0]
	cmdArgs := args[1:]

	cmd := exec.Command(command, cmdArgs...)
	cmd.Stdin = os.Stdin

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "chop: failed to run %s: %v\n", command, err)
			os.Exit(1)
		}
	}

	raw := string(output)

	// Build fixture name
	subcommand := ""
	if len(cmdArgs) > 0 {
		subcommand = cmdArgs[0]
	}
	ts := time.Now().Format("20060102-150405")
	var baseName string
	if subcommand != "" {
		baseName = fmt.Sprintf("%s-%s-%s", command, subcommand, ts)
	} else {
		baseName = fmt.Sprintf("%s-%s", command, ts)
	}

	fixtureDir := filepath.Join("tests", "fixtures")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "chop: failed to create fixtures dir: %v\n", err)
		os.Exit(1)
	}

	rawPath := filepath.Join(fixtureDir, baseName+".txt")
	if err := os.WriteFile(rawPath, []byte(raw), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "chop: failed to write raw fixture: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "raw:      %s\n", rawPath)

	// Apply filter if available
	filter := filters.Get(command, cmdArgs)
	if filter != nil {
		filtered, ferr := filter(raw)
		if ferr == nil {
			filteredPath := filepath.Join(fixtureDir, baseName+".filtered.txt")
			if err := os.WriteFile(filteredPath, []byte(filtered), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "chop: failed to write filtered fixture: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "filtered: %s\n", filteredPath)
			}
			fmt.Print(filtered)
		} else {
			fmt.Print(raw)
		}
	} else {
		fmt.Print(raw)
	}

	os.Exit(exitCode)
}

func runGain(args []string) {
	showHistory := false
	for _, a := range args {
		if a == "--history" {
			showHistory = true
		}
	}

	if showHistory {
		records, err := tracking.GetHistory(20)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chop: failed to read history: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(tracking.FormatHistory(records))
		return
	}

	stats, err := tracking.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chop: failed to read stats: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(tracking.FormatGain(stats))
}
