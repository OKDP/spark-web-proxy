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

type CachedSparkApp struct {
	BaseURL   string
	PodName   string
	AppID     string
	Namespace string
	Status    string
}

// SparkAppsStore holds a concurrent map of Spark applications, keyed by appId.
var (
	SparkAppsStore = struct {
		Instances sync.Map
	}{}
)

func (app CachedSparkApp) IsRunning() bool {
	return app.Status == string(AppRunning)
}

func (app CachedSparkApp) IsCompleted() bool {
	return !app.IsRunning()
}

// AddOrUpdateSparkApp adds a new SparkApp to the map or updates an existing one
func AddOrUpdateSparkApp(app *CachedSparkApp) {
	SparkAppsStore.Instances.Store(app.AppID, app)
}

// MakeSparkAppCompleted updates SparkApp to AppUnknown status
func MakeSparkAppCompleted(appID string) {
	app, found := GetSparkApp(appID)
	if found {
		app.Status = string(AppUnknown)
	} else {
		app = &CachedSparkApp{
			AppID:  appID,
			Status: string(AppUnknown),
		}
	}

	SparkAppsStore.Instances.Store(appID, app)
}

// Delete removes a SparkApp from the map
func DeleteSparkApp(appID string) {
	SparkAppsStore.Instances.Delete(appID)
}

// DeleteSparkAppByName removes a Spark application from the map by its PodName
// and returns the deleted SparkApp.
//
// It iterates over the sync.Map, finds the matching SparkApp by PodName,
// deletes the entry, and returns the deleted SparkApp.
//
// Parameters:
//   - podName: The name of the pod to be removed.
//
// Returns:
//   - (SparkApp, bool): The deleted SparkApp and a boolean indicating success.
//
// Example usage:
//
//	deletedApp, found := DeleteSparkAppByPodName("spark-pod-123")
//	if found {
//	    fmt.Println("Deleted SparkApp:", deletedApp)
//	}
func DeleteSparkAppByName(podName string) (*CachedSparkApp, bool) {
	var deletedApp *CachedSparkApp
	var found bool

	SparkAppsStore.Instances.Range(func(key, value interface{}) bool {
		if app, ok := value.(*CachedSparkApp); ok && app.PodName == podName {
			deletedApp = app
			found = true
			SparkAppsStore.Instances.Delete(key)
			return false
		}
		return true
	})

	return deletedApp, found
}

// Get retrieves a SparkApp from the map by appID
func GetSparkApp(appID string) (*CachedSparkApp, bool) {
	value, exists := SparkAppsStore.Instances.Load(appID)
	if exists {
		return value.(*CachedSparkApp), exists
	}
	return &CachedSparkApp{}, false
}

// ListSparkApps retrieves all SparkApps from the map
func ListSparkApps() []*CachedSparkApp {
	var apps []*CachedSparkApp
	SparkAppsStore.Instances.Range(func(_, value interface{}) bool {
		if app, ok := value.(CachedSparkApp); ok {
			apps = append(apps, &app)
		}
		return true
	})
	return apps
}

// GetProperty retrieves the value for the specified property name from the SparkProperties slice.
// It returns the value as a string and a boolean indicating whether the property was found.
//
// Parameters:
//   - propertyName (string): The name of the property to retrieve.
//
// Returns:
//   - (string, bool): The value of the property if found, and true. If the property is not found, it returns an empty string and false.
//
// Example:
//
//	response := SparkHistoryEnvironmentResponse{
//	    SparkProperties: [][]string{
//	        {"spark.acls.enable", "true"},
//	        {"spark.app.id", "spark-xyz123"},
//	    },
//	}
//	value, found := response.GetProperty("spark.app.id")
//	fmt.Println(value, found) // Output: "spark-xyz123 true"
func (r HistorySparkAppEnvironment) GetProperty(propertyName string) (string, bool) {
	for _, property := range r.SparkProperties {
		if property[0] == propertyName {
			return property[1], true
		}
	}
	return "_", false
}
