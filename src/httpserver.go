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

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/apache/openserverless-streaming-proxy/handlers"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set CORS headers
		allowOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
		if allowOrigin == "" {
			allowOrigin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)

		allowMethods := os.Getenv("CORS_ALLOW_METHODS")
		if allowMethods == "" {
			allowMethods = "GET, POST, OPTIONS"
		}
		w.Header().Set("Access-Control-Allow-Methods", allowMethods)

		allowHeaders := os.Getenv("CORS_ALLOW_HEADERS")
		if allowHeaders == "" {
			allowHeaders = "*"
		}
		w.Header().Set("Access-Control-Allow-Headers", allowHeaders)

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startHTTPServer(streamingProxyAddr string, apihost string) {
	httpPort := os.Getenv("HTTP_SERVER_PORT")
	if httpPort == "" {
		httpPort = "80"
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Streamer proxy running"))
	})

	router.HandleFunc("GET /web/{ns}/{action}", handlers.WebActionStreamHandler(streamingProxyAddr, apihost))
	router.HandleFunc("GET /web/{ns}/{pkg}/{action}", handlers.WebActionStreamHandler(streamingProxyAddr, apihost))
	router.HandleFunc("GET /action/{ns}/{action}", handlers.ActionStreamHandler(streamingProxyAddr, apihost))
	router.HandleFunc("GET /action/{ns}/{pkg}/{action}", handlers.ActionStreamHandler(streamingProxyAddr, apihost))

	router.HandleFunc("POST /web/{ns}/{action}", handlers.WebActionStreamHandler(streamingProxyAddr, apihost))
	router.HandleFunc("POST /web/{ns}/{pkg}/{action}", handlers.WebActionStreamHandler(streamingProxyAddr, apihost))
	router.HandleFunc("POST /action/{ns}/{action}", handlers.ActionStreamHandler(streamingProxyAddr, apihost))
	router.HandleFunc("POST /action/{ns}/{pkg}/{action}", handlers.ActionStreamHandler(streamingProxyAddr, apihost))

	corsEnabled := os.Getenv("CORS_ENABLED")
	useCors := corsEnabled == "1" || corsEnabled == "true"

	server := &http.Server{
		Addr: ":" + httpPort,
		Handler: func() http.Handler {
			if useCors {
				return corsMiddleware(router)
			}
			return router
		}(),
	}

	log.Println("HTTP server listening on port", httpPort)
	if err := server.ListenAndServe(); err != nil {
		log.Println("Error starting HTTP server:", err)
	}
}
