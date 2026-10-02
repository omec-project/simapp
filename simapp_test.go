// Copyright (c) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestDispatchGroupExpandsImsiRange(t *testing.T) {
	SimappConfig = Config{Configuration: &Configuration{}}
	group := &DevGroup{
		Name:      "imsi-range-group",
		ImsiStart: "123456789123456",
		ImsiEnd:   "123456789123460",
	}
	configMessages := make(chan configMessage, 1)

	dispatchGroup(configMessages, group, add_op, nil)

	select {
	case message := <-configMessages:
		var dispatchedGroup DevGroup
		if err := json.Unmarshal(message.msgPtr.Bytes(), &dispatchedGroup); err != nil {
			t.Fatalf("unmarshal dispatched group: %v", err)
		}

		expectedImsis := []string{"123456789123456", "123456789123457", "123456789123458", "123456789123459", "123456789123460"}
		if len(dispatchedGroup.Imsis) != len(expectedImsis) {
			t.Fatalf("dispatched %d IMSIs, want %d", len(dispatchedGroup.Imsis), len(expectedImsis))
		}
		for index, expected := range expectedImsis {
			if dispatchedGroup.Imsis[index] != expected {
				t.Errorf("dispatched IMSI %d = %q, want %q", index, dispatchedGroup.Imsis[index], expected)
			}
		}
		if dispatchedGroup.ImsiStart != group.ImsiStart || dispatchedGroup.ImsiEnd != group.ImsiEnd {
			t.Errorf("dispatched IMSI range = %q-%q, want %q-%q", dispatchedGroup.ImsiStart, dispatchedGroup.ImsiEnd, group.ImsiStart, group.ImsiEnd)
		}
	default:
		t.Fatal("dispatchGroup did not enqueue a message")
	}
}

func TestDispatchGroupImsiRangeOverridesImsiList(t *testing.T) {
	SimappConfig = Config{Configuration: &Configuration{}}
	group := &DevGroup{
		Name:      "range-group",
		Imsis:     []string{"999999999999999"},
		ImsiStart: "123456789123456",
		ImsiEnd:   "123456789123458",
	}
	configMessages := make(chan configMessage, 1)

	dispatchGroup(configMessages, group, add_op, nil)

	select {
	case message := <-configMessages:
		var dispatchedGroup DevGroup
		if err := json.Unmarshal(message.msgPtr.Bytes(), &dispatchedGroup); err != nil {
			t.Fatalf("unmarshal dispatched group: %v", err)
		}

		expectedImsis := []string{"123456789123456", "123456789123457", "123456789123458"}
		if len(dispatchedGroup.Imsis) != len(expectedImsis) {
			t.Fatalf("dispatched %d IMSIs, want %d: %v", len(dispatchedGroup.Imsis), len(expectedImsis), dispatchedGroup.Imsis)
		}
		for index, expected := range expectedImsis {
			if dispatchedGroup.Imsis[index] != expected {
				t.Errorf("dispatched IMSI %d = %q, want %q", index, dispatchedGroup.Imsis[index], expected)
			}
		}

		if len(group.Imsis) != 1 || group.Imsis[0] != "999999999999999" {
			t.Errorf("configured IMSI list was mutated: %v", group.Imsis)
		}
	default:
		t.Fatal("dispatchGroup did not enqueue a message")
	}
}

func TestDispatchGroupExpandsMsisdnRange(t *testing.T) {
	SimappConfig = Config{Configuration: &Configuration{}}
	group := &DevGroup{
		Name:        "msisdn-range-group",
		MsisdnStart: "msisdn-9000000001",
		MsisdnEnd:   "msisdn-9000000005",
	}
	configMessages := make(chan configMessage, 1)

	dispatchGroup(configMessages, group, add_op, nil)

	select {
	case message := <-configMessages:
		var dispatchedGroup DevGroup
		if err := json.Unmarshal(message.msgPtr.Bytes(), &dispatchedGroup); err != nil {
			t.Fatalf("unmarshal dispatched group: %v", err)
		}

		expectedMsisdns := []string{"msisdn-9000000001", "msisdn-9000000002", "msisdn-9000000003", "msisdn-9000000004", "msisdn-9000000005"}
		if len(dispatchedGroup.Msisdns) != len(expectedMsisdns) {
			t.Fatalf("dispatched %d MSISDNs, want %d", len(dispatchedGroup.Msisdns), len(expectedMsisdns))
		}
		for index, expected := range expectedMsisdns {
			if dispatchedGroup.Msisdns[index] != expected {
				t.Errorf("dispatched MSISDN %d = %q, want %q", index, dispatchedGroup.Msisdns[index], expected)
			}
		}
		if dispatchedGroup.MsisdnStart != group.MsisdnStart || dispatchedGroup.MsisdnEnd != group.MsisdnEnd {
			t.Errorf("dispatched MSISDN range = %q-%q, want %q-%q", dispatchedGroup.MsisdnStart, dispatchedGroup.MsisdnEnd, group.MsisdnStart, group.MsisdnEnd)
		}
	default:
		t.Fatal("dispatchGroup did not enqueue a message")
	}
}

func TestDispatchGroupMsisdnRangeOverridesMsisdnList(t *testing.T) {
	SimappConfig = Config{Configuration: &Configuration{}}
	group := &DevGroup{
		Name:        "range-group",
		Msisdns:     []string{"msisdn-9999999999"},
		MsisdnStart: "msisdn-8000000001",
		MsisdnEnd:   "msisdn-8000000003",
	}
	configMessages := make(chan configMessage, 1)

	dispatchGroup(configMessages, group, add_op, nil)

	select {
	case message := <-configMessages:
		var dispatchedGroup DevGroup
		if err := json.Unmarshal(message.msgPtr.Bytes(), &dispatchedGroup); err != nil {
			t.Fatalf("unmarshal dispatched group: %v", err)
		}

		expectedMsisdns := []string{"msisdn-8000000001", "msisdn-8000000002", "msisdn-8000000003"}
		if len(dispatchedGroup.Msisdns) != len(expectedMsisdns) {
			t.Fatalf("dispatched %d MSISDNs, want %d: %v", len(dispatchedGroup.Msisdns), len(expectedMsisdns), dispatchedGroup.Msisdns)
		}
		for index, expected := range expectedMsisdns {
			if dispatchedGroup.Msisdns[index] != expected {
				t.Errorf("dispatched MSISDN %d = %q, want %q", index, dispatchedGroup.Msisdns[index], expected)
			}
		}

		if len(group.Msisdns) != 1 || group.Msisdns[0] != "msisdn-9999999999" {
			t.Errorf("configured MSISDN list was mutated: %v", group.Msisdns)
		}
	default:
		t.Fatal("dispatchGroup did not enqueue a message")
	}
}

func TestCompareGroupDetectsRangeChanges(t *testing.T) {
	tests := []struct {
		mutate func(*DevGroup)
		name   string
	}{
		{
			name: "imsi range",
			mutate: func(group *DevGroup) {
				group.ImsiEnd = "123456789123459"
			},
		},
		{
			name: "msisdn range",
			mutate: func(group *DevGroup) {
				group.MsisdnEnd = "msisdn-7000000004"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groupOld := &DevGroup{
				Imsis:       []string{"123456789123451"},
				ImsiStart:   "123456789123451",
				ImsiEnd:     "123456789123453",
				Msisdns:     []string{"msisdn-7000000001"},
				MsisdnStart: "msisdn-7000000001",
				MsisdnEnd:   "msisdn-7000000003",
			}
			groupNew := *groupOld
			test.mutate(&groupNew)

			if !compareGroup(&groupNew, groupOld) {
				t.Fatalf("compareGroup returned false for %s change", test.name)
			}
		})
	}
}

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
