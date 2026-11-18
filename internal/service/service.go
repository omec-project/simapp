// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/omec-project/simapp/internal/factory"
	"github.com/omec-project/simapp/internal/logger"
	"github.com/omec-project/simapp/internal/provision"
	"github.com/urfave/cli/v3"
)

type SIMApp struct {
	syncServer        *http.Server
	subProvisionEndpt factory.SubProvisionEndpt
}

var simappCli = []cli.Flag{
	&cli.StringFlag{
		Name:     "cfg",
		Usage:    "simapp config file",
		Required: true,
	},
}

func (*SIMApp) GetCliCmd() (flags []cli.Flag) {
	return simappCli
}

// Initialize loads the config file. A config that fails to load is logged and
// does not stop SIMApp, as before.
func (s *SIMApp) Initialize(_ context.Context, c *cli.Command) error {
	logger.SimappLog.Infoln("SIMApp started")
	provision.InitMsgChan()

	absPath, err := filepath.Abs(c.String("cfg"))
	if err != nil {
		logger.SimappLog.Errorln(err)
		return err
	}

	if err = factory.InitConfigFactory(absPath, &s.subProvisionEndpt); err != nil {
		logger.SimappLog.Errorln(err)
	}
	if factory.SimappConfig.Configuration != nil {
		provision.InitClient(factory.SimappConfig.Info)
	}
	return nil
}

func syncConfig(w http.ResponseWriter, req *http.Request) {
	_, err := fmt.Fprintf(w, "OK\n")
	if err != nil {
		logger.SimappLog.Errorln(err)
	}
	provision.DispatchAllGroups(provision.MsgChan, nil)
	provision.DispatchAllNetworkSlices(provision.MsgChan, nil)
}

// Start provisions the webui and then serves /synchronize until the process is terminated.
func (s *SIMApp) Start() error {
	// The /synchronize server, created up front so that a SIGTERM in any phase
	// can shut it down.
	s.syncServer = &http.Server{Addr: ":8080", ReadHeaderTimeout: 10 * time.Second}
	s.handleTermination()

	go provision.SendMessage(provision.MsgChan, s.subProvisionEndpt)
	go provision.WatchConfig()

	// Wait for webui to be ready before dispatching any messages.
	provision.WaitForWebui(s.subProvisionEndpt)

	// Provisioning must follow a strict order because the webconsole's sync logic
	// (triggered on device-group/network-slice creation) checks whether subscriber
	// auth data already exists in the DB. If subscribers aren't committed yet,
	// policy data is never written and UEs get REGISTRATION-REJECT.
	//
	// Phase 1: Subscribers (auth data must exist first)
	// Phase 2: Device groups (reference subscriber IMSIs)
	// Phase 3: Network slices (reference device groups; creation triggers final sync)

	// Phase 1: Subscribers
	var subscriberWg sync.WaitGroup
	provision.DispatchAllSubscribers(provision.MsgChan, &subscriberWg)
	subscriberWg.Wait()
	logger.SimappLog.Infoln("phase 1 complete: subscribers provisioned")

	// Phase 2: Device groups
	var groupWg sync.WaitGroup
	provision.DispatchAllGroups(provision.MsgChan, &groupWg)
	groupWg.Wait()
	logger.SimappLog.Infoln("phase 2 complete: device groups provisioned")

	// Phase 3: Network slices
	var sliceWg sync.WaitGroup
	provision.DispatchAllNetworkSlices(provision.MsgChan, &sliceWg)
	sliceWg.Wait()
	logger.SimappLog.Infoln("phase 3 complete: network slices provisioned")

	http.HandleFunc("/synchronize", syncConfig)
	err := s.syncServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		// Note: as per the `ListenAndServe` documentation: "ListenAndServe always returns a non-nil error."
		logger.SimappLog.Errorln(err)
	}
	for {
		time.Sleep(100 * time.Second)
	}
}

// handleTermination stops simapp on SIGTERM (a pod delete or a rollout) or
// Ctrl-C, in whatever phase it is: it shuts the /synchronize server down and
// exits 0.
func (s *SIMApp) handleTermination() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signals
		logger.SimappLog.Infof("received %v, terminating SIMApp", sig)
		s.Terminate()
		os.Exit(0)
	}()
}

// Terminate shuts the /synchronize server down, waiting at most 5 seconds for
// requests in flight.
func (s *SIMApp) Terminate() {
	stopServer(s.syncServer, 5*time.Second)
	logger.SimappLog.Infoln("SIMApp terminated")
}

// stopServer shuts the server down, waiting at most timeout for requests in
// flight. A server that was never started returns at once.
func stopServer(server *http.Server, timeout time.Duration) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.SimappLog.Warnf("/synchronize server shutdown: %v", err)
	}
}
