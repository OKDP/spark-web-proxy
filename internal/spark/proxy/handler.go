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

package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/okdp/spark-history-web-proxy/internal/logging"
	"github.com/okdp/spark-history-web-proxy/internal/utils"
)

type ReverseProxyHandler interface {
	ModifyRequest(upstreamURL *url.URL) func(*http.Request)
	ModifyResponse() func(*http.Response) error
}

// DefaultErrorHandler returns a function that handles errors by logging the
// error details and sending an HTTP 502 (Bad Gateway) response with the error message.
//
// This error handler is typically used to handle proxy errors in situations where
// a service behind the proxy returns an unexpected error. The function logs the
// error with the URL path and details, and then responds with a standardized
// message to the client, indicating that a proxy error occurred.
//
// Parameters:
//   - rw: The `http.ResponseWriter` used to write the error response to the client.
//   - req: The incoming `http.Request` containing the original request details.
//   - err: The error that occurred during the request processing.
//
// Returns:
//   - A function of type `func(http.ResponseWriter, *http.Request, error)` that handles
//     the error and sends the appropriate response back to the client.
func DefaultErrorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Error("An error was occured when accessing url: %s, \ndetails: %+v", req.URL.Path, err)
		http.Error(rw, fmt.Sprintf("Proxy Error: %s", err.Error()), http.StatusBadGateway)
	}
}

func SparkUIErrorHandler(fromURL *url.URL) func(http.ResponseWriter, *http.Request, error) {
	defaultHandler := DefaultErrorHandler()

	return func(rw http.ResponseWriter, req *http.Request, err error) {
		if strings.Contains(fromURL.Path, "/kill") && utils.IsBrowserRequest(req) {
			previousPage := utils.CleanKillURLPath(fromURL.Path)
			log.Info("A spark job or stage kill was received '%s', redirecting to previous page: %s", fromURL.Path, previousPage)
			rw.Header().Set("Location", previousPage)
			rw.WriteHeader(http.StatusFound)
			return
		}

		defaultHandler(rw, req, err)
	}
}
