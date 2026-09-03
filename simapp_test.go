// Copyright (c) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"testing"
)

func TestDispatchGroupExpandsImsiRange(t *testing.T) {
	SimappConfig = Config{Configuration: &Configuration{ConfigSliceDevGroup: true}}
	group := &DevGroup{
		Name:      "range-group",
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
	SimappConfig = Config{Configuration: &Configuration{ConfigSliceDevGroup: true}}
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
	SimappConfig = Config{Configuration: &Configuration{ConfigSliceDevGroup: true}}
	group := &DevGroup{
		Name:        "range-group",
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
	SimappConfig = Config{Configuration: &Configuration{ConfigSliceDevGroup: true}}
	group := &DevGroup{
		Name:        "range-group",
		Msisdns:     []string{"msisdn-9999999999"},
		MsisdnStart: "msisdn-9000000001",
		MsisdnEnd:   "msisdn-9000000003",
	}
	configMessages := make(chan configMessage, 1)

	dispatchGroup(configMessages, group, add_op, nil)

	select {
	case message := <-configMessages:
		var dispatchedGroup DevGroup
		if err := json.Unmarshal(message.msgPtr.Bytes(), &dispatchedGroup); err != nil {
			t.Fatalf("unmarshal dispatched group: %v", err)
		}

		expectedMsisdns := []string{"msisdn-9000000001", "msisdn-9000000002", "msisdn-9000000003"}
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
