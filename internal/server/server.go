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

	"github.com/okdp/spark-history-web-proxy/internal/config"
	log "github.com/okdp/spark-history-web-proxy/internal/logging"
)

func NewSparkUIProxyServer(config *config.ApplicationConfig) *http.Server {

	// Set up Gin router
	gin.SetMode(config.Proxy.Mode)
	r := gin.New()
	r.Use(log.Logger()...)
	r.Use(gin.Recovery())


	proxy := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", config.Proxy.ListenAddress, config.Proxy.Port),
	}

	return proxy
}
