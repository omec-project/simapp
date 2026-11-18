// Copyright (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestStopServerEndsServing(t *testing.T) {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &http.Server{ReadHeaderTimeout: time.Second}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()

	stopServer(server, time.Second)

	select {
	case err := <-served:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("expected http.ErrServerClosed, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server still serving after stopServer")
	}
}

func TestStopServerNeverStartedReturnsAtOnce(t *testing.T) {
	server := &http.Server{Addr: "127.0.0.1:0", ReadHeaderTimeout: time.Second}
	done := make(chan struct{})
	go func() {
		stopServer(server, 5*time.Second)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stopServer blocked on a server that was never started")
	}
}
