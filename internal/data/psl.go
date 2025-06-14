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
package data

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/funilrys/givilsta/internal/helpers"
)

func fetchPSLMapping() (map[string][]string, error) {
	var mapping map[string][]string

	mappingURL := "https://raw.githubusercontent.com/PyFunceble/public-suffix/master/public-suffix.json"

	data, err := helpers.FetchURL(mappingURL)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return mapping, nil
}

func NewPSLExtensions() *PSLExtensions {
	mapping, err := fetchPSLMapping()
	if err != nil {
		panic(fmt.Sprintf("failed to fetch public-suffix: %v", err))
	}

	var extensions = make([]string, 0, len(mapping))
	var suffixes []string

	for tld, tldSuffixes := range mapping {
		suffixes = append(suffixes, tldSuffixes...)
		extensions = append(extensions, tld)
	}

	suffixesRegexPattern := helpers.JoinWithPipe(suffixes)
	extensionsRegexPattern := helpers.JoinWithPipe(extensions)

	regexPattern := `(?i)^(` + suffixesRegexPattern + `|` + extensionsRegexPattern + `)$`

	regex := regexp.MustCompile(regexPattern)
	suffixesRegex := regexp.MustCompile(`(?i)^(` + suffixesRegexPattern + `)$`)
	extensionsRegex := regexp.MustCompile(`(?i)^(` + extensionsRegexPattern + `)$`)

	return &PSLExtensions{
		upstream:        mapping,
		Extensions:      suffixes,
		Suffixes:        extensions,
		Regex:           regex,
		SuffixesRegex:   suffixesRegex,
		ExtensionsRegex: extensionsRegex,
	}
}

func (fun *PSLExtensions) GetUpstream() map[string][]string {
	return fun.upstream
}

func (fun *PSLExtensions) GetExtensions() []string {
	return fun.Extensions
}

func (fun *PSLExtensions) GetSuffixes() []string {
	return fun.Suffixes
}

func (fun *PSLExtensions) GetRegex() *regexp.Regexp {
	return fun.Regex
}
