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
package ruler

import (
	"log/slog"
	"regexp"
)

type InternalRuler struct {
	strict            map[string][]string
	ends              map[string][]string
	present           map[string][]string
	regex             string
	compiled_regexp   *regexp.Regexp
	handle_complement bool
	extensions        []string
	logger            *slog.Logger

	// Flags for different rule types
	FlagsAll     []string
	FlagsReg     []string
	FlagsRzdb    []string
	AllowedFlags []string
	// Default flag for each rule type
	FlagAll  string
	FlagReg  string
	FlagRzdb string
}
