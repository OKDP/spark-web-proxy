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

package spark

import (
	"net/http"
	"net/url"

	log "github.com/okdp/spark-history-web-proxy/internal/logging"
	"github.com/okdp/spark-history-web-proxy/internal/spark/proxy"
)

type DefaultSparkHandler struct {
}

func NewDefaultSparkHandler(upstreamURL *url.URL) *proxy.SparkReverseProxy {
	return proxy.NewSparkReverseProxy(DefaultSparkHandler{}, upstreamURL)
}

func (c DefaultSparkHandler) ModifyRequest(upstreamURL *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = upstreamURL.Scheme
		req.URL.Host = upstreamURL.Host
		req.Host = upstreamURL.Host
		upstreamURL.RawQuery = req.URL.RawQuery
		upstreamURL.RawFragment = req.URL.RawFragment
		req.URL = upstreamURL
		log.Info("Request Method: %s, URL: %s, Host: %s, Headers: %v", req.Method, req.URL.String(), req.Host, req.Header)
	}
}

func (c DefaultSparkHandler) ModifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		//resp.Header.Set("X-Modified-By", "Go-Proxy")
		resp.TransferEncoding = []string{"identity"}
		log.Info("Response Status: %d, Headers: %v", resp.StatusCode, resp.Header)
		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if location == "" {
				log.Warn("No Location header found in the response")
				return nil
			}
			parsedURL, err := url.Parse(location)
			if err != nil {
				log.Error("Error parsing Location URL: %v", err)
				return nil
			}

			parsedURL.Scheme = ""
			parsedURL.Host = ""

			newLocation := parsedURL.String()
			resp.Header.Set("Location", newLocation)

			log.Info("Rewritten Location Header: %s", newLocation)
		}

		return nil
	}
}

func (c DefaultSparkHandler) ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Error("Proxy Error for: %+v", err)
	http.Error(w, "Proxy Error", http.StatusBadGateway)
}
