package fileselect

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	filepath "path"
	"slices"
	"strings"
)

type item struct {
	name        string
	message     string
	defaultPath string
	response    *string
	mustExist   bool
}

var Quiet *bool

var expectedItems []item

// Open defines a request for a file path that should exist (unless a - is entered)
func Open(name, message, defaultPath string) *string {
	return path(name, message, defaultPath, true)
}

// Create defines a request for a file path that doesn't have to exist
func Create(name, message, defaultPath string) *string {
	return path(name, message, defaultPath, false)
}

func path(name, message, defaultPath string, mustExist bool) *string {
	response := defaultPath
	expectedItems = append(expectedItems, item{
		name:        name,
		message:     message,
		defaultPath: defaultPath,
		response:    &response,
		mustExist:   mustExist,
	})
	return &response
}

// Usage is an overwritable function that is used to print command line usage help
var Usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", flag.CommandLine.Name())
	fmt.Fprintf(flag.CommandLine.Output(), "  %s [options]", filepath.Base(strings.ReplaceAll(flag.CommandLine.Name(), "\\", "/")))
	for _, ei := range expectedItems {
		fmt.Fprintf(flag.CommandLine.Output(), " [%s]", ei.name)
	}
	fmt.Fprintf(flag.CommandLine.Output(), "\n\n")
	flag.PrintDefaults()
	for _, ei := range expectedItems {
		fmt.Fprintf(flag.CommandLine.Output(), "  [%s]\n        %s (default \"%s\")\n", ei.name, ei.message, ei.defaultPath)
	}
}

// Parse replaces flag.Parse, to add in file selection parsing
func Parse() {
	flag.Usage = Usage
	if !flag.Parsed() {
		flag.Parse()
	}
	if Quiet == nil {
		f := false
		Quiet = &f
	}

	args := slices.Clone(flag.Args())
	for i := 0; i < len(expectedItems); i++ {
		item := &expectedItems[i]
		if len(args) > i && args[i] != "" {
			*item.response = args[i]
		} else {
			if !*Quiet {
				if !getFile(item) {
					i-- // retry this
					continue
				}
			}
		}

		if *item.response == "-" {
			continue
		}

		if item.mustExist {
			_, err := os.Stat(*item.response)
			if err != nil {
				if len(args) > i {
					args[i] = "" // don't keep retrying this argument - ask for user input
				}
				fmt.Fprintf(os.Stderr, "File for %s must exist, but '%s' doesn't\n", item.name, *item.response)
				if !*Quiet {
					i-- // retry this
					continue
				}
			}
		}
	}
}

func getFile(item *item) bool {
	fmt.Fprintf(os.Stderr, "%s. %s=[%s]: ", item.message, item.name, item.defaultPath)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := strings.TrimSpace(scanner.Text())
	*item.response = text
	if *item.response == "" {
		*item.response = item.defaultPath
		item.defaultPath = "-"
		return true
	}

	if text[0] == '\'' && text[len(text)-1] == '\'' || text[0] == '"' && text[len(text)-1] == '"' {
		*item.response = (text)[1 : len(text)-1]
	}
	return true
}
