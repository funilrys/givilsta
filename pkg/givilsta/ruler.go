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
package givilsta

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/funilrys/givilsta/internal/ruler"
)

// NewGivilstaRuler creates a new instance of our GivilstaRuler.
func NewGivilstaRuler(handle_complement bool, logger *slog.Logger) GivilstaRuler {
	intRuler := ruler.NewInternalRuler(handle_complement, logger)

	return &givilstaRuler{intRuler: intRuler, logger: logger}
}

// NewGivilstaRulerWithLogger creates a new instance of our GivilstaRuler with a logger.
func (g *givilstaRuler) Logger() *slog.Logger {
	return g.logger
}

// AddRule indexes a rule to the GivilstaRuler.
// Args:
//
//	rule: The rule to add.
//
// Returns:
//
//	bool: true if the rule was added successfully, false otherwise.
func (g *givilstaRuler) AddRule(rule string) bool {
	return g.intRuler.AddRule(rule)
}

// AddRuleWithFlag indexes a rule to the GivilstaRuler with a specific flag.
// Args:
//
//	rule: The rule to add.
//	flag: The flag to use for the rule.
//
// Returns:
//
//	bool: true if the rule was added successfully, false otherwise.
func (g *givilstaRuler) AddRuleWithFlag(rule string, flag Flags) bool {
	return g.intRuler.AddRule(fmt.Sprintf("%s%s", flag, rule))
}

// RemoveRule removes a rule from the GivilstaRuler.
// Args:
//
//	rule: The rule to remove.
//
// Returns:
//
//	bool: true if the rule was removed successfully, false otherwise.
func (g *givilstaRuler) RemoveRule(rule string) bool {
	return g.intRuler.RemoveRule(rule)
}

// RemoveRuleWithFlag removes a rule from the GivilstaRuler with a specific flag.
// Args:
//
//	rule: The rule to remove.
//	flag: The flag to use for the rule.
//
// Returns:
//
//	bool: true if the rule was removed successfully, false otherwise.
func (g *givilstaRuler) RemoveRuleWithFlag(rule string, flag Flags) bool {
	return g.intRuler.RemoveRule(fmt.Sprintf("%s%s", flag, rule))
}

// IsSubjectWhitelisted checks if a subject is whitelisted.
// Args:
//
//	subject: The subject to check.
//
// Returns:
//
//	bool: true if the subject is whitelisted, false otherwise.
func (g *givilstaRuler) IsSubjectWhitelisted(subject string) bool {
	return g.intRuler.IsWhitelisted(subject)
}

// IsSubjectBlacklisted checks if a subject is blacklisted.
// Args:
//
//	subject: The subject to check.
//
// Returns:
//
//	bool: true if the subject is blacklisted, false otherwise.
func (g *givilstaRuler) IsSubjectBlacklisted(subject string) bool {
	return !g.IsSubjectWhitelisted(subject)
}

// Same as IsSubjectWhitelisted, but assume that the given line come straight from
// one of the supported format: hosts file or plain text (maybe others in the future).
//
// Please note that this function will return a list of whitelisted subjects, because
// it assumes that the given line can contain multiple subjects separated by spaces.
//
// Args:
//
//	line: The line to check for whitelisted subjects.
//
// Returns:
//
//	[]string: A list of whitelisted subjects found in the line.
func (g *givilstaRuler) GetWhitelistedFromLine(line string) []string {
	normalized_line := strings.TrimSpace(line)

	if normalized_line == "" || strings.HasPrefix(normalized_line, "#") {
		return []string{}
	}

	if strings.Contains(normalized_line, "#") {
		normalized_line = normalized_line[:strings.Index(normalized_line, "#")]
	}

	subjects := strings.Fields(normalized_line)
	subjects = slices.Compact(subjects)

	var result []string

	for _, subject := range subjects {
		if g.IsSubjectWhitelisted(subject) {
			result = append(result, subject)
		}
	}

	return result
}

// Same as IsSubjectBlacklisted, but assume that the given line come straight from
// one of the supported format: hosts file or plain text (maybe others in the future).
//
// Please note that this function will return a list of blacklisted subjects, because
// it assumes that the given line can contain multiple subjects separated by spaces.
//
// Args:
//
//	line: The line to check for blacklisted subjects.
//
// Returns:
//
//	[]string: A list of blacklisted subjects found in the line.
func (g *givilstaRuler) GetBlacklistedFromLine(line string) []string {
	normalized_line := strings.TrimSpace(line)

	if normalized_line == "" || strings.HasPrefix(normalized_line, "#") {
		return []string{}
	}

	if strings.Contains(normalized_line, "#") {
		normalized_line = normalized_line[:strings.Index(normalized_line, "#")]
	}

	subjects := strings.Fields(normalized_line)
	subjects = slices.Compact(subjects)

	var result []string

	for _, subject := range subjects {
		if g.IsSubjectBlacklisted(subject) {
			result = append(result, subject)
		}
	}

	return result
}
