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

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	log.Println("HTTP server listening on port", httpPort)
	if err := server.ListenAndServe(); err != nil {
		log.Println("Error starting HTTP server:", err)
	}
}
