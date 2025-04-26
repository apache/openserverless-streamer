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
	"log"
	"net/http"

	"github.com/apache/openserverless-streaming-proxy/tcp"
)

func ActionStreamHandler(streamingProxyAddr string, apihost string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, done := context.WithCancel(r.Context())

		namespace, actionToInvoke := getNamespaceAndAction(r)
		log.Printf("Private Action request: %s (%s)", actionToInvoke, namespace)

		apiKey, err := extractAuthToken(r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			done()
			return
		}

		// Create OpenWhisk client
		client := NewOpenWhiskClient(apihost, apiKey, namespace)

		// opens a socket for listening in a random port
		sock, err := tcp.SetupTcpServer(ctx, streamingProxyAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		enrichedBody, err := injectHostPortInBody(r, sock.Host, sock.Port)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		// invoke the action
		_, httpResp, err := client.Actions.Invoke(actionToInvoke, enrichedBody, false, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		// We need to handle status in the range from 200 to 299
		// as success, and everything else as an error.
		// In particular, we need to handle 202 Accepted
		// as a success, because the action is invoked
		// asynchronously and the response is not available yet.
		// We also need to handle 204 No Content as a success,
		// because the action is invoked and there is no response.
		// It seems that the invoker is releasing a 202 Accepted
		// after 60 seconds, so we need to handle that as well.
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			http.Error(w, "Error invoking action: "+httpResp.Status, http.StatusInternalServerError)
			done()
			return
		}

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
			}
		}
	}
}
