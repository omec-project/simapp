// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package provision

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/omec-project/simapp/internal/factory"
	"github.com/omec-project/simapp/internal/logger"
)

var client *http.Client

// InitClient sets the http client used to talk to the webui.
func InitClient(info *factory.Info) {
	httpVersion := 1
	if info != nil {
		httpVersion = info.HttpVersion
	}
	if httpVersion == 2 {
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     0,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		}
		// h2c requires HTTP1 to stay unset; setting it would make the transport fall back to HTTP/1.1
		transport.Protocols = new(http.Protocols)
		transport.Protocols.SetUnencryptedHTTP2(true)
		client = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	} else {
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     0,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		}
		client = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	}
}

func getNextBackoffInterval(retry, interval uint) uint {
	mFactor := 1.5
	nextInterval := float64(retry*interval) * mFactor

	if nextInterval > 10 {
		return 10
	}

	return uint(nextInterval)
}

func sendHttpReqMsg(req *http.Request) (*http.Response, error) {
	// Keep sending request to Http server until response is success
	var retries uint = 0
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.SimappLog.Errorln(err)
	}
	for {
		cloneReq := req.Clone(context.Background())
		req.Body = io.NopCloser(bytes.NewReader(body))
		cloneReq.Body = io.NopCloser(bytes.NewReader(body))
		rsp, err := client.Do(cloneReq)
		retries += 1
		if err != nil {
			if rsp != nil {
				if rsp.StatusCode == http.StatusConflict {
					logger.SimappLog.Infof("http response StatusConflict, skipping retry")
					err = req.Body.Close()
					if err != nil {
						logger.SimappLog.Errorln(err)
					}
					return rsp, nil
				}
			}
			nextInterval := getNextBackoffInterval(retries, 2)
			logger.SimappLog.Errorf("http req send error [%v], retrying after %d sec", err, nextInterval)
			time.Sleep(time.Second * time.Duration(nextInterval))
			continue
		}

		if rsp.StatusCode == http.StatusAccepted ||
			rsp.StatusCode == http.StatusOK || rsp.StatusCode == http.StatusNoContent ||
			rsp.StatusCode == http.StatusCreated {
			logger.SimappLog.Debugln("config push success")
			err = req.Body.Close()
			if err != nil {
				logger.SimappLog.Errorln(err)
			}
			return rsp, nil
		} else {
			if rsp.StatusCode == http.StatusConflict {
				logger.SimappLog.Infof("http response StatusConflict, skipping retry")
				err = req.Body.Close()
				if err != nil {
					logger.SimappLog.Errorln(err)
				}
				return rsp, nil
			}
			// 4xx client errors indicate a bad request that retrying won't fix
			if rsp.StatusCode >= 400 && rsp.StatusCode < 500 {
				logger.SimappLog.Errorf("http rsp client error [%s], not retrying", http.StatusText(rsp.StatusCode))
				err = req.Body.Close()
				if err != nil {
					logger.SimappLog.Errorln(err)
				}
				return rsp, fmt.Errorf("client error %d: %s", rsp.StatusCode, http.StatusText(rsp.StatusCode))
			}
			nextInterval := getNextBackoffInterval(retries, 2)
			logger.SimappLog.Infof("http rsp error [%s], retrying after %d sec", http.StatusText(rsp.StatusCode), nextInterval)
			err = rsp.Body.Close()
			if err != nil {
				logger.SimappLog.Infoln(err)
			}
			time.Sleep(time.Second * time.Duration(nextInterval))
		}
	}
}

func WaitForWebui(subProvisionEndpt factory.SubProvisionEndpt) {
	ip := strings.TrimSpace(subProvisionEndpt.Addr)
	readinessURL := httpProtocol + ip + ":" + subProvisionEndpt.Port + "/config/v1/device-group/"

	logger.SimappLog.Infoln("waiting for webui to be ready at", readinessURL)
	for {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, readinessURL, nil)
		if err != nil {
			logger.SimappLog.Errorf("failed to create readiness request: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		rsp, err := client.Do(req)
		if err != nil {
			logger.SimappLog.Infof("webui not ready: %v, retrying in 2 seconds...", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := rsp.Body.Close(); err != nil {
			logger.SimappLog.Errorf("failed to close readiness response body: %v", err)
		}
		if rsp.StatusCode >= 200 && rsp.StatusCode < 300 {
			logger.SimappLog.Infoln("webui is ready, starting message processing")
			return
		}
		logger.SimappLog.Infof("webui returned status %d, retrying in 2 seconds...", rsp.StatusCode)
		time.Sleep(2 * time.Second)
	}
}
