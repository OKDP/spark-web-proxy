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
	"sync"
)

type SparkApp struct {
	InternalURL string
	Namespace   string
	Status      string
}

var (
	SparkApps = struct {
		Instances sync.Map
	}{}
)

func (app SparkApp) IsRunning() bool {
	return app.Status == string(AppRunning)
}

func (app SparkApp) IsCompleted() bool {
	return !app.IsRunning()
}

// AddOrUpdateSparkApp adds a new SparkApp to the map or updates an existing one
func AddOrUpdateSparkApp(appID string, app SparkApp) {
	SparkApps.Instances.Store(appID, app)
}

// Delete removes a SparkApp from the map
func DeleteSparkApp(appID string) {
	SparkApps.Instances.Delete(appID)
}

// Get retrieves a SparkApp from the map by appID
func GetSparkApp(appID string) (SparkApp, bool) {
	value, exists := SparkApps.Instances.Load(appID)
	if exists {
		return value.(SparkApp), exists
	}
	return SparkApp{}, false
}

// ListSparkApps retrieves all SparkApps from the map
func ListSparkApps() []SparkApp {
	var apps []SparkApp
	SparkApps.Instances.Range(func(_, value interface{}) bool {
		if app, ok := value.(SparkApp); ok {
			apps = append(apps, app)
		}
		return true
	})
	return apps
}
