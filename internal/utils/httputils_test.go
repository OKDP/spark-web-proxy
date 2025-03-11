/*
 *    Copyright 2025 okdp.io
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package utils

import (
	"testing"
)

// TestValidateURL tests the ValidateURL function to ensure it panics on invalid URLs
func TestValidateURL(t *testing.T) {
	// Given
	validURLs := []string{
		"https://example.com",
		"http://example.com",
		"ftp://example.com",
		"https://sub.example.com/path?query=1",
	}

	invalidURLs := []string{
		"invalid-url",
		"htp:/wrong.url",
		"",
		"example.com",
		"https://",
	}

	// Test valid URLs (should NOT panic)
	for _, url := range validURLs {
		t.Run("Valid: "+url, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ValidateURL(%q) panicked unexpectedly: %v", url, r)
				}
			}()
			ValidateURL(url, "The URL is not valid")
		})
	}

	// Test invalid URLs
	for _, url := range invalidURLs {
		t.Run("Invalid: "+url, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("ValidateURL(%q) did NOT panic as expected", url)
				}
			}()
			ValidateURL(url, "The URL is not valid")
		})
	}
}
