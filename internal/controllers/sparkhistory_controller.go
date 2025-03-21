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

package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/okdp/spark-history-web-proxy/internal/config"
	"github.com/okdp/spark-history-web-proxy/internal/constants"
	"github.com/okdp/spark-history-web-proxy/internal/discovery"
	log "github.com/okdp/spark-history-web-proxy/internal/logging"
	"github.com/okdp/spark-history-web-proxy/internal/model"
	"github.com/okdp/spark-history-web-proxy/internal/spark"
)

type SparkkHistoryController struct {
	sparkHistoryBaseURL string
	sparkHistoryBase    string
	sparkUIProxyBase    string
}

func NewSparkkHistoryController(config *config.ApplicationConfig) *SparkkHistoryController {
	controller := &SparkkHistoryController{
		sparkHistoryBaseURL: config.GetSparkHistoryBaseURL(),
		sparkHistoryBase:    constants.SparkHistoryBase,
		sparkUIProxyBase:    strings.TrimSpace(config.Spark.UI.ProxyBase),
	}

	log.Info("Spark History K8S Service URL: %s, Spark UI Proxy base: %s", controller.sparkHistoryBaseURL, controller.sparkUIProxyBase)
	return controller
}

func (r SparkkHistoryController) HandleHistoryApp(c *gin.Context) {

	appID := c.Param("appID")
	jobPath := c.Param("path")

	sparkApp, found := model.GetSparkApp(appID)

	// The application was started in cluster mode and is running
	if found && sparkApp.IsRunning() {
		r.redirectToSparkUI(c, appID)
		return
	}

	// The application was started in client or cluster mode and was not present locally
	if !found {
		log.Debug("The application '%s' was not found locally, checking in spark history ...", appID)
		sparkApp, _ := discovery.ResolveSparkAppFromHistory(c.Request, r.sparkHistoryBaseURL, appID)
		if sparkApp.IsRunning() {
			r.redirectToSparkUI(c, appID)
			return
		}
	}

	log.Debug("The application '%s' was started in client or cluster mode and is completed, forward to spark history: %s", appID, r.sparkHistoryBaseURL)

	upstreamURL, err := url.Parse(fmt.Sprintf("%s%s/%s%s", r.sparkHistoryBaseURL, r.sparkHistoryBase, appID, jobPath))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid upstream URL: %s", upstreamURL)})
		return
	}

	spark.ServeSparkHistory(c, upstreamURL, appID)
}

func (r SparkkHistoryController) HandleDefault(c *gin.Context) {
	path := c.Request.URL.Path

	upstreamURL, err := url.Parse(fmt.Sprintf("%s%s", r.sparkHistoryBaseURL, path))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid upstream URL: %s", upstreamURL)})
		return
	}

	spark.ServeSparkHistory(c, upstreamURL, "")
}

func (r SparkkHistoryController) redirectToSparkUI(c *gin.Context, appID string) {
	c.Request.URL.Path = fmt.Sprintf("%s/%s/jobs/", r.sparkUIProxyBase, appID)
	log.Debug("The application '%s' is running, redirect to spark ui '%s'", appID, c.Request.URL.String())
	c.Redirect(http.StatusFound, c.Request.URL.String())
}
