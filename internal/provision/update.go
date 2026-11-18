// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package provision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/omec-project/simapp/internal/factory"
	"github.com/omec-project/simapp/internal/logger"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v4"
)

func isImsiInSubscribers(imsi int, subscribers []*factory.Subscriber) (bool, *factory.Subscriber) {
	for _, subscriber := range subscribers {
		start, err := strconv.Atoi(subscriber.UeIdStart)
		if err != nil {
			logger.SimappLog.Errorln("error in Atoi with UeIdStart", err)
			continue
		}
		end, err := strconv.Atoi(subscriber.UeIdEnd)
		if err != nil {
			logger.SimappLog.Errorln("error in Atoi with UeIdEnd", err)
			continue
		}
		if imsi >= start && imsi <= end {
			return true, subscriber
		}
	}
	return false, nil
}

// UpdateConfig diffs the config file f against the running one and dispatches the changes.
func UpdateConfig(f string) error {
	if content, err := os.ReadFile(f); err != nil {
		return err
	} else {
		NewSimappConfig := factory.Config{}

		if yamlErr := yaml.Unmarshal(content, &NewSimappConfig); yamlErr != nil {
			return yamlErr
		}
		if NewSimappConfig.Configuration == nil {
			logger.SimappLog.Infoln("configuration parsing failed", NewSimappConfig.Configuration)
			return nil
		}

		logger.SimappLog.Infoln("number of subscriber ranges in updated config", len(factory.SimappConfig.Configuration.Subscriber))
		var newImsiList []uint64
		for _, newSubscribers := range NewSimappConfig.Configuration.Subscriber {
			logger.SimappLog.Infoln("Subscribers:")
			logger.SimappLog.Infoln("UeIdStart", newSubscribers.UeIdStart)
			logger.SimappLog.Infoln("UeIdEnd", newSubscribers.UeIdEnd)
			logger.SimappLog.Infoln("PlmnId", newSubscribers.PlmnId)
			logger.SimappLog.Infoln("OPc", newSubscribers.OPc)
			logger.SimappLog.Infoln("OP", newSubscribers.OP)
			logger.SimappLog.Infoln("Key", newSubscribers.Key)
			logger.SimappLog.Infoln("SequenceNumber", newSubscribers.SequenceNumber)

			newStart, err := strconv.Atoi(newSubscribers.UeIdStart)
			if err != nil {
				logger.SimappLog.Errorln("error in Atoi with UeIdStart", err)
				continue
			}
			newEnd, err := strconv.Atoi(newSubscribers.UeIdEnd)
			if err != nil {
				logger.SimappLog.Errorln("error in Atoi with UeIdEnd", err)
				continue
			}
			for i := newStart; i <= newEnd; i++ {
				newImsiList = append(newImsiList, uint64(i))

				found, existingSubscriber := isImsiInSubscribers(i, factory.SimappConfig.Configuration.Subscriber)
				if found {
					if compareSubscriber(newSubscribers, existingSubscriber) {
						logger.SimappLog.Warnln("subscriber provision not support modify yet")
					}
					continue
				}
				// add subscriber to chan
				newSubscribers.UeId = fmt.Sprintf("%015d", i)

				b, err := json.Marshal(newSubscribers)
				if err != nil {
					logger.SimappLog.Errorln("error in marshal with newSubscriber", err)
					continue
				}
				reqMsgBody := bytes.NewBuffer(b)
				var msg configMessage
				msg.msgPtr = reqMsgBody
				msg.msgType = subscriber
				msg.name = newSubscribers.UeId
				msg.msgOp = add_op
				MsgChan <- msg
			}
		}
		// delete all the existing subscribers not show up in new config.
		for _, subscribers := range factory.SimappConfig.Configuration.Subscriber {
			start, err := strconv.Atoi(subscribers.UeIdStart)
			if err != nil {
				logger.SimappLog.Errorln("error in Atoi with UeIdStart", err)
				continue
			}
			end, err := strconv.Atoi(subscribers.UeIdEnd)
			if err != nil {
				logger.SimappLog.Errorln("error in Atoi with UeIdEnd", err)
				continue
			}
			for k := start; k <= end; k++ {
				if !slices.Contains(newImsiList, uint64(k)) {
					logger.SimappLog.Infoln("going to delete subscriber:", k)
					b, err := json.Marshal("")
					if err != nil {
						logger.SimappLog.Infoln("error in marshal with subscriber", err)
						continue
					}
					reqMsgBody := bytes.NewBuffer(b)
					var msg configMessage
					msg.msgPtr = reqMsgBody
					msg.msgType = subscriber
					msg.name = fmt.Sprintf("%015d", k)
					msg.msgOp = delete_op
					MsgChan <- msg
				}
			}
		}
		factory.SimappConfig.Configuration.Subscriber = NewSimappConfig.Configuration.Subscriber
		// end process subscriber update

		for _, group := range factory.SimappConfig.Configuration.DevGroup {
			group.Visited = false
		}
		for _, groupNew := range NewSimappConfig.Configuration.DevGroup {
			found := false
			for _, groupOld := range factory.SimappConfig.Configuration.DevGroup {
				if groupNew.Name == groupOld.Name {
					configChange := compareGroup(groupNew, groupOld)
					if configChange {
						// send Group Put
						logger.SimappLog.Infoln("updated group config", groupNew.Name)
						dispatchGroup(MsgChan, groupNew, modify_op, nil)
						// find all slices which are using this device group and mark them modified
						for _, slice := range factory.SimappConfig.Configuration.NetworkSlice {
							if slices.Contains(slice.DevGroups, groupOld.Name) {
								slice.Modified = true
							}
						}
					} else {
						logger.SimappLog.Infoln("config not updated for group", groupNew.Name)
					}
					found = true
					groupOld.Visited = true
					break
				}
			}
			if !found {
				// new Group - Send Post
				logger.SimappLog.Infoln("new group config", groupNew.Name)
				dispatchGroup(MsgChan, groupNew, add_op, nil)
			}
		}
		// visit all groups see if slice is deleted...if found = false
		for _, group := range factory.SimappConfig.Configuration.DevGroup {
			if !group.Visited {
				logger.SimappLog.Infoln("group deleted", group.Name)
				dispatchGroup(MsgChan, group, delete_op, nil)
				// find all slices which are using this device group and mark them modified
				for _, slice := range factory.SimappConfig.Configuration.NetworkSlice {
					if slices.Contains(slice.DevGroups, group.Name) {
						slice.Modified = true
					}
				}
			}
		}

		factory.SimappConfig.Configuration.DevGroup = NewSimappConfig.Configuration.DevGroup

		// visit all sliceOld see if slice is deleted...if found = false
		for _, slice := range factory.SimappConfig.Configuration.NetworkSlice {
			slice.Visited = false
		}

		for _, sliceNew := range NewSimappConfig.Configuration.NetworkSlice {
			found := false
			for _, sliceOld := range factory.SimappConfig.Configuration.NetworkSlice {
				if sliceNew.Name == sliceOld.Name {
					configChange := compareNetworkSlice(sliceNew, sliceOld)
					if sliceOld.Modified {
						logger.SimappLog.Infoln("updated slice config", sliceNew.Name)
						sliceOld.Modified = false
						dispatchNetworkSlice(MsgChan, sliceNew, modify_op, nil)
					} else if configChange {
						// send Slice Put
						logger.SimappLog.Infoln("updated slice config", sliceNew.Name)
						dispatchNetworkSlice(MsgChan, sliceNew, modify_op, nil)
					} else {
						logger.SimappLog.Infoln("config not updated for slice", sliceNew.Name)
					}
					found = true
					sliceOld.Visited = true
					break
				}
			}
			if !found {
				// new Slice - Send Post
				logger.SimappLog.Infoln("new slice config", sliceNew.Name)
				dispatchNetworkSlice(MsgChan, sliceNew, add_op, nil)
			}
		}
		// visit all sliceOld see if slice is deleted...if found = false
		for _, slice := range factory.SimappConfig.Configuration.NetworkSlice {
			if !slice.Visited {
				logger.SimappLog.Infoln("slice deleted", slice.Name)
				dispatchNetworkSlice(MsgChan, slice, delete_op, nil)
			}
		}
		factory.SimappConfig.Configuration.NetworkSlice = NewSimappConfig.Configuration.NetworkSlice
	}
	return nil
}

func WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.SimappLog.Infoln("config file changed:", e.Name)
		if err := UpdateConfig("config/simapp.yaml"); err != nil {
			logger.SimappLog.Errorln("error in loading updated configuration", err)
		} else {
			logger.SimappLog.Infoln("successfully updated configuration")
		}
	})
	logger.SimappLog.Infoln("watchConfig done")
}
