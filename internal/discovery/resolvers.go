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

package discovery

import (
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	sparkclient "github.com/okdp/spark-web-proxy/internal/discovery/resolvers/rest"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
	"github.com/okdp/spark-web-proxy/internal/utils"
)

func ResolveSparkAppFromPod(pod *corev1.Pod) (*model.CachedSparkApp, error) {
	sparkUIURL := fmt.Sprintf("http://%s:%d", pod.Status.PodIP, utils.GetSparkUIPort(pod))
	sparkApp := &model.CachedSparkApp{
		BaseURL:   sparkUIURL,
		PodName:   pod.Name,
		AppID:     utils.GetSparkAppID(pod),
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
	}

	model.AddOrUpdateSparkApp(sparkApp)

	return sparkApp, nil
}

func ResolveSparkAppFromHistory(request *http.Request, sparkHistoryBaseURL string, appID string) (*model.CachedSparkApp, error) {
	sparkClient, err := sparkclient.NewSparkHistoryAppsClient(request, sparkHistoryBaseURL)
	if err != nil {
		log.Error("Unable to create new spark history client:", err)
		return nil, err
	}
	appInfo, err := sparkClient.GetApplicationInfo(appID)
	if err != nil {
		log.Error("Unable to get spark application '%s' status from spark history, %w", appID, err)
		return &model.CachedSparkApp{
			Status: string(model.AppUnknown),
		}, err
	}
	sparkAppEnv, err := sparkClient.GetEnvironment(appID)
	if err != nil {
		log.Error("Get the application '%s' environment properties from spark history: %w", appID, err)
		return &model.CachedSparkApp{
			Status: string(model.AppUnknown),
		}, err
	}
	sparkDriverHost, _ := sparkAppEnv.GetProperty("spark.driver.host")
	sparkDriverPort, _ := sparkAppEnv.GetProperty("spark.ui.port")
	sparkAppID, _ := sparkAppEnv.GetProperty("spark.app.id")
	sparkAppName, _ := sparkAppEnv.GetProperty("spark.app.name")
	sparkAppNamespace, _ := sparkAppEnv.GetProperty("spark.kubernetes.namespace")
	sparkUIBaseURL := fmt.Sprintf("http://%s:%s", sparkDriverHost, sparkDriverPort)

	sparkApp := &model.CachedSparkApp{
		BaseURL:   sparkUIBaseURL,
		PodName:   sparkAppName,
		AppID:     sparkAppID,
		Namespace: sparkAppNamespace,
		Status:    string(model.AppUnknown),
	}

	if appInfo.IsRunning() {
		sparkApp.Status = string(model.AppRunning)
	} else {
		model.AddOrUpdateSparkApp(sparkApp)
	}
	return sparkApp, err
}
