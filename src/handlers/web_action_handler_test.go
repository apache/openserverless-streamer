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
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebActionHandler(t *testing.T) {
	streamingProxyAddr := "localhost"

	testMux := http.NewServeMux()
	testMux.HandleFunc("POST /api/v1/web/{ns}/{pkg}/{action}", func(w http.ResponseWriter, r *http.Request) {
		namespace := r.PathValue("ns")
		pkg := r.PathValue("pkg")
		action := r.PathValue("action")

		// read STREAM_HOST and STREAM_PORT from the body
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		jsonData := map[string]interface{}{}
		err := json.NewDecoder(buf).Decode(&jsonData)
		require.NoError(t, err)

		host, ok := jsonData["STREAM_HOST"].(string)
		require.True(t, ok)
		port, ok := jsonData["STREAM_PORT"].(string)
		require.True(t, ok)

		msg := fmt.Sprintf("Invoked action: %s/%s/%s", namespace, pkg, action)
		err = sendTcpSocketMsg(host, port, msg)
		require.NoError(t, err)

		w.Write([]byte("ok"))
	})
	ts := httptest.NewServer(testMux)

	realMux := http.NewServeMux()
	realMux.HandleFunc("POST /web/{ns}/{action}", WebActionStreamHandler(streamingProxyAddr, ts.URL))
	realMux.HandleFunc("POST /web/{ns}/{pkg}/{action}", WebActionStreamHandler(streamingProxyAddr, ts.URL))
	realMux.HandleFunc("GET /web/{ns}/{action}", WebActionStreamHandler(streamingProxyAddr, ts.URL))
	realMux.HandleFunc("GET /web/{ns}/{pkg}/{action}", WebActionStreamHandler(streamingProxyAddr, ts.URL))

	server := httptest.NewServer(realMux)

	defer server.Close()
	defer ts.Close()

	t.Run("post", func(t *testing.T) {
		body := []byte(`{}`)
		bodyReader := bytes.NewReader(body)

		resp, err := http.Post(server.URL+"/web/testns/testaction", "application/json", bodyReader)

		require.NoError(t, err)

		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		require.Equal(t, "Invoked action: testns/default/testaction\n", buf.String())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/web/testns/testaction")

		require.NoError(t, err)

		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		require.Equal(t, "Invoked action: testns/default/testaction\n", buf.String())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAsyncPostWebAction(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		body           []byte
		expectedErrMsg string
		handler        http.HandlerFunc
	}{
		{
			name: "Successful request",
			url:  "/success",
			body: []byte(`{"key": "value"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name:           "Error in request creation",
			url:            "1231",
			body:           []byte(`{"key": "value"}`),
			expectedErrMsg: "no such host",
		},
		{
			name: "Non-200 status code",
			url:  "/error",
			body: []byte(`{"key": "value"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedErrMsg: "not ok (500 Internal Server Error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errChan := make(chan error, 1)

			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				defer server.Close()
				tt.url = server.URL + tt.url
			}

			asyncPostWebAction(errChan, tt.url, tt.body)
			select {
			case err := <-errChan:
				require.NotEmpty(t, tt.expectedErrMsg)
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			default:
			}
		})
	}
}

// sendTcpSocketMsg connects to a server at the given host and port, sends a message, and closes the connection.
func sendTcpSocketMsg(host, port, message string) error {
	address := net.JoinHostPort(host, port)

	// Connect to the server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to %s: %w", address, err)
	}
	defer conn.Close()

	// Write the message to the connection
	_, err = conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", address, err)
	}

	return nil
}
