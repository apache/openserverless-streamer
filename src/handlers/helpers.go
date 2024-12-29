// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func injectHostPortInBody(r *http.Request, tcpServerHost string, tcpServerPort string) (map[string]interface{}, error) {
	body := r.Body
	defer body.Close()

	jsonBody := make(map[string]interface{})
	if err := json.NewDecoder(body).Decode(&jsonBody); err != nil {
		return nil, err
	}

	jsonBody["STREAM_HOST"] = tcpServerHost
	jsonBody["STREAM_PORT"] = tcpServerPort
	return jsonBody, nil
}

func getNamespaceAndAction(r *http.Request) (string, string) {
	namespace := r.PathValue("ns")
	pkg := r.PathValue("pkg")
	action := r.PathValue("action")

	actionToInvoke := action
	if pkg != "" {
		actionToInvoke = pkg + "/" + action
	}

	return namespace, actionToInvoke
}

func extractAuthToken(r *http.Request) (string, error) {
	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("Missing Authorization header")
	}

	// get the apikey without the Bearer prefix
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	return apiKey, nil
}

func ensureProtocolScheme(url string) string {
	if url == "" {
		return url
	}
	if strings.HasPrefix(url, "http://") {
		return url
	}
	if strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://" + url
}

func ensurePackagePresent(actionToInvoke string) string {
	if actionToInvoke == "" {
		return ""
	}
	if !strings.Contains(actionToInvoke, "/") {
		actionToInvoke = "default" + "/" + actionToInvoke
	}
	return actionToInvoke
}
