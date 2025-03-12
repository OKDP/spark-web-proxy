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
	"net/http"
	"net/http/httputil"
	"net/url"
)

type SparkReverseProxy struct {
	*httputil.ReverseProxy
}

func NewSparkReverseProxy(c ReverseProxyHandler, upstreamURL *url.URL) *SparkReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	proxy.Director = c.ModifyRequest(upstreamURL)
	proxy.ModifyResponse = c.ModifyResponse()
	proxy.ErrorHandler = DefaultErrorHandler()
	return &SparkReverseProxy{proxy}
}

func (p *SparkReverseProxy) WithSparkUIErrorHandler(fromURL *url.URL) *SparkReverseProxy {
	p.ReverseProxy.ErrorHandler = SparkUIErrorHandler(fromURL)
	return p
}

func (p *SparkReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ReverseProxy.ServeHTTP(w, r)
}
