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

package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/okdp/spark-history-web-proxy/internal/config"
	"github.com/okdp/spark-history-web-proxy/internal/constants"
	"github.com/okdp/spark-history-web-proxy/internal/controllers"
	"github.com/okdp/spark-history-web-proxy/internal/k8s/informers"
	log "github.com/okdp/spark-history-web-proxy/internal/logging"
)

func NewSparkUIProxyServer(config *config.ApplicationConfig) *http.Server {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("Failed to load Kubernetes in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatal("Failed to create Kubernetes client: %v", err)
	}

	informer := informers.NewSparkAppInformer(config)

	go informer.WatchSparkApps(clientset)

	// Set up Gin router
	gin.SetMode(config.Proxy.Mode)
	r := gin.New()
	r.Use(log.Logger()...)
	r.Use(gin.Recovery())

	// Spark UI
	sparkui := controllers.NewSparkUIController(config)
	sparkhistory := controllers.NewSparkkHistoryController(config)
	// Spark UI Handler
	r.Any(fmt.Sprintf("%s/:appID/*path", config.Spark.UI.ProxyBase), sparkui.HandleLiveApp)
	// Spark history Handlers
	r.Any("/history/:appID/*path", sparkhistory.HandleHistoryApp)
	r.Any("/static/*path", sparkhistory.HandleDefault)
	r.Any("/api/v1/applications", sparkhistory.HandleDefault)
	r.Any("/api/v1/applications/*path", sparkhistory.HandleDefault)
	r.Any("/history/", sparkhistory.HandleDefault)
	r.Any("/home/", sparkhistory.HandleDefault)
	r.Any("/jobs/", sparkhistory.HandleDefault)
	r.Any("/", sparkhistory.HandleDefault)

	r.GET(constants.HealthzURI, controllers.Healthz)
	r.GET(constants.ReadinessURI, controllers.Readiness)

	proxy := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", config.Proxy.ListenAddress, config.Proxy.Port),
	}

	return proxy
}
