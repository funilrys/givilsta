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

import "testing"

func TestIdnazeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"xn--ls8h.xn--ls8h", "xn--ls8h.xn--ls8h"},              // already IDNA ASCII
		{"saarbrücken.saarland", "xn--saarbrcken-feb.saarland"}, // IDNA conversion
		{"localhost", "localhost"},                              // should not change
	}

	for _, test := range tests {
		result := idnazeString(test.input)
		if result != test.expected {
			t.Errorf("idnazeString(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestIDnazeVeryLongString(t *testing.T) {
	longInput := "saarbrücken.saarland " + string(make([]byte, 1000)) // long input
	expected := "xn--saarbrcken-feb.saarland " + string(make([]byte, 1000))

	result := idnazeString(longInput)
	if result != expected {
		t.Errorf("idnazeString(long input) = %q; want %q", result, expected)
	}
}

func TestIdnaze(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"local", "local"},
		{"0.0.0.0", "0.0.0.0"},
		{"example.com", "example.com"},
		{"    example.org  ", "example.org"},
		{"xn--ls8h.xn--ls8h", "xn--ls8h.xn--ls8h"},
		{"saarbrücken.saarland", "xn--saarbrcken-feb.saarland"},
		{"localhost", "localhost"},
		{"# comment", "# comment"},
		{"# saarbrücken", "# saarbrücken"},
		{"saarbrücken.saarland # comment", "xn--saarbrcken-feb.saarland # comment"},
		{"saarbrücken.saarland # comment ### comment", "xn--saarbrcken-feb.saarland # comment ### comment"},
		{"0.0.0.0 saarbrücken.saarland", "0.0.0.0 xn--saarbrcken-feb.saarland"},
		{"saarbrücken.saarland   # comment", "xn--saarbrcken-feb.saarland   # comment"},
		{"saarbrücken.saarland\t# comment", "xn--saarbrcken-feb.saarland\t# comment"},
		{"saarbrücken.saarland\t    # comment\t### comment", "xn--saarbrcken-feb.saarland\t    # comment\t### comment"},
	}

	for _, test := range tests {
		result, err := idnaze(test.input)
		if err != nil {
			t.Errorf("idnaze(%q) returned error: %v", test.input, err)
			continue
		}
		if result != test.expected {
			t.Errorf("idnaze(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://saarbrücken.saarland", "http://xn--saarbrcken-feb.saarland"},
		{"https://saarbrücken.de", "https://xn--saarbrcken-feb.de"},
		{"ftp://saarbrücken.saarland", "ftp://xn--saarbrcken-feb.saarland"},
		{"http://localhost:8080", "http://localhost:8080"},
		{"http://saarbrücken.com/path?query=1", "http://xn--saarbrcken-feb.com/path?query=1"},
		{"//saarbrücken.com/path?query=1", "//xn--saarbrcken-feb.com/path?query=1"},
	}

	for _, test := range tests {
		result := NormalizeURL(test.input)
		if result != test.expected {
			t.Errorf("NormalizeURL(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestNormalizeSubjects(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"example.com", "example.com"},
		{"  example.org   ", "example.org"},
		{"xn--ls8h.xn--ls8h", "xn--ls8h.xn--ls8h"},
		{"saarbrücken.saarland", "xn--saarbrcken-feb.saarland"},
		{"www.example.com", "example.com"},
		{"localhost", "localhost"},
		{"# comment", ""},
		{"saarbrücken.saarland # comment", "xn--saarbrcken-feb.saarland"},
	}

	for _, test := range tests {
		result := NormalizeSubject(test.input)
		if result != test.expected {
			t.Errorf("normalizeSubjects(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestExctractNetLocationFromURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com", "example.com"},
		{"https://www.example.com/path?query=1", "www.example.com"},
		{"ftp://example.com/resource", "example.com"},
		{"http://localhost:8080", "localhost"},
		{"https://localhost", "localhost"},
		{"/path/to/resource", ""},
	}

	for _, test := range tests {
		result, err := ExtractNetLocationFromURL(test.input)
		if err != nil {
			t.Errorf("ExtractNetLocationFromURL(%q) returned error: %v", test.input, err)
			continue
		}
		if result != test.expected {
			t.Errorf("ExtractNetLocationFromURL(%q) = %q; want %q", test.input, result, test.expected)
		}
	}

}
