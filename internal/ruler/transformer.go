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
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
)

// idnazeString converts a subject to its IDNA ASCII representation.
//
// Args:
//
//	subject: The subject to convert.
//
// Note:
//
//	It returns the original subject if the conversion fails.
func idnazeString(subject string) string {
	result, err := idna.ToASCII(subject)

	if err != nil {
		return subject
	}
	return result
}

// idnaze processes a string, converting its subjects to IDNA ASCII representation.
// It handles both tab and space as separators, and preserves comments.
func idnaze(s string) (string, error) {
	s = strings.TrimSpace(s)

	var separator string
	tab := "\t"
	space := " "

	regex_to_skip, _ := regexp.Compile(`localhost$|localdomain$|local$|broadcasthost$|0\.0\.0\.0$|allhosts$|allnodes$|allrouters$|localnet$|loopback$|mcastprefix$`)

	if s == "" || strings.HasPrefix(s, "#") || regex_to_skip.MatchString(s) {
		return s, nil
	}

	if strings.Contains(s, tab) {
		separator = tab
	} else if strings.Contains(s, space) {
		separator = space
	} else {
		separator = ""
	}

	if separator != "" {
		var subjects string
		var idnazed_subjects []string

		comment := ""

		if strings.Contains(s, "#") {
			x := strings.SplitN(s, "#", 2)

			subjects, comment = x[0], x[1]
		} else {
			subjects = s
		}

		for _, subject := range strings.Split(subjects, separator) {
			if subject == "" || regex_to_skip.MatchString(subject) {
				idnazed_subjects = append(idnazed_subjects, subject)
				continue
			}

			if strings.Contains(subject, "#") {
				var current_comment string
				x := strings.SplitN(subject, "#", 1)

				if len(x) == 2 {
					subject, current_comment = x[0], x[1]
				} else {
					return "", fmt.Errorf("invalid format: %s", subject)
				}

				idnazed_subjects = append(idnazed_subjects, fmt.Sprintf("%s #%s", idnazeString(subject), current_comment))
			} else {
				idnazed_subjects = append(idnazed_subjects, idnazeString(subject))
			}
		}

		if comment != "" {
			return fmt.Sprintf("%s#%s", strings.Join(idnazed_subjects, separator), comment), nil
		} else {
			return strings.Join(idnazed_subjects, separator), nil
		}
	}

	return idnazeString(s), nil
}

// NormalizeURL normalizes a URL by converting its network location to IDNA ASCII representation.
// It handles both "http://" and "https://" prefixes, and returns the original URL if it cannot be normalized.
//
// Args:
//
//	urlStr: The URL string to normalize.
//
// Returns:
//
//	A normalized URL string, or the original URL if normalization fails.
func NormalizeURL(urlStr string) string {
	netloc, err := ExtractNetLocationFromURL(urlStr)

	if err != nil {
		return urlStr
	}

	idnazedNetloc, err := idnaze(netloc)

	if err != nil {
		return urlStr
	}

	return strings.Replace(urlStr, netloc, idnazedNetloc, 1)
}

// NormalizeSubject normalizes a subject for further processing.
//
// Args:
//
//		subject: The subject to normalize.
//	 complementHandling: Whether to handle complements (e.g., "www.example.com" vs "example.com" and vice versa).
//
// Returns:
//
//	A normalized subject string, or an empty string if the subject is invalid.
func NormalizeSubject(subject string, complementHandling bool) string {
	subject = strings.TrimSpace(subject)

	if subject == "" || strings.HasPrefix(subject, "#") {
		return ""
	}

	if strings.HasPrefix(subject, "http://") || strings.HasPrefix(subject, "https://") {
		return NormalizeURL(subject)
	}

	if strings.Contains(subject, "#") {
		subject = strings.TrimSpace(subject[:strings.Index(subject, "#")-1])
	}

	idnazedSubject, err := idnaze(subject)

	if err != nil {
		idnazedSubject = subject
	}

	if complementHandling {
		idnazedSubject = strings.TrimPrefix(idnazedSubject, "www.")
	}

	return idnazedSubject
}

// NormalizeRule normalizes a rule for further processing.
// It handles both URLs and plain text rules, converting them to IDNA ASCII representation.
//
// Args:
//
//	rule: The rule to normalize.
//
// Returns:
// A normalized rule string, or an empty string if the rule is invalid.
func NormalizeRule(rule string) string {
	rule = strings.TrimSpace(rule)

	if rule == "" || strings.HasPrefix(rule, "#") {
		return ""
	}

	if strings.HasPrefix(rule, "http://") || strings.HasPrefix(rule, "https://") {
		return NormalizeURL(rule)
	}

	if strings.Contains(rule, "#") {
		rule = strings.TrimSpace(rule[:strings.Index(rule, "#")-1])
	}

	idnazedRule, err := idnaze(rule)

	if err != nil {
		return rule
	}

	return idnazedRule
}

// ExtractNetLocationFromURL extracts the network location (host) from a given URL.
// If the URL does not contain a host, it returns the path.
// If the URL is empty or cannot be parsed, it returns an error.
//
// This function also removes any leading "http://" or "https://" and trailing path segments.
//
// Args:
//
//	rawURL: The URL string from which to extract the network location.
//
// Returns:
//
//	A string containing the network location or path, and an error if applicable.
func ExtractNetLocationFromURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	urlObj, err := url.Parse(rawURL)

	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	var result string

	if urlObj.Host == "" && urlObj.Path != "" {
		result = urlObj.Path
	} else if urlObj.Host != "" {
		if strings.Contains(urlObj.Host, ":") {
			// port filtering
			result = strings.Split(urlObj.Host, ":")[0]
		} else {
			result = urlObj.Host
		}
	} else {
		result = rawURL
	}

	if strings.Contains(result, "//") {
		result = result[strings.Index(result, "//")+2:]
	}

	if strings.Contains(result, "/") {
		result = result[:strings.Index(result, "/")]
	}

	return result, nil
}
