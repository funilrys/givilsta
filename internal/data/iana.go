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

func fetchIANAMapping() (map[string]*string, error) {
	var mapping map[string]*string

	mappingURL := "https://raw.githubusercontent.com/PyFunceble/iana/master/iana-domains-db.json"

	data, err := helpers.FetchURL(mappingURL)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return mapping, nil
}

func NewIANAExtensions() *IANAExtensions {
	mapping, err := fetchIANAMapping()
	if err != nil {
		panic(fmt.Sprintf("failed to fetch iana-domains-db: %v", err))
	}

	extensions := make([]string, 0, len(mapping))

	for extension := range mapping {
		extensions = append(extensions, extension)
	}

	regexPattern := `(?i)^(` + helpers.JoinWithPipe(extensions) + `)$`
	regex := regexp.MustCompile(regexPattern)

	return &IANAExtensions{
		upstream:   mapping,
		Extensions: extensions,
		Regex:      regex,
	}
}
