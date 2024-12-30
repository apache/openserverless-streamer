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

package tcp

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandleConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamDataChan := make(chan []byte, 10)
	server := &SocketsServer{
		ctx:            ctx,
		StreamDataChan: streamDataChan,
	}

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleConnection(serverConn)
	}()

	// Write data to the client connection
	testData := []byte("test data")
	clientConn.Write(testData)

	// Read data from the streamDataChan
	select {
	case receivedData := <-streamDataChan:
		require.Equal(t, testData, receivedData)
	case <-time.After(1 * time.Second):
		require.Fail(t, "Timeout waiting for data")
	}

	// Cancel the context to stop the server
	cancel()
	wg.Wait()
}

func TestHandleConnectionContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	streamDataChan := make(chan []byte, 10)
	server := &SocketsServer{
		ctx:            ctx,
		StreamDataChan: streamDataChan,
	}

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleConnection(serverConn)
	}()

	// Cancel the context to stop the server
	cancel()
	wg.Wait()

	// Write data to the client connection after context is cancelled
	testData := []byte("test data")
	clientConn.Write(testData)

	// Ensure no data is read from the streamDataChan
	select {
	case <-streamDataChan:
		t.Error("Expected no data, but got some")
	case <-time.After(100 * time.Millisecond):
		// Expected case
	}
}

func TestSetupTcpServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamingProxyAddr := "localhost"
	server, err := SetupTcpServer(ctx, streamingProxyAddr)
	require.NoError(t, err)
	defer server.listener.Close()

	require.NotNil(t, server.ctx)
	require.NotNil(t, server.listener)
	require.NotNil(t, server.StreamDataChan)
	require.NotEmpty(t, server.Host)
	require.NotEmpty(t, server.Port)
}

func TestSetupTcpServerInvalidAddress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamingProxyAddr := "invalid address"
	_, err := SetupTcpServer(ctx, streamingProxyAddr)
	require.Error(t, err)
}
