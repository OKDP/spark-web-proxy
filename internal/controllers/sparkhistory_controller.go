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
	log "github.com/okdp/spark-history-web-proxy/internal/logging"
	"github.com/okdp/spark-history-web-proxy/internal/model"
	"github.com/okdp/spark-history-web-proxy/internal/spark"
	"github.com/okdp/spark-history-web-proxy/internal/utils"
)

type SparkkHistoryController struct {
	URL              string
	sparkHistoryBase string
	sparkUIProxyBase string
}

func NewSparkkHistoryController(config *config.ApplicationConfig) *SparkkHistoryController {
	controller := &SparkkHistoryController{
		URL: fmt.Sprintf("%s://%s:%d", config.Spark.History.Scheme,
			config.Spark.History.Service,
			config.Spark.History.Port),
		sparkHistoryBase: constants.SparkHistoryBase,
		sparkUIProxyBase: strings.TrimSpace(config.Spark.UI.ProxyBase),
	}
	utils.ValidateURL(controller.URL, fmt.Sprintf("The Spark History Server URL is not valid (Scheme: %s, Service: %s, Port: %d)",
		config.Spark.History.Scheme,
		config.Spark.History.Service,
		config.Spark.History.Port))

	log.Info("Spark History K8S Service URL: %s, Spark UI Proxy base: %s", controller.URL, controller.sparkUIProxyBase)
	return controller
}

func (r SparkkHistoryController) HandleHistoryApp(c *gin.Context) {

	appID := c.Param("appID")
	jobPath := c.Param("path")

	sparkApp, found := model.GetSparkApp(appID)

	if found && sparkApp.IsRunning() {
		c.Request.URL.Path = fmt.Sprintf("%s/%s/jobs/", r.sparkUIProxyBase, appID)
		log.Debug("The application ID '%s' is running, redirect to spark ui '%s'", appID, c.Request.URL.String())
		c.Redirect(http.StatusFound, c.Request.URL.String())
		return
	}

	upstreamURL, err := url.Parse(fmt.Sprintf("%s%s/%s%s", r.URL, r.sparkHistoryBase, appID, jobPath))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid upstream URL"})
		return
	}

	spark.
		NewDefaultSparkHandler(upstreamURL).
		ServeHTTP(c.Writer, c.Request)
}

func (r SparkkHistoryController) HandleDefault(c *gin.Context) {
	path := c.Request.URL.Path

	upstreamURL, err := url.Parse(fmt.Sprintf("%s%s", r.URL, path))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid upstream URL"})
		return
	}

	spark.
		NewDefaultSparkHandler(upstreamURL).
		ServeHTTP(c.Writer, c.Request)
}
