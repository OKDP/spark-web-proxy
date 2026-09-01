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

package spark

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/okdp/spark-web-proxy/internal/config"
	log "github.com/okdp/spark-web-proxy/internal/logging"
)

// TestMain installs the global logger the handlers under test log through. It
// stays nil until SetupGlobalLogger is called, which otherwise only happens
// when the command starts the server. The error level keeps the debug lines
// the rewriting emits out of the test output.
func TestMain(m *testing.M) {
	log.SetupGlobalLogger(config.Logging{Level: "error", Format: "console"})
	os.Exit(m.Run())
}

func TestDefaultSparkHandlerModifyResponse(t *testing.T) {
	const upstream = "http://spark-history.spark.svc.cluster.local:18080"

	upstreamURL, err := url.Parse(upstream)
	if err != nil {
		t.Fatalf("invalid upstream URL: %v", err)
	}

	tests := []struct {
		name     string
		status   int
		location string
		want     string
	}{
		{
			name:     "absolute redirect to the upstream is made relative",
			status:   http.StatusFound,
			location: upstream + "/history/app-123/jobs/",
			want:     "/history/app-123/jobs/",
		},
		{
			name:     "query string is preserved when the redirect is made relative",
			status:   http.StatusFound,
			location: upstream + "/history/app-123/jobs/?status=running&sort=%2Bid",
			want:     "/history/app-123/jobs/?status=running&sort=%2Bid",
		},
		{
			// A redirect to the bare upstream carries no path, so stripping
			// the authority leaves nothing behind. Spark addresses its
			// redirects to a path, so the two cases below record that
			// behaviour rather than guard against it.
			name:     "redirect to the bare upstream is left without a path",
			status:   http.StatusFound,
			location: upstream,
			want:     "",
		},
		{
			name:     "redirect to the bare upstream keeps only its query",
			status:   http.StatusFound,
			location: upstream + "?status=completed",
			want:     "?status=completed",
		},
		{
			name:     "relative redirect is left untouched",
			status:   http.StatusFound,
			location: "/history/app-456/stages/",
			want:     "/history/app-456/stages/",
		},
		{
			// Regression test for issue #33: the auth filter answers an
			// unauthenticated request with a 302 to the identity provider.
			// Stripping its scheme and host made the browser resolve the path
			// against the proxy, so the OIDC flow never started.
			name:     "cross origin redirect to the identity provider is preserved",
			status:   http.StatusFound,
			location: "https://keycloak.example.com/realms/okdp/protocol/openid-connect/auth?client_id=spark-history&response_type=code&redirect_uri=https%3A%2F%2Fspark-web-proxy.example.com%2Fhome",
			want:     "https://keycloak.example.com/realms/okdp/protocol/openid-connect/auth?client_id=spark-history&response_type=code&redirect_uri=https%3A%2F%2Fspark-web-proxy.example.com%2Fhome",
		},
		{
			name:     "same host but different scheme is a different origin",
			status:   http.StatusFound,
			location: "https://spark-history.spark.svc.cluster.local:18080/home/",
			want:     "https://spark-history.spark.svc.cluster.local:18080/home/",
		},
		{
			name:     "non redirect response is left untouched",
			status:   http.StatusOK,
			location: "https://keycloak.example.com/realms/okdp",
			want:     "https://keycloak.example.com/realms/okdp",
		},
		{
			name:     "status codes other than 302 are out of scope",
			status:   http.StatusMovedPermanently,
			location: upstream + "/history/",
			want:     upstream + "/history/",
		},
	}

	modify := DefaultSparkHandler{}.ModifyResponse(upstreamURL)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tc.status, Header: http.Header{}}
			resp.Header.Set("Location", tc.location)

			if err := modify(resp); err != nil {
				t.Fatalf("ModifyResponse returned an error: %v", err)
			}

			if got := resp.Header.Get("Location"); got != tc.want {
				t.Errorf("Location = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSameOrigin(t *testing.T) {
	const upstream = "http://spark-history:18080"

	tests := []struct {
		name     string
		target   string
		upstream string
		want     bool
	}{
		{
			name:     "identical origin",
			target:   upstream + "/home",
			upstream: upstream,
			want:     true,
		},
		{
			name:     "http default port spelled out on one side only",
			target:   "http://spark-history:80/home",
			upstream: "http://spark-history",
			want:     true,
		},
		{
			name:     "https default port spelled out on one side only",
			target:   "https://spark-history:443/home",
			upstream: "https://spark-history",
			want:     true,
		},
		{
			name:     "host comparison is case insensitive",
			target:   "http://Spark-History:18080/home",
			upstream: upstream,
			want:     true,
		},
		{
			name:     "scheme relative target inherits the upstream scheme",
			target:   "//spark-history:18080/home",
			upstream: upstream,
			want:     true,
		},
		{
			name:     "ipv6 authority",
			target:   "http://[fd00::1]:18080/home",
			upstream: "http://[fd00::1]:18080",
			want:     true,
		},
		{
			name:     "different scheme is a different origin",
			target:   "https://spark-history:18080/home",
			upstream: upstream,
			want:     false,
		},
		{
			name:     "different port is a different origin",
			target:   "http://spark-history:8080/home",
			upstream: upstream,
			want:     false,
		},
		{
			name:     "different host is a different origin",
			target:   "https://idp.example.com/realms/okdp",
			upstream: upstream,
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target, err := url.Parse(tc.target)
			if err != nil {
				t.Fatalf("invalid target URL: %v", err)
			}

			upstreamURL, err := url.Parse(tc.upstream)
			if err != nil {
				t.Fatalf("invalid upstream URL: %v", err)
			}

			got := sameOrigin(target, upstreamURL.Scheme, upstreamURL.Host)
			if got != tc.want {
				t.Errorf("sameOrigin(%q, upstream %q) = %v, want %v", tc.target, tc.upstream, got, tc.want)
			}
		})
	}
}
