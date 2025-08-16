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
	"testing"
)

func testGetNewRuler() *InternalRuler {
	return NewInternalRuler(false, slog.Default())
}

func testGetNewRulerWithComplementsHandling() *InternalRuler {
	return NewInternalRuler(true, slog.Default())
}

func TestRuleHandling(t *testing.T) {
	ruler := testGetNewRuler()

	addRuleTests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"# comment", false},
		{"foo.example.com", true},
		{"all,example.com", true},
	}

	for _, test := range addRuleTests {
		result := ruler.AddRule(test.input)
		if result != test.expected {
			t.Errorf("AddRule(%q) = %v; want %v", test.input, result, test.expected)
		}
	}

	removeRuleTests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"# comment", false},
		{"foo.example.com", true},
		{"all,example.com", true},
	}

	for _, test := range removeRuleTests {
		result := ruler.RemoveRule(test.input)
		if result != test.expected {
			t.Errorf("RemoveRule(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestIsWhitelisted(t *testing.T) {
	ruler := testGetNewRuler()

	ruler.AddRule("foo.example.com")
	ruler.AddRule("https://saarbrücken.saarland/foo/bar")
	ruler.AddRule("ALL .org")
	ruler.AddRule("ALL .foo.saarlouis.de")
	ruler.AddRule("ALL foo")
	ruler.AddRule("REG vöklingen.*")
	ruler.AddRule("RZDB güter")

	tests := []struct {
		subject  string
		expected bool
	}{
		{"", false},
		{"example.com", false},
		{"bar.example.com", false},
		{"foo.example.com", true},
		{"bar.example.org", true},
		{"saarbrücken.org", true},
		{"https://saarbrücken.saarland", false},
		{"https://saarbrücken.saarland/foo/bar", true},
		{"vöklingen.de", true},
		{"vöklingen.com", true},
		{"vöklingen.org", true},
		{"https://vöklingen.org", true},
		{"https://vöklingen.com", true},
		{"https://vöklingen.de", true},
		{"https://vöklingen.de/foo/bar", true},
		{"https://vöklingen.org/bar/foo", true},
		{"güter.de", true},
		{"güter.com", true},
		{"güter.org", true},
		{"güter.net", true},
		{"https://güter.de", true},
		{"https://güter.com", true},
		{"https://güter.org", true},
		{"saarlouis.de", false},
		{"www.foo.saarlouis.de", true},
		{"foo.saarlouis.de", true},
		{"bar.foo.saarlouis.de", true},
		{"www.bar.foo.saarlouis.de", true},
		{"foo", true},
		{"bar.foo", true},
		{"foo.foo", true},
	}

	for _, test := range tests {
		result := ruler.IsWhitelisted(test.subject)
		if result != test.expected {
			t.Errorf("IsWhitelisted(%q) = %v; want %v", test.subject, result, test.expected)
		}
	}
}

func TestIsWhitelistedWithComplements(t *testing.T) {
	ruler := testGetNewRulerWithComplementsHandling()

	ruler.AddRule("foo.example.com")
	ruler.AddRule("https://saarbrücken.saarland/foo/bar")
	ruler.AddRule("ALL .org")
	ruler.AddRule("ALL .foo.saarlouis.de")
	ruler.AddRule("ALL foo")
	ruler.AddRule("RZDB www.güter")

	tests := []struct {
		subject  string
		expected bool
	}{
		{"", false},
		{"example.com", false},
		{"bar.example.com", false},
		{"foo.example.com", true},
		{"www.foo.example.com", true},
		{"bar.example.org", true},
		{"saarbrücken.org", true},
		{"https://saarbrücken.saarland", false},
		{"https://saarbrücken.saarland/foo/bar", true},
		{"https://www.saarbrücken.saarland/foo/bar", true},
		{"güter.de", true},
		{"www.güter.de", true},
		{"güter.com", true},
		{"www.güter.com", true},
		{"güter.org", true},
		{"www.güter.org", true},
		{"güter.net", true},
		{"www.güter.net", true},
		{"https://www.güter.de", true},
		{"https://güter.com", true},
		{"https://www.güter.org", true},
		{"saarlouis.de", false},
		{"www.foo.saarlouis.de", true},
		{"foo.saarlouis.de", true},
		{"bar.foo.saarlouis.de", true},
		{"www.bar.foo.saarlouis.de", true},
		{"foo", true},
		{"bar.foo", true},
		{"foo.foo", true}}

	for _, test := range tests {
		result := ruler.IsWhitelisted(test.subject)
		if result != test.expected {
			t.Errorf("IsWhitelisted(%q) = %v; want %v", test.subject, result, test.expected)
		}
	}
}

func TestHasFlag(t *testing.T) {
	ruler := testGetNewRuler()

	result := ruler.HasFlag([]string{"all "}, "all saarbrücken")

	if !result {
		t.Errorf("HasFlag() = true; want false")
	}

	result = ruler.HasFlag([]string{"all "}, "saarbrücken.saarland")

	if result {
		t.Errorf("HasFlag() = false; want true")
	}
}
