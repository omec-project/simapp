// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package factory

import (
	"os"

	"github.com/omec-project/simapp/internal/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.yaml.in/yaml/v4"
)

var SimappConfig Config

func InitConfigFactory(f string, subProvisionEndpt *SubProvisionEndpt) error {
	logger.SimappLog.Infoln("function called", f)
	if content, err := os.ReadFile(f); err != nil {
		logger.SimappLog.Infoln("readfile failed called", err)
		return err
	} else {
		SimappConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &SimappConfig); yamlErr != nil {
			logger.SimappLog.Errorln("yaml parsing failed", yamlErr)
			return yamlErr
		}
	}

	if SimappConfig.Logger.APP != nil {
		if SimappConfig.Logger.APP.DebugLevel != "" {
			if level, err := zapcore.ParseLevel(SimappConfig.Logger.APP.DebugLevel); err != nil {
				logger.SimappLog.Warnf("Simapp log level [%s] is invalid, set to [info] level",
					SimappConfig.Logger.APP.DebugLevel)
				logger.SetLogLevel(zap.InfoLevel)
			} else {
				logger.SetLogLevel(level)
			}
		} else {
			logger.SimappLog.Warnln("Simapp log level not set. Default set to [info] level")
			logger.SetLogLevel(zap.InfoLevel)
		}
	}

	if SimappConfig.Configuration == nil {
		logger.SimappLog.Infoln("configuration parsing failed", SimappConfig.Configuration)
		return nil
	}

	logger.SimappLog.Infof("subscriber provision endpoint: %s:%s", SimappConfig.Configuration.SubProvisionEndpt.Addr, SimappConfig.Configuration.SubProvisionEndpt.Port)
	subProvisionEndpt.Addr = SimappConfig.Configuration.SubProvisionEndpt.Addr
	subProvisionEndpt.Port = SimappConfig.Configuration.SubProvisionEndpt.Port

	viper.SetConfigFile(f)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return err
	}
	return nil
}
