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

// JoinWithPipe joins elements of a slice into a single string separated by pipes ("|").
// If the slice is empty, it returns an empty string.
//
// This function is useful for creating a regex pattern or a similar structure.
//
// Args:
//
//	elements: A slice of strings to be joined.
//
// Returns:
//
//	A string with elements joined by pipes. If the slice is empty, returns an empty string.
func JoinWithPipe(elements []string) string {
	if len(elements) == 0 {
		return ""
	}

	result := elements[0]
	for _, elem := range elements[1:] {
		result += "|" + elem
	}
	return result
}
