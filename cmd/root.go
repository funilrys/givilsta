/*
Copyright © 2025 Nissar Chababy

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/funilrys/givilsta/internal/helpers"
	"github.com/funilrys/givilsta/pkg/givilsta"
	"github.com/spf13/cobra"
)

var ProjectVersion string
var sourceFile string
var outputFile string
var whitelistFiles []string
var whitelistALLFiles []string
var whitelistREGFiles []string
var whitelistRZDBFiles []string

var bypassFiles []string
var bypassALLFiles []string
var bypassREGFiles []string
var bypassRZDBFiles []string

var handleComplement bool
var logLevel string

var rootCmd = &cobra.Command{
	Use:   "givilsta",
	Short: "A different whitelisting mechanism for blocklist maintainers.",
	Long: `Givilsta is a tool designed to implement a different approach to whitelisting for maintainers of blocklists.

It tries to provide a more flexible and powerful approach to maintaining
whitelist lists for blocklist maintainers.`,

	Run: func(cmd *cobra.Command, args []string) {
		if sourceFile == "" {
			log.Fatal("Error: source must be specified.")
		}

		if len(whitelistFiles) == 0 && len(whitelistALLFiles) == 0 &&
			len(whitelistREGFiles) == 0 && len(whitelistRZDBFiles) == 0 {
			log.Fatal("Error: at least one whitelist file must be specified.")
		}

		var slogLevel slog.Level

		switch strings.ToLower(logLevel) {
		case "debug":
			slogLevel = slog.LevelDebug
		case "info":
			slogLevel = slog.LevelInfo
		case "warn":
			slogLevel = slog.LevelWarn
		case "error":
			slogLevel = slog.LevelError
		default:
			fmt.Fprintf(os.Stderr, "Warning: Unrecognized log-level '%s' from config. Defaulting to 'error'.\n", logLevel)
			slogLevel = slog.LevelError
		}

		logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slogLevel,
		}))
		slog.SetDefault(logger)

		processCleanup()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of your application",
	Long:  `All software has versions. This is your application's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Givilsta: %s\n", ProjectVersion)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().StringVarP(&sourceFile, "source", "s", "", "The source file to cleanup.")

	rootCmd.Flags().StringSliceVarP(&whitelistFiles, "whitelist", "w", []string{}, "The whitelist file to use for the cleanup.\nCan be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&whitelistALLFiles, "whitelist-all", "a", []string{}, "The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'ALL' flag.\nCan be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&whitelistREGFiles, "whitelist-regex", "r", []string{}, "The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'REG' flag.\nCan be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&whitelistRZDBFiles, "whitelist-rzdb", "z", []string{}, "The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'RZDB' flag.\nCan be specified multiple times.")

	rootCmd.Flags().StringSliceVarP(&bypassFiles, "bypass", "B", []string{}, `The bypass file to use for the cleanup. This file(s) is used to ensure that some some whitelisting rules are never applied.
Simply put any of the known rules in this file(s) and they will be ignored during the cleanup process.
Can be specified multiple times.`)
	rootCmd.Flags().StringSliceVarP(&bypassALLFiles, "bypass-all", "A", []string{}, "The bypass file to use for the cleanup. Any entries in this file-s will be prefixed with the 'ALL' flag.\nCan be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&bypassREGFiles, "bypass-regex", "R", []string{}, "The bypass file to use for the cleanup. Any entries in this file-s will be prefixed with the 'REG' flag.\nCan be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&bypassRZDBFiles, "bypass-rzdb", "Z", []string{}, "The bypass file to use for the cleanup. Any entries in this file-s will be prefixed with the 'RZDB' flag.\nCan be specified multiple times.")

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "The output file to write the cleaned up subjects to. If not specified, we will print to stdout.")

	rootCmd.Flags().BoolVarP(&handleComplement, "handle-complement", "c", false, `Whether to handle complements subjects or not.
A complement subject is www.example.com when the subject is example.com - and vice-versa.
is useful for domains that have a 'www' subdomain and want them to be whitelisted when the domain
without 'wwww' prefix is whitelist listed.`)

	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "error", "The log level to use. Can be one of: debug, info, warn, error.")
}

func processRuleFile(targetFile string, whitelistFlag givilsta.Flags, index int, ruler givilsta.GivilstaRuler, logger *slog.Logger, dirName string, bypass bool) {
	logger.Debug("Processing whitelist file.", slog.String("file", targetFile), slog.String("flag", string(whitelistFlag)))

	var targetFileName string
	if helpers.IsUrl(targetFile) {
		if !bypass {
			targetFileName = filepath.Join(dirName, fmt.Sprintf("whitelist-%s-%d.list", strings.ToLower(string(whitelistFlag)), index))
		} else {
			targetFileName = filepath.Join(dirName, fmt.Sprintf("bypass-%s-%d.list", strings.ToLower(string(whitelistFlag)), index))
			logger.Debug("Processing bypass file from URL.", slog.String("file", targetFile), slog.String("targetFile", targetFileName))
		}

		logger.Debug("Fetching file from URL.", slog.String("file", targetFile))
		err := helpers.FetchURLToFile(targetFile, targetFileName)

		if err != nil {
			logger.Error("Error fetching file from URL.", slog.String("file", targetFile), slog.String("error", err.Error()))
			fmt.Printf("Error fetching file from URL '%s': %v\n", targetFile, err)
			os.Exit(1)
		}

		logger.Debug("Processing file from URL.", slog.String("file", targetFile), slog.String("targetFile", targetFileName))

		helpers.IterFile(targetFileName, func(line string) {
			if !bypass {
				if whitelistFlag == givilsta.NoFlag {
					ruler.AddRule(line)
				} else {
					ruler.AddRuleWithFlag(line, whitelistFlag)
				}
			} else {
				if whitelistFlag == givilsta.NoFlag {
					ruler.RemoveRule(line)
				} else {
					ruler.RemoveRuleWithFlag(line, whitelistFlag)
				}
			}
		})
	} else {
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			logger.Error("Whitelist file does not exist.", slog.String("file", targetFile))
			fmt.Printf("Error: Whitelist file '%s' does not exist.\n", targetFile)
			os.Exit(1)
		}

		logger.Debug("Processing whitelist file.", slog.String("file", targetFile))

		helpers.IterFile(targetFile, func(line string) {
			if !bypass {
				if whitelistFlag == givilsta.NoFlag {
					ruler.AddRule(line)
				} else {
					ruler.AddRuleWithFlag(line, whitelistFlag)
				}
			} else {
				if whitelistFlag == givilsta.NoFlag {
					ruler.RemoveRule(line)
				} else {
					ruler.RemoveRuleWithFlag(line, whitelistFlag)
				}
			}
		})
	}
}

func processCleanup() {
	ruler := givilsta.NewGivilstaRuler(handleComplement, slog.Default())
	logger := ruler.Logger()

	dirName, err := os.MkdirTemp("", "givilsta")
	defer func() {
		if err := os.RemoveAll(dirName); err != nil {
			logger.Error("Error removing temporary directory.", slog.String("dir", dirName), slog.String("error", err.Error()))
			fmt.Printf("Error removing temporary directory '%s': %v\n", dirName, err)
			os.Exit(1)
		}
	}()

	if err != nil {
		log.Fatal("Failed to create temporary directory:", err)
	}

	for index, whitelistFile := range whitelistFiles {
		processRuleFile(whitelistFile, givilsta.NoFlag, index, ruler, logger, dirName, false)
	}

	for index, whitelistALLFile := range whitelistALLFiles {
		processRuleFile(whitelistALLFile, givilsta.FlagAll, index, ruler, logger, dirName, false)
	}

	for index, whitelistREGFile := range whitelistREGFiles {
		processRuleFile(whitelistREGFile, givilsta.FlagReg, index, ruler, logger, dirName, false)
	}

	for index, whitelistRZDBFile := range whitelistRZDBFiles {
		processRuleFile(whitelistRZDBFile, givilsta.FlagRzdb, index, ruler, logger, dirName, false)
	}

	for index, bypassFile := range bypassFiles {
		processRuleFile(bypassFile, givilsta.NoFlag, index, ruler, logger, dirName, true)
	}

	for index, bypassALLFile := range bypassALLFiles {
		processRuleFile(bypassALLFile, givilsta.FlagAll, index, ruler, logger, dirName, true)
	}

	for index, bypassREGFile := range bypassREGFiles {
		processRuleFile(bypassREGFile, givilsta.FlagReg, index, ruler, logger, dirName, true)
	}

	for index, bypassRZDBFile := range bypassRZDBFiles {
		processRuleFile(bypassRZDBFile, givilsta.FlagReg, index, ruler, logger, dirName, true)
	}

	if outputFile != "" {
		targetTempFile := filepath.Join(dirName, "output.list")

		helpers.WriteFileFromIter(targetTempFile, func(yield func(string)) {
			logger.Debug("Writing output to file.", slog.String("file", outputFile))
			helpers.IterFile(sourceFile, func(line string) {
				if strings.TrimSpace(line) != "" && ruler.IsSubjectBlacklisted(line) {
					yield(line)
				}
			})
		})

		// We do not have the guarantee that both temp and output files are in
		// the same filesystem, so we copy the temp file to the output file.
		err := helpers.CopyFile(targetTempFile, outputFile)

		if err != nil {
			logger.Error("Error copying temporary file to output file.", slog.String("tempFile", targetTempFile), slog.String("outputFile", outputFile), slog.String("error", err.Error()))
			fmt.Printf("Error copying temporary file '%s' to output file '%s': %v\n", targetTempFile, outputFile, err)
			os.Exit(1)
		}
	} else {
		logger.Debug("No output file specified, printing to stdout.")

		helpers.IterFile(sourceFile, func(line string) {
			if strings.TrimSpace(line) != "" && ruler.IsSubjectBlacklisted(line) {
				fmt.Println(line)
			}
		})
	}
}
