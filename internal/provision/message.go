// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package provision

import (
	"bytes"
	"fmt"
	"sync"
)

const (
	add_op = iota
	modify_op
	delete_op
)

const (
	device_group = iota
	network_slice
	subscriber
)

const httpProtocol = "http://"

type configMessage struct {
	msgPtr  *bytes.Buffer
	wg      *sync.WaitGroup
	name    string
	msgType int
	msgOp   int
}

func (msg configMessage) String() string {
	var msgType, msgOp string
	switch msg.msgOp {
	case add_op:
		msgOp = "add_op"
	case modify_op:
		msgOp = "modify_op"
	case delete_op:
		msgOp = "delete_op"
	}

	switch msg.msgType {
	case device_group:
		msgType = "device-group"
	case network_slice:
		msgType = "network-slice"
	case subscriber:
		msgType = "subscriber"
	}
	return fmt.Sprintf("config msg name [%v], type [%v], op [%v]", msg.name, msgType, msgOp)
}

// MsgChan carries the configuration messages from the dispatchers to SendMessage.
var MsgChan chan configMessage

// InitMsgChan creates MsgChan.
func InitMsgChan() { MsgChan = make(chan configMessage, 100) }
