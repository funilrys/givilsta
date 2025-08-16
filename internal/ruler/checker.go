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
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/funilrys/givilsta/internal/data"
)

// Our internal constructor
func NewInternalRuler(handle_complement bool, logger *slog.Logger) *InternalRuler {
	var FlagsAll = []string{"ALL ", "ALL:", "ALL#", "ALL,", "ALL@"}
	var FlagsReg = []string{"REG ", "REG:", "REG#", "REG,", "REG@"}
	var FlagsRzdb = []string{"RZD ", "RZD:", "RZD#", "RZD,", "RZD@", "RZDB ", "RZDB:", "RZDB#", "RZDB,", "RZDB@"}

	var AllowedFlags = append(append([]string{}, FlagsAll...), append(FlagsReg, FlagsRzdb...)...)

	// ALL: the "ends-with" rule.
	var FlagAll = "ALL#"
	// REG: the regular expression rule.
	var FlagReg = "REG#"
	// RZDB: the RZDB rule.
	var FlagRzdb = "RZDB#"

	return &InternalRuler{
		strict:            make(map[string][]string),
		ends:              make(map[string][]string),
		present:           make(map[string][]string),
		regex:             "",
		compiled_regexp:   nil,
		extensions:        []string{},
		handle_complement: handle_complement,
		logger:            logger,

		// Shared flags for different rule types
		FlagsAll:     FlagsAll,
		FlagsReg:     FlagsReg,
		FlagsRzdb:    FlagsRzdb,
		AllowedFlags: AllowedFlags,
		// Default flag for each rule type
		FlagAll:  FlagAll,
		FlagReg:  FlagReg,
		FlagRzdb: FlagRzdb,
	}
}

// AddRule adds a rule to the whitelist checker.
//
// Args:
//
//	rule: The rule to add.
//
// Returns:
//
//	bool: true if the rule was added successfully, false otherwise.
func (fun *InternalRuler) AddRule(rule string) bool {
	normalizedRule := NormalizeRule(rule)

	logger := fun.logger.With(
		slog.String("rule", rule),
		slog.String("normalizedRule", normalizedRule),
	)
	logger.Debug("Adding rule")

	if normalizedRule == "" {
		logger.Debug("Rule is empty or a comment, skipping")
		return false
	}

	return fun.parseAllFlaggedRule(normalizedRule) || fun.parseRegexFlaggedRule(normalizedRule) || fun.parseRZDBFlagedRule(normalizedRule) || fun.parsePlainRule(normalizedRule)
}

// RemoveRule removes a rule from the whitelist checker.
//
// Args:
//
//	rule: The rule to remove.
//
// Returns:
//
//	bool: true if the rule was removed successfully, false otherwise.
func (fun *InternalRuler) RemoveRule(rule string) bool {
	normalizedRule := NormalizeRule(rule)

	logger := fun.logger.With(
		slog.String("rule", rule),
		slog.String("normalizedRule", normalizedRule),
	)
	logger.Debug("Removing rule")

	if normalizedRule == "" {
		logger.Debug("Rule is empty or a comment, skipping")
		return false
	}

	return fun.unparseAllFlaggedRule(normalizedRule) || fun.unparseRegexFlaggedRule(normalizedRule) || fun.unparseRZDBFlagedRule(normalizedRule) || fun.unparsePlainRule(normalizedRule)
}

func (fun *InternalRuler) IsWhitelisted(subject string) bool {
	normalizedSubject := NormalizeSubject(subject)

	logger := fun.logger.With(
		slog.String("subject", subject),
		slog.String("normalizedSubject", normalizedSubject),
	)
	logger.Debug("Checking subject")

	if normalizedSubject == "" {
		logger.Debug("Normalized subject is empty, skipping")

		return false
	}

	var subjects []string

	netloc, err := ExtractNetLocationFromURL(normalizedSubject)

	if err != nil {
		// If we cannot extract the net location, we consider the subject as a path.
		logger.Debug("Failed to extract net location.", slog.String("error", err.Error()))
		return false
	}

	subjects = append(subjects, netloc)

	if strings.HasPrefix(normalizedSubject, "http://") || strings.HasPrefix(normalizedSubject, "https://") {
		// We do this in order to handle the case that someone put an URL in the whitelist.
		subjects = append(subjects, normalizedSubject)
	}

	for _, sub := range subjects {
		commonKey := fun.commonSearchKeyFromRule(sub)

		if rules, ok := fun.strict[commonKey]; ok && slices.Contains(rules, sub) {
			logger.Debug("Subject found in strict rules", slog.String("extractedSubject", sub))
			return true
		}

		logger.Debug("Subject not found in strict rules. Continuing search", slog.String("extractedSubject", sub))

		if rules, ok := fun.present[commonKey]; ok && slices.Contains(rules, sub) {
			logger.Debug("Subject found in present rules", slog.String("extractedSubject", sub))
			return true
		}

		logger.Debug("Subject not found in present rules. Continuing search", slog.String("extractedSubject", sub))

		endKey := fun.endsSearchKeyFromRule(sub)

		if rules, ok := fun.ends[endKey]; ok {
			for _, rule := range rules {
				if strings.HasSuffix(sub, rule) {
					logger.Debug("Subject found in ends rules", slog.String("extractedSubject", sub), slog.String("rule", rule))
					return true
				}
			}
		}

		logger.Debug("Subject not found in ends rules. Continuing search", slog.String("extractedSubject", sub))

		if fun.compiled_regexp != nil && fun.compiled_regexp.MatchString(sub) {
			logger.Debug("Subject found in regex rules", slog.String("extractedSubject", sub))
			return true
		}

		logger.Debug("Subject not found in regex rules.", slog.String("extractedSubject", sub))
	}

	logger.Debug("Subject not matched any rule")

	return false
}

func (fun *InternalRuler) commonSearchKeyFromRule(rule string) string {
	if len(rule) < 4 {
		return rule
	}

	return rule[:4]
}

func (fun *InternalRuler) endsSearchKeyFromRule(rule string) string {
	if len(rule) < 3 {
		return rule
	}

	return rule[len(rule)-3:]
}

func (fun *InternalRuler) getKnownExtensions() []string {
	if len(fun.extensions) == 0 {
		fun.extensions = append(fun.extensions, data.NewIANAExtensions().Extensions...)
		fun.extensions = append(fun.extensions, data.NewPSLExtensions().Suffixes...)
	}

	return fun.extensions
}

func (fun *InternalRuler) pushStrictRule(rule string) {
	searchKey := fun.commonSearchKeyFromRule(rule)

	fun.strict[searchKey] = append(fun.strict[searchKey], rule)

	fun.logger.Debug("Pushed strict rule", slog.String("rule", rule), slog.String("searchKey", searchKey))
}

func (fun *InternalRuler) pullStrictRule(rule string) {
	searchKey := fun.commonSearchKeyFromRule(rule)

	if _, ok := fun.strict[searchKey]; ok {
		for i, r := range fun.strict[searchKey] {
			if r == rule {
				fun.strict[searchKey] = append(fun.strict[searchKey][:i], fun.strict[searchKey][i+1:]...)

				fun.logger.Debug("Pulled strict rule", slog.String("rule", rule), slog.String("searchKey", searchKey))
				break
			}
		}
	}
}

func (fun *InternalRuler) pushEndsRule(rule string) {
	searchKey := fun.endsSearchKeyFromRule(rule)

	fun.ends[searchKey] = append(fun.ends[searchKey], rule)

	fun.logger.Debug("Pushed ends rule", slog.String("rule", rule), slog.String("searchKey", searchKey))
}

func (fun *InternalRuler) pullEndsRule(rule string) {
	searchKey := fun.endsSearchKeyFromRule(rule)

	if _, ok := fun.ends[searchKey]; ok {
		for i, r := range fun.ends[searchKey] {
			if r == rule {
				fun.ends[searchKey] = append(fun.ends[searchKey][:i], fun.ends[searchKey][i+1:]...)

				fun.logger.Debug("Pulled ends rule", slog.String("rule", rule), slog.String("searchKey", searchKey))
				break
			}
		}
	}
}

func (fun *InternalRuler) pushRegexRule(rule string) {
	if fun.regex == "" {
		fun.regex = rule
	} else {
		fun.regex = fmt.Sprintf("%s|%s", fun.regex, rule)
	}

	if fun.compiled_regexp == nil {
		fun.compiled_regexp = regexp.MustCompile(fun.regex)
	} else {
		fun.compiled_regexp = regexp.MustCompile(fmt.Sprintf("%s|%s", fun.compiled_regexp.String(), rule))
	}

	fun.logger.Debug("Pushed regex rule", slog.String("rule", rule), slog.String("regexp", fun.regex))
}

func (fun *InternalRuler) pullRegexRule(rule string) {
	if fun.regex == "" {
		return
	}

	if fun.compiled_regexp == nil {
		return
	}

	fun.regex = strings.ReplaceAll(fun.regex, rule, "")

	if fun.regex == "" {
		fun.compiled_regexp = nil
	} else {
		fun.compiled_regexp = regexp.MustCompile(fun.regex)
	}

	fun.logger.Debug("Pulled regex rule", slog.String("rule", rule), slog.String("regexp", fun.regex))
}

func (fun *InternalRuler) HasFlag(flags []string, rule string) bool {
	for _, flag := range flags {
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(rule)), strings.ToLower(flag)) {
			return true
		}
	}
	return false
}

func (fun *InternalRuler) cleanupFlags(flags []string, rule string) string {
	for _, flag := range flags {
		if !fun.HasFlag([]string{flag}, rule) {
			continue
		}
		rule = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(rule, flag), strings.ToLower(flag)))
	}
	return rule
}

func (fun *InternalRuler) parseAllFlaggedRule(rule string) bool {
	if !fun.HasFlag(fun.FlagsAll, rule) {
		fun.logger.Debug("Rule does not match the ALL flags, skipping", slog.String("rule", rule))

		// Nothing to do.
		return false
	}

	record := fun.cleanupFlags(fun.FlagsAll, rule)

	if strings.HasPrefix(record, `.`) {
		if strings.Count(record, ".") > 1 {
			new_record := strings.TrimPrefix(record, ".")
			if fun.handle_complement {
				fun.pushStrictRule(fmt.Sprintf("www.%s", new_record))
			}
			fun.pushStrictRule(new_record)
		}
		fun.pushEndsRule(record)
	} else {
		fun.pushEndsRule(fmt.Sprintf(".%s", record))
		fun.pushStrictRule(record)
	}

	return true
}

func (fun *InternalRuler) unparseAllFlaggedRule(rule string) bool {
	if !fun.HasFlag(fun.FlagsAll, rule) {
		fun.logger.Debug("Rule does not match the ALL flags, skipping", slog.String("rule", rule))

		// Nothing to do.
		return false
	}

	record := fun.cleanupFlags(fun.FlagsAll, rule)

	if strings.HasPrefix(record, `.`) {
		if strings.Count(record, ".") > 1 {
			new_record := strings.TrimPrefix(record, ".")

			if fun.handle_complement {
				fun.pullStrictRule(fmt.Sprintf("www.%s", new_record))
			}
			fun.pullStrictRule(new_record)
		}
		fun.pullEndsRule(record)
	} else {
		// We except the record to starts with a dot.
		fun.pullEndsRule(fmt.Sprintf(".%s", record))
		fun.pullStrictRule(record)
	}

	return true
}

func (fun *InternalRuler) parseRegexFlaggedRule(rule string) bool {
	if !fun.HasFlag(fun.FlagsReg, rule) {
		fun.logger.Debug("Rule does not match the REG flags, skipping", slog.String("rule", rule))
		// Nothing to do.
		return false
	}

	fun.pushRegexRule(fun.cleanupFlags(fun.FlagsReg, rule))

	return true
}

func (fun *InternalRuler) unparseRegexFlaggedRule(rule string) bool {
	if !fun.HasFlag(fun.FlagsReg, rule) {
		fun.logger.Debug("Rule does not match the REG flags, skipping", slog.String("rule", rule))
		// Nothing to do.
		return false
	}

	fun.pullRegexRule(fun.cleanupFlags(fun.FlagsReg, rule))

	return true
}

func (fun *InternalRuler) parseRZDBFlagedRule(rule string) bool {
	if !fun.HasFlag(fun.FlagsRzdb, rule) {
		fun.logger.Debug("Rule does not match the RZDB flags, skipping", slog.String("rule", rule))
		// Nothing to do.
		return false
	}

	record := fun.cleanupFlags(fun.FlagsRzdb, rule)

	if fun.handle_complement && strings.HasPrefix(record, "www.") {
		record = strings.TrimPrefix(record, "www.")
	}

	if fun.handle_complement && strings.HasPrefix(record, "www.") {
		record = strings.TrimPrefix(record, "www.")
	}

	for _, extension := range fun.getKnownExtensions() {
		fun.pushStrictRule(fmt.Sprintf("%s.%s", record, extension))

		if fun.handle_complement {
			fun.pushStrictRule(fmt.Sprintf("www.%s.%s", record, extension))
		}
	}

	return true
}

func (fun *InternalRuler) unparseRZDBFlagedRule(rule string) bool {
	if !fun.HasFlag(fun.FlagsRzdb, rule) {
		fun.logger.Debug("Rule does not match the RZDB flags, skipping", slog.String("rule", rule))
		// Nothing to do.
		return false
	}

	record := fun.cleanupFlags(fun.FlagsRzdb, rule)

	if fun.handle_complement && strings.HasPrefix(record, "www.") {
		record = strings.TrimPrefix(record, "www.")
	}

	for _, extension := range fun.getKnownExtensions() {
		fun.pullStrictRule(fmt.Sprintf("%s.%s", record, extension))

		if fun.handle_complement {
			fun.pullStrictRule(fmt.Sprintf("www.%s.%s", record, extension))
		}
	}

	return true
}

func (fun *InternalRuler) parsePlainRule(rule string) bool {
	if fun.handle_complement {
		if strings.HasPrefix(rule, "http://") || strings.HasPrefix(rule, "https://") {
			netloc, err := ExtractNetLocationFromURL(rule)

			if err != nil {
				fun.logger.Debug("Failed to extract net location from rule", slog.String("rule", rule), slog.String("error", err.Error()))
				return false
			}

			if strings.HasPrefix(netloc, "www.") {
				fun.pushStrictRule(strings.ReplaceAll(rule, netloc, strings.TrimPrefix(netloc, "www.")))
			} else {
				fun.pushStrictRule(strings.ReplaceAll(rule, netloc, fmt.Sprintf("www.%s", netloc)))
			}
		} else {
			if strings.HasPrefix(rule, "www.") {
				fun.pushStrictRule(strings.TrimPrefix(rule, "www."))
			} else {
				fun.pushStrictRule(fmt.Sprintf("www.%s", rule))
			}
		}
	}

	fun.pushStrictRule(rule)

	return true
}

func (fun *InternalRuler) unparsePlainRule(rule string) bool {
	if fun.handle_complement {
		if strings.HasPrefix(rule, "http://") || strings.HasPrefix(rule, "https://") {
			netloc, err := ExtractNetLocationFromURL(rule)

			if err != nil {
				fun.logger.Debug("Failed to extract net location from rule", slog.String("rule", rule), slog.String("error", err.Error()))
				return false
			}

			if strings.HasPrefix(netloc, "www.") {
				fun.pullStrictRule(strings.ReplaceAll(rule, netloc, strings.TrimPrefix(netloc, "www.")))
			} else {
				fun.pullStrictRule(strings.ReplaceAll(rule, netloc, fmt.Sprintf("www.%s", netloc)))
			}
		} else {
			if strings.HasPrefix(rule, "www.") {
				fun.pullStrictRule(strings.TrimPrefix(rule, "www."))
			} else {
				fun.pullStrictRule(fmt.Sprintf("www.%s", rule))
			}
		}
	}

	fun.pullStrictRule(rule)

	return true
}
