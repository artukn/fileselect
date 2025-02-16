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

type FilePath struct {
	name        string
	message     string
	defaultPath string
	flagValue   *string
	response    string
	mustExist   bool
}

// type FilePath string

var Quiet *bool

var expectedArgs []*FilePath

// Open defines a request for a file path that should exist (unless a - is entered)
func Open(name, defaultPath, message string) *FilePath {
	fp := Create(name, defaultPath, message)
	fp.mustExist = true
	return fp
}

// Create defines a request for a file path that doesn't have to exist
func Create(name, defaultPath, message string) *FilePath {
	it := FilePath{
		name:        name,
		message:     message,
		defaultPath: defaultPath,
	}
	expectedArgs = append(expectedArgs, &it)
	return &it
}

func (fp *FilePath) AsFlag(message string) *FilePath {
	fp.flagValue = flag.String(fp.name, "", message)
	return fp
}

func (fp *FilePath) Value() string {
	return fp.response
}

// Usage is an overwritable function that is used to print command line usage help
var Usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", flag.CommandLine.Name())
	fmt.Fprintf(flag.CommandLine.Output(), "  %s [options]", filepath.Base(strings.ReplaceAll(flag.CommandLine.Name(), "\\", "/")))
	for _, ei := range expectedArgs {
		if ei.flagValue != nil {
			continue
		}
		fmt.Fprintf(flag.CommandLine.Output(), " [%s]", ei.name)
	}
	fmt.Fprintf(flag.CommandLine.Output(), "\n\n")
	flag.PrintDefaults()
	for _, ei := range expectedArgs {
		if ei.flagValue != nil {
			continue
		}
		fmt.Fprintf(flag.CommandLine.Output(), "  [%s]\n        %s (default \"%s\")\n", ei.name, ei.message, ei.defaultPath)
	}
}

// Parse replaces flag.Parse, to add in file selection parsing
func Parse() {
	if !flag.Parsed() {
		flag.Parse()
	}
	if Quiet == nil {
		f := false
		Quiet = &f
	}

	args := slices.Clone(flag.Args())
	for i, fp := range expectedArgs {
		if fp.flagValue != nil {
			args = slices.Insert(args, i, *fp.flagValue)
			continue
		}
		if len(args) <= i {
			args = append(args, "")
		}
	}
	for i := 0; i < len(expectedArgs); i++ {
		fp := expectedArgs[i]
		if args[i] != "" {
			fp.response = args[i]
		} else {
			if !*Quiet {
				if !getFile(fp) {
					i-- // retry this
					continue
				}
			}
		}

		if fp.response == "-" {
			continue
		}

		if fp.mustExist {
			_, err := os.Stat(fp.response)
			if err != nil {
				if len(args) > i {
					args[i] = "" // don't keep retrying this argument - ask for user input
				}
				fmt.Fprintf(os.Stderr, "File for %s must exist, but '%s' doesn't\n", fp.name, fp.response)
				if !*Quiet {
					i-- // retry this
					continue
				}
			}
		}
	}
}

func getFile(fp *FilePath) bool {
	fmt.Fprintf(os.Stderr, "%s. %s=[%s]: ", fp.message, fp.name, fp.defaultPath)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := strings.TrimSpace(scanner.Text())
	fp.response = text
	if fp.response == "" {
		fp.response = fp.defaultPath
		fp.defaultPath = "-"
		return true
	}

	if text[0] == '\'' && text[len(text)-1] == '\'' || text[0] == '"' && text[len(text)-1] == '"' {
		fp.response = (text)[1 : len(text)-1]
	}
	return true
}
