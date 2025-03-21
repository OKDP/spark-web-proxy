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

package sparkclient

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/okdp/spark-web-proxy/internal/constants"
	restclient "github.com/okdp/spark-web-proxy/internal/discovery/resolvers/rest/client"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
)

type SparkHistoryAppsClient struct {
	*restclient.SparkHistoryClient
}

func NewSparkHistoryAppsClient(request *http.Request, sparkHistoryBaseURL string) (*SparkHistoryAppsClient, error) {
	client, err := restclient.NewSparkHistoryClient(request, sparkHistoryBaseURL)
	return &SparkHistoryAppsClient{
		client,
	}, err
}

func (c *SparkHistoryAppsClient) GetApplicationInfo(appID string) (*model.HistorySparkApp, error) {
	c.Request.URL.Path = fmt.Sprintf("%s/%s", constants.SparkHistoryAppsEndpoint, appID)

	log.Debug("Get the application status '%s' from URL: %s", appID, c.Request.URL.String())

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		log.Error("Failed to get status for application '%s' from URL %s: %w", appID, c.Request.URL.Path, err)
		return nil, err
	}

	return doResponse[model.HistorySparkApp](resp, appID)
}

func (c *SparkHistoryAppsClient) GetEnvironment(appID string) (*model.HistorySparkAppEnvironment, error) {
	c.Request.URL.Path = fmt.Sprintf("%s/%s/%s", constants.SparkHistoryAppsEndpoint, appID, "environment")

	log.Debug("Get the application '%s' environment properties from URL: %s", appID, c.Request.URL.String())

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		log.Error("Failed to get environment properties for application '%s' from URL %s: %w", appID, c.Request.URL.String(), err)
		return nil, err
	}

	return doResponse[model.HistorySparkAppEnvironment](resp, appID)
}

func doResponse[T any](response *http.Response, appID string) (*T, error) {
	var object T
	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	// Read all decompressed content
	body, err := io.ReadAll(gzReader)
	if err != nil {
		log.Error("Failed to read body response for application '%s': %w", appID, err)
		return nil, err
	}

	err = json.Unmarshal([]byte(string(body)), &object)
	if err != nil {
		log.Error("Failed to parse body response for application '%s': %w", appID, err)
	}

	// Re-compress the modified body
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err = gzWriter.Write(body)
	if err != nil {
		log.Error("Failed to write body response for application '%s': %w", appID, err)
		return nil, err
	}
	err = gzWriter.Close()
	if err != nil {
		log.Error("Failed to close body response for application '%s': %w", appID, err)
		return nil, err
	}
	response.Body = io.NopCloser(&buf)
	response.ContentLength = int64(buf.Len())
	response.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	response.Header.Set("Content-Encoding", "gzip")

	return &object, nil
}
