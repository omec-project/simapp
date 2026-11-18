// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"context"
	"os"

	"github.com/omec-project/simapp/internal/logger"
	"github.com/omec-project/simapp/internal/service"
	"github.com/urfave/cli/v3"
)

var SIMApp = &service.SIMApp{}

func main() {
	app := &cli.Command{}
	app.Name = "simapp"
	logger.SimappLog.Infoln(app.Name)
	app.Usage = "SIMApp"
	app.UsageText = "simapp -cfg <simapp_config_file.yaml>"
	app.Action = action
	app.Flags = SIMApp.GetCliCmd()
	if err := app.Run(context.Background(), os.Args); err != nil {
		logger.SimappLog.Fatalf("SIMApp run error: %v", err)
	}
}

func action(ctx context.Context, c *cli.Command) error {
	if err := SIMApp.Initialize(ctx, c); err != nil {
		return err
	}
	return SIMApp.Start()
}
