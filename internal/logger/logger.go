// SPDX-FileCopyrightText: 2024 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	utillogger "github.com/omec-project/util/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log         *zap.Logger
	SimappLog   *zap.SugaredLogger
	atomicLevel zap.AtomicLevel
)

func init() {
	log, atomicLevel = utillogger.NewZap()
	SimappLog = log.Sugar().With("component", "SIMApp", "category", "App")
}

// SetLogLevel: set the log level (panic|fatal|error|warn|info|debug)
func SetLogLevel(level zapcore.Level) {
	SimappLog.Infoln("set log level:", level)
	atomicLevel.SetLevel(level)
}
