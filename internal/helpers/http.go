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
package helpers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// FetchURL fetches the content of the given URL and returns it as a string.
// Args:
//   - rawUrl: The URL to fetch.
//
// Returns:
//   - The content of the URL as a string.
func FetchURL(rawUrl string) (string, error) {
	resp, err := http.Get(rawUrl)

	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Panicf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// FetchURLToFile fetches the content of the given URL and writes it to a file.
// Args:
//   - rawUrl: The URL to fetch.
//   - filePath: The path to the file where the content will be written.
//
// Returns:
//   - An error if the fetch or write operation fails.
func FetchURLToFile(rawUrl, filePath string) error {
	resp, err := http.Get(rawUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Panicf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Panicf("error closing stream: %v", err)
		}
	}()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write response body to file: %w", err)
	}

	return nil
}

// IsUrl checks if the given string is a valid URL.
// Args:
//   - str: The string to check.
//
// Returns:
//   - true if the string is a valid URL, false otherwise.
func IsUrl(str string) bool {
	urlObj, err := url.Parse(str)

	if err != nil {
		return false
	}

	return urlObj.Scheme != "" && urlObj.Host != ""
}
