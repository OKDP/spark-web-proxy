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

// HistorySparkApp represents the structure of a Spark application info JSON response.
type HistorySparkApp struct {
	ID       string                   `json:"id"`
	Name     string                   `json:"name"`
	Attempts []HistorySparkAppAttempt `json:"attempts"`
}

// HistorySparkAppAttempt represents each attempt of the Spark application.
type HistorySparkAppAttempt struct {
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
	LastUpdated      string `json:"lastUpdated"`
	Duration         int64  `json:"duration"`
	SparkUser        string `json:"sparkUser"`
	Completed        bool   `json:"completed"`
	AppSparkVersion  string `json:"appSparkVersion"`
	StartTimeEpoch   int64  `json:"startTimeEpoch"`
	EndTimeEpoch     int64  `json:"endTimeEpoch"`
	LastUpdatedEpoch int64  `json:"lastUpdatedEpoch"`
}

// HistorySparkAppEnvironment represents the JSON structure for Spark history environment response (/applications/[app-id]/environment).
type HistorySparkAppEnvironment struct {
	SparkProperties [][]string `json:"sparkProperties"`
}

// IsRunning checks if the Spark application is still running.
// It returns true if at least one attempt meets any of the following conditions:
// 1. Completed is false
// 2. Duration is 0
// 3. EndTimeEpoch is -1
func (app HistorySparkApp) IsRunning() bool {
	for _, attempt := range app.Attempts {
		if !attempt.Completed ||
			attempt.Duration == 0 ||
			attempt.EndTimeEpoch == -1 {
			return true
		}
	}
	return false
}
