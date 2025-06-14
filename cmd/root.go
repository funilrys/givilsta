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

	rootCmd.Flags().StringSliceVarP(&whitelistFiles, "whitelist", "w", []string{}, "The whitelist file to use for the cleanup. Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&whitelistALLFiles, "whitelist-all", "a", []string{}, "The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'ALL' flag. Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&whitelistREGFiles, "whitelist-regex", "r", []string{}, "The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'REG' flag. Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&whitelistRZDBFiles, "whitelist-rzdb", "z", []string{}, "The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'RZDB' flag. Can be specified multiple times.")

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "The output file to write the cleaned up subjects to. If not specified, we will print to stdout.")

	rootCmd.Flags().BoolVarP(&handleComplement, "handle-complement", "c", false, `Whether to handle complements subjects or not.
	A complement subject is www.example.com when the subject is example.com - and vice-versa.
	This is useful for domains that have a 'www' subdomain and want them to be whitelisted when the domain
	(without 'wwww' prefix) is whitelist listed.`)

	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "error", "The log level to use. Can be one of: debug, info, warn, error.")
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
		if helpers.IsUrl(whitelistFile) {
			targetFileName := filepath.Join(dirName, fmt.Sprintf("whitelist-%d.list", index))

			logger.Debug("Fetching whitelist file from URL.", slog.String("file", whitelistFile))
			err := helpers.FetchURLToFile(whitelistFile, targetFileName)

			if err != nil {
				logger.Error("Error fetching whitelist file from URL.", slog.String("file", whitelistFile), slog.String("error", err.Error()))
				fmt.Printf("Error fetching whitelist file from URL '%s': %v\n", whitelistFile, err)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist file from URL.", slog.String("file", whitelistFile), slog.String("targetFile", targetFileName))
			helpers.IterFile(targetFileName, func(line string) {
				ruler.AddRule(line)
			})
		} else {
			if _, err := os.Stat(whitelistFile); os.IsNotExist(err) {
				logger.Error("Whitelist file does not exist.", slog.String("file", whitelistFile))
				fmt.Printf("Error: Whitelist file '%s' does not exist.\n", whitelistFile)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist file.", slog.String("file", whitelistFile))
			helpers.IterFile(whitelistFile, func(line string) {
				ruler.AddRule(line)
			})
		}
	}

	for index, whitelistALLFile := range whitelistALLFiles {
		if helpers.IsUrl(whitelistALLFile) {
			targetFileName := filepath.Join(dirName, fmt.Sprintf("whitelist-all-%d.list", index))

			logger.Debug("Fetching whitelist file from URL.", slog.String("file", whitelistALLFile))

			err := helpers.FetchURLToFile(whitelistALLFile, targetFileName)
			if err != nil {
				logger.Error("Error fetching whitelist file from URL.", slog.String("file", whitelistALLFile), slog.String("error", err.Error()))
				fmt.Printf("Error fetching whitelist ALL file from URL '%s': %v\n", whitelistALLFile, err)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist ALL file from URL.", slog.String("file", whitelistALLFile), slog.String("targetFile", targetFileName))
			helpers.IterFile(targetFileName, func(line string) {
				ruler.AddRuleWithFlag(line, givilsta.FlagAll)
			})
		} else {
			if _, err := os.Stat(whitelistALLFile); os.IsNotExist(err) {
				logger.Error("Whitelist ALL file does not exist.", slog.String("file", whitelistALLFile))
				fmt.Printf("Error: Whitelist ALL file '%s' does not exist.\n", slog.String("file", whitelistALLFile))
				os.Exit(1)
			}

			logger.Debug("Processing whitelist ALL file.", slog.String("file", whitelistALLFile))
			helpers.IterFile(whitelistALLFile, func(line string) {
				ruler.AddRuleWithFlag(line, givilsta.FlagAll)
			})
		}
	}

	for index, whitelistREGFile := range whitelistREGFiles {
		if helpers.IsUrl(whitelistREGFile) {
			targetFileName := filepath.Join(dirName, fmt.Sprintf("whitelist-regex-%d.list", index))

			fmt.Printf("Fetching whitelist REG file from URL: %s\n", whitelistREGFile)

			err := helpers.FetchURLToFile(whitelistREGFile, targetFileName)

			if err != nil {
				logger.Error("Error fetching whitelist REG file from URL.", slog.String("file", whitelistREGFile), slog.String("error", err.Error()))
				fmt.Printf("Error fetching whitelist REG file from URL '%s': %v\n", whitelistREGFile, err)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist REG file from URL.", slog.String("file", whitelistREGFile), slog.String("targetFile", targetFileName))

			helpers.IterFile(targetFileName, func(line string) {
				ruler.AddRuleWithFlag(line, givilsta.FlagReg)
			})
		} else {
			if _, err := os.Stat(whitelistREGFile); os.IsNotExist(err) {
				logger.Error("Whitelist REG file does not exist.", slog.String("file", whitelistREGFile))
				fmt.Printf("Error: Whitelist REG file '%s' does not exist.\n", whitelistREGFile)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist REG file.", slog.String("file", whitelistREGFile))
			helpers.IterFile(whitelistREGFile, func(line string) {
				ruler.AddRuleWithFlag(line, givilsta.FlagReg)
			})
		}
	}

	for index, whitelistRZDBFile := range whitelistRZDBFiles {
		if helpers.IsUrl(whitelistRZDBFile) {
			targetFileName := filepath.Join(dirName, fmt.Sprintf("whitelist-rzdb-%d.list", index))

			logger.Debug("Fetching whitelist RZDB file from URL.", slog.String("file", whitelistRZDBFile))

			err := helpers.FetchURLToFile(whitelistRZDBFile, targetFileName)
			if err != nil {
				logger.Error("Error fetching whitelist RZDB file from URL.", "file", whitelistRZDBFile, "error", err)
				fmt.Printf("Error fetching whitelist RZDB file from URL '%s': %v\n", whitelistRZDBFile, err)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist RZDB file from URL.", slog.String("file", whitelistRZDBFile), slog.String("targetFile", targetFileName))

			helpers.IterFile(targetFileName, func(line string) {
				ruler.AddRuleWithFlag(line, givilsta.FlagRzdb)
			})
		} else {
			if _, err := os.Stat(whitelistRZDBFile); os.IsNotExist(err) {
				logger.Error("Whitelist RZDB file does not exist.", slog.String("file", whitelistRZDBFile))
				fmt.Printf("Error: Whitelist RZDB file '%s' does not exist.\n", whitelistRZDBFile)
				os.Exit(1)
			}

			logger.Debug("Processing whitelist RZDB file.", slog.String("file", whitelistRZDBFile))
			helpers.IterFile(whitelistRZDBFile, func(line string) {
				ruler.AddRuleWithFlag(line, givilsta.FlagRzdb)
			})
		}
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
