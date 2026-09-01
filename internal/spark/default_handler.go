/*
 *    Copyright 2025 The OKDP Authors.
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

// Package spark provides HTTP handlers and proxy wiring for serving Spark UI
// and Spark History endpoints through the reverse proxy.
package spark

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/spark/proxy"
)

// DefaultSparkHandler implements proxy.ReverseProxyHandler for Spark UI and
// Spark History requests.
type DefaultSparkHandler struct {
}

// NewDefaultSparkHandler creates a Spark reverse proxy configured with the
// default request/response rewriting behavior.
func NewDefaultSparkHandler(upstreamURL *url.URL, appID string) *proxy.SparkReverseProxy {
	return proxy.NewSparkReverseProxy(DefaultSparkHandler{}, upstreamURL, appID)
}

// ServeSparkHistory proxies Spark History requests to the configured upstream.
func ServeSparkHistory(c *gin.Context, upstreamURL *url.URL, appID string) {
	NewDefaultSparkHandler(upstreamURL, appID).
		ServeHTTP(c.Writer, c.Request)
}

// ServeSparkUI proxies Spark UI requests to the configured upstream and applies
// Spark UI–specific error handling (for redirects and fallback behavior).
func ServeSparkUI(c *gin.Context, upstreamURL *url.URL, appID string) {
	NewDefaultSparkHandler(upstreamURL, appID).
		WithSparkUIErrorHandler(c.Request.URL).
		ServeHTTP(c.Writer, c.Request)
}

// ModifyRequest returns a function that rewrites the incoming request URL to
// target the provided upstream URL.
func (c DefaultSparkHandler) ModifyRequest(upstreamURL *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = upstreamURL.Scheme
		req.URL.Host = upstreamURL.Host
		req.Host = upstreamURL.Host
		upstreamURL.RawQuery = req.URL.RawQuery
		upstreamURL.RawFragment = req.URL.RawFragment
		req.URL = upstreamURL
	}
}

// ModifyResponse returns a function that rewrites redirect Location headers so
// they remain relative when responses pass through the reverse proxy.
//
// Only redirects targeting the upstream are rewritten. A redirect to another
// origin, such as an authentication filter sending the browser to an identity
// provider, is passed through unchanged: stripping its scheme and host would
// make the browser resolve it against the proxy.
func (c DefaultSparkHandler) ModifyResponse(upstreamURL *url.URL) func(*http.Response) error {
	// Read here rather than in the closure: the director mutates upstreamURL
	// while serving a request.
	upstreamScheme, upstreamHost := upstreamURL.Scheme, upstreamURL.Host

	return func(resp *http.Response) error {
		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if location == "" {
				log.Warn("No Location header found in the response")
				return nil
			}
			parsedURL, err := url.Parse(location)
			if err != nil {
				log.Error("Error parsing Location URL: %+v", err)
				return nil
			}

			// Already relative: there is nothing to strip.
			if parsedURL.Host == "" {
				return nil
			}

			// A redirect to another origin keeps its scheme and host.
			if !sameOrigin(parsedURL, upstreamScheme, upstreamHost) {
				log.Debug("Location header '%s' targets another origin than the upstream '%s://%s', left untouched", location, upstreamScheme, upstreamHost)
				return nil
			}

			parsedURL.Scheme = ""
			parsedURL.Host = ""

			newLocation := parsedURL.String()
			resp.Header.Set("Location", newLocation)

			log.Debug("Rewritten Location Header: %s", newLocation)
			return nil
		}

		return nil
	}
}

// sameOrigin reports whether the redirect target and the upstream share the
// same origin, that is the same scheme, host and port. "http://spark-history"
// and "https://spark-history" are distinct origins, so a redirect to the
// latter must not be rewritten as if it were the former.
func sameOrigin(target *url.URL, upstreamScheme, upstreamHost string) bool {
	// A scheme-relative Location such as //spark-history:18080/jobs/ carries no
	// scheme of its own and adopts the one it is served over, here the scheme
	// of the upstream that produced the response.
	if target.Scheme == "" {
		target = &url.URL{Scheme: upstreamScheme, Host: target.Host}
	}

	upstream := &url.URL{Scheme: upstreamScheme, Host: upstreamHost}

	return strings.EqualFold(target.Scheme, upstream.Scheme) &&
		strings.EqualFold(target.Hostname(), upstream.Hostname()) &&
		portOrDefault(target) == portOrDefault(upstream)
}

// portOrDefault returns the port the URL is addressed on, falling back to the
// default port of its scheme when the authority leaves it out, so that
// "spark-history" and "spark-history:80" compare equal over http. Hostname and
// Port handle bracketed IPv6 authorities.
func portOrDefault(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	return schemePort(strings.ToLower(u.Scheme))
}

// schemePort returns the default port of a URI scheme, or an empty string when
// the scheme defines none.
//
// Only the HTTP and HTTPS default ports need special handling.
//
// A concrete Kubernetes example is a Spark History Service exposed as:
//
//	spec:
//	  ports:
//	    - port: 80
//	      targetPort: 18080
//
// With an HTTP upstream, GetSparkHistoryBaseURL builds
// "http://spark-history-server:80". Jetty may omit the default port in an
// absolute redirect and return "http://spark-history-server/history/...".
//
// The equivalent case exists for an HTTPS upstream explicitly configured on
// port 443. Without default-port normalization, these redirects would not be
// recognised as targeting the upstream and an internal cluster address could
// be exposed to the browser.
func schemePort(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}
