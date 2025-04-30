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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/apache/openserverless-streaming-proxy/tcp"
)

func WebActionStreamHandler(streamingProxyAddr string, apihost string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, done := context.WithCancel(r.Context())

		namespace, actionToInvoke := getNamespaceAndAction(r)
		log.Printf("Web Action requested: %s (%s)", actionToInvoke, namespace)

		// opens a socket for listening in a random port
		sock, err := tcp.SetupTcpServer(ctx, streamingProxyAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		// parse the json body and add STREAM_HOST and STREAM_PORT
		enrichedBody, err := injectHostPortInBody(r, sock.Host, sock.Port)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		// invoke the action
		actionToInvoke = ensurePackagePresent(actionToInvoke)

		jsonData, err := json.Marshal(enrichedBody)
		if err != nil {
			http.Error(w, "Error encoding JSON body: "+err.Error(), http.StatusInternalServerError)
			done()
			return
		}
		url := fmt.Sprintf("%s/api/v1/web/%s/%s", apihost, namespace, actionToInvoke)

		// Read headers and set them in the request
		headers := make(map[string]string)
		for key, values := range r.Header {
			// Use the first value for the header
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}

		errChan := make(chan error)
		defer close(errChan)
		go asyncPostWebAction(errChan, url, jsonData, headers)

		// Flush the headers
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			done()
			return
		}

		for {
			select {
			case data, isChannelOpen := <-sock.StreamDataChan:
				if !isChannelOpen {
					done()
					return
				}
				_, err := w.Write([]byte(string(data) + "\n"))
				if err != nil {
					http.Error(w, "failed to write data: "+err.Error(), http.StatusInternalServerError)
					done()
					return
				}
				flusher.Flush()

			case <-r.Context().Done():
				log.Println("HTTP Client closed connection")
				done()
				return

			case err := <-errChan:
				log.Println("Error invoking action:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				done()
				return
			}
		}
	}
}

func asyncPostWebAction(errChan chan error, url string, body []byte, headers map[string]string) {
	bodyReader := strings.NewReader(string(body))

	req, err := http.NewRequest("POST", ensureProtocolScheme(url), bodyReader)
	if err != nil {
		errChan <- err
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", bodyReader.Len()))
	req.ContentLength = int64(bodyReader.Len())

	// Aggiungi le intestazioni opzionali
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	httpResp, err := client.Do(req)
	if err != nil {
		errChan <- err
		return
	}

	// Gestione dello stato HTTP
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		errChan <- fmt.Errorf("not ok (%s)", httpResp.Status)
	} else {
		log.Printf("Received status code %d: %s", httpResp.StatusCode, httpResp.Status)
	}
}
