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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			expectedErrMsg: "connect: no route to host",
		},
		{
			name: "Non-200 status code",
			url:  "/error",
			body: []byte(`{"key": "value"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedErrMsg: "Not OK (500 Internal Server Error)",
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

			go asyncPostWebAction(errChan, tt.url, tt.body)

			err := <-errChan
			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
