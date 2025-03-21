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

package model

import (
	"testing"
)

// TestIsRunning tests the IsRunning method of SparkAppInfo.
func TestIsRunning(t *testing.T) {
	tests := []struct {
		name     string
		app      HistorySparkApp
		expected bool
	}{
		{
			name: "Running application (Completed false)",
			app: HistorySparkApp{
				ID:   "spark-123",
				Name: "TestApp",
				Attempts: []HistorySparkAppAttempt{
					{Completed: false},
				},
			},
			expected: true,
		},
		{
			name: "Running application (Duration 0)",
			app: HistorySparkApp{
				ID:   "spark-456",
				Name: "TestApp",
				Attempts: []HistorySparkAppAttempt{
					{Completed: true, Duration: 0},
				},
			},
			expected: true,
		},
		{
			name: "Running application (EndTimeEpoch -1)",
			app: HistorySparkApp{
				ID:   "spark-789",
				Name: "TestApp",
				Attempts: []HistorySparkAppAttempt{
					{Completed: true, Duration: 100, EndTimeEpoch: -1},
				},
			},
			expected: true,
		},
		{
			name: "Completed application",
			app: HistorySparkApp{
				ID:   "spark-999",
				Name: "TestApp",
				Attempts: []HistorySparkAppAttempt{
					{Completed: true, Duration: 100, EndTimeEpoch: 1742487647315},
				},
			},
			expected: false,
		},
		{
			name: "No attempts",
			app: HistorySparkApp{
				ID:       "spark-000",
				Name:     "TestApp",
				Attempts: []HistorySparkAppAttempt{},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.app.IsRunning()
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}
