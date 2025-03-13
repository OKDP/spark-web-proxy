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
)

type SparkUIController struct {
	sparkHistoryBase string
	sparkUIProxyBase string
}

func NewSparkUIController(config *config.ApplicationConfig) *SparkUIController {
	return &SparkUIController{
		sparkHistoryBase: constants.SparkHistoryBase,
		sparkUIProxyBase: strings.TrimSpace(config.Spark.UI.ProxyBase),
	}
}

func (r SparkUIController) HandleLiveApp(c *gin.Context) {
	appID := c.Param("appID")
	sparkAppPath := strings.TrimPrefix(c.Param("path"), "/")

	sparkApp, found := model.GetSparkApp(appID)
	
	if !found || sparkApp.IsCompleted() {
		c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, r.sparkUIProxyBase, r.sparkHistoryBase)
		log.Debug("The application ID '%s' was completed, redirect to spark history '%s'", appID, c.Request.URL.String())
		c.Redirect(http.StatusFound, c.Request.URL.String())
		return
	}

	upstreamURL, err := url.Parse(fmt.Sprintf("%s/%s", sparkApp.InternalURL, sparkAppPath))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid upstream URL"})
		return
	}

	if r.sparkUIProxyBase != "/proxy" {
		sparkUIRoot := fmt.Sprintf("%s/%s", r.sparkUIProxyBase, appID)
		log.Debug("Set spark UI root for application '%s' to: %s", appID, sparkUIRoot)
		c.Request.Header.Add("X-Forwarded-Context", sparkUIRoot)
	}

	spark.
		NewDefaultSparkHandler(upstreamURL).
		WithSparkUIErrorHandler(c.Request.URL).
		ServeHTTP(c.Writer, c.Request)
}
