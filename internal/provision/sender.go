// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package provision

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/omec-project/simapp/internal/factory"
	"github.com/omec-project/simapp/internal/logger"
)

// SendMessage pushes the messages of msgChan to the webui.
func SendMessage(msgChan chan configMessage, subProvisionEndpt factory.SubProvisionEndpt) {
	var devGroupHttpend string
	var networkSliceHttpend string
	var subscriberHttpend string

	logger.SimappLog.Infof("subscriber provision endpoint in sendMessage: %s:%s", subProvisionEndpt.Addr, subProvisionEndpt.Port)

	ip := strings.TrimSpace(subProvisionEndpt.Addr)

	logger.SimappLog.Infoln("webui running at", ip)
	devGroupHttpend = httpProtocol + ip + ":" + subProvisionEndpt.Port + "/config/v1/device-group/"
	logger.SimappLog.Infoln("device trigger http endpoint", devGroupHttpend)
	networkSliceHttpend = httpProtocol + ip + ":" + subProvisionEndpt.Port + "/config/v1/network-slice/"
	logger.SimappLog.Infoln("network slice http endpoint", networkSliceHttpend)
	subscriberHttpend = httpProtocol + ip + ":" + subProvisionEndpt.Port + "/api/subscriber/imsi-"
	logger.SimappLog.Infoln("subscriber http endpoint", subscriberHttpend)

	// Create a worker pool with configurable max concurrent requests
	maxWorkers := factory.SimappConfig.Configuration.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = 50
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	logger.SimappLog.Infof("starting message processor with %d parallel workers", maxWorkers)

	for msg := range msgChan {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(msg configMessage) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			var httpend string
			logger.SimappLog.Debugln("received message from channel", msg)
			switch msg.msgType {
			case device_group:
				httpend = devGroupHttpend + msg.name
			case network_slice:
				httpend = networkSliceHttpend + msg.name
			case subscriber:
				httpend = subscriberHttpend + msg.name
			}
			var rsp *http.Response
			var httpErr error
			for {
				if msg.msgOp == add_op {
					logger.SimappLog.Infof("post message [%v] to %v", msg.String(), httpend)
					req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, httpend, msg.msgPtr)
					if err != nil {
						logger.SimappLog.Errorf("an error occurred %v", err)
						time.Sleep(1 * time.Second)
						continue
					}

					req.Header.Set("Content-Type", "application/json; charset=utf-8")
					rsp, httpErr = sendHttpReqMsg(req)
					if httpErr != nil {
						logger.SimappLog.Errorf("post message [%v] returned error [%v]", httpend, httpErr.Error())
					}

					logger.SimappLog.Infof("message POST %v success", rsp.StatusCode)
				} else if msg.msgOp == modify_op {
					logger.SimappLog.Infof("put message [%v] to %v", msg.String(), httpend)

					req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, httpend, msg.msgPtr)
					// Handle Error
					if err != nil {
						logger.SimappLog.Errorf("an error occurred %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					// set the request header Content-Type for json
					req.Header.Set("Content-Type", "application/json; charset=utf-8")
					rsp, httpErr = sendHttpReqMsg(req)
					if httpErr != nil {
						logger.SimappLog.Errorf("put message [%v] returned error [%v]", httpend, httpErr.Error())
					}

					logger.SimappLog.Infof("message PUT %v success", rsp.StatusCode)
				} else if msg.msgOp == delete_op {
					logger.SimappLog.Infof("delete message [%v] to %v", msg.String(), httpend)

					req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, httpend, msg.msgPtr)
					// Handle Error
					if err != nil {
						logger.SimappLog.Errorf("an error occurred %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					// set the request header Content-Type for json
					req.Header.Set("Content-Type", "application/json; charset=utf-8")
					rsp, httpErr = sendHttpReqMsg(req)
					if httpErr != nil {
						logger.SimappLog.Errorf("delete message [%v] returned error [%v]", httpend, httpErr.Error())
					}
					logger.SimappLog.Infof("message DEL %v success", rsp.StatusCode)
				}
				err := rsp.Body.Close()
				if err != nil {
					logger.SimappLog.Errorln(err)
				}
				if msg.wg != nil {
					msg.wg.Done()
				}
				break
			}
		}(msg)
	}

	// Wait for all goroutines to complete before exiting
	wg.Wait()
	logger.SimappLog.Infoln("all messages processed, message sender shutting down")
}
