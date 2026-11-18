// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package provision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/omec-project/simapp/internal/factory"
	"github.com/omec-project/simapp/internal/logger"
)

// DispatchAllSubscribers queues every configured entry for provisioning.
func DispatchAllSubscribers(configMsgChan chan configMessage, wg *sync.WaitGroup) {
	logger.SimappLog.Infoln("number of subscriber ranges", len(factory.SimappConfig.Configuration.Subscriber))
	for _, subscribers := range factory.SimappConfig.Configuration.Subscriber {
		logger.SimappLog.Infof("subscribers: UeIdStart: %s, UeIdEnd: %s, PlmnId: %s, OPc: %s, OP: %s, Key: %s, SequenceNumber: %s",
			subscribers.UeIdStart, subscribers.UeIdEnd, subscribers.PlmnId, subscribers.OPc,
			subscribers.OP, subscribers.Key, subscribers.SequenceNumber)
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
		for i := start; i <= end; i++ {
			subscribers.UeId = fmt.Sprintf("%015d", i)
			logger.SimappLog.Debugln("ueId", subscribers.UeId)
			b, err := json.Marshal(subscribers)
			if err != nil {
				logger.SimappLog.Errorln("error in marshal with subscribers", err)
				continue
			}
			reqMsgBody := bytes.NewBuffer(b)
			var msg configMessage
			msg.msgPtr = reqMsgBody
			msg.msgType = subscriber
			msg.name = subscribers.UeId
			msg.msgOp = add_op
			if wg != nil {
				wg.Add(1)
				msg.wg = wg
			}
			configMsgChan <- msg
		}
	}
}

func splitNumericSuffix(value string) (string, uint64, int, error) {
	index := len(value)
	for index > 0 && value[index-1] >= '0' && value[index-1] <= '9' {
		index--
	}
	if index == len(value) {
		return "", 0, 0, fmt.Errorf("missing numeric suffix")
	}

	number, err := strconv.ParseUint(value[index:], 10, 64)
	if err != nil {
		return "", 0, 0, err
	}
	return value[:index], number, len(value) - index, nil
}

func dispatchGroup(configMsgChan chan configMessage, group *factory.DevGroup, msgOp int, wg *sync.WaitGroup) {
	groupCopy := *group
	group = &groupCopy
	if msgOp != delete_op && (group.ImsiStart != "" || group.ImsiEnd != "") {
		if group.ImsiStart == "" || group.ImsiEnd == "" {
			logger.SimappLog.Errorf("incomplete IMSI range %q-%q for group %s (both imsi-start and imsi-end are required)", group.ImsiStart, group.ImsiEnd, group.Name)
			return
		}
		start, startErr := strconv.ParseUint(group.ImsiStart, 10, 64)
		end, endErr := strconv.ParseUint(group.ImsiEnd, 10, 64)
		if startErr != nil || endErr != nil || start > end {
			logger.SimappLog.Errorf("invalid IMSI range %q-%q for group %s", group.ImsiStart, group.ImsiEnd, group.Name)
			return
		}
		diff := end - start
		if diff == ^uint64(0) {
			logger.SimappLog.Errorf("IMSI range too large %q-%q for group %s", group.ImsiStart, group.ImsiEnd, group.Name)
			return
		}
		count := diff + 1
		if count > uint64(^uint(0)>>1) {
			logger.SimappLog.Errorf("IMSI range too large %q-%q for group %s", group.ImsiStart, group.ImsiEnd, group.Name)
			return
		}
		group.Imsis = make([]string, 0, int(count))
		for imsi := start; ; imsi++ {
			group.Imsis = append(group.Imsis, fmt.Sprintf("%015d", imsi))
			if imsi == end {
				break
			}
		}
	}
	if msgOp != delete_op && (group.MsisdnStart != "" || group.MsisdnEnd != "") {
		if group.MsisdnStart == "" || group.MsisdnEnd == "" {
			logger.SimappLog.Errorf("incomplete MSISDN range %q-%q for group %s (both msisdn-start and msisdn-end are required)", group.MsisdnStart, group.MsisdnEnd, group.Name)
			return
		}
		prefix, start, width, startErr := splitNumericSuffix(group.MsisdnStart)
		endPrefix, end, endWidth, endErr := splitNumericSuffix(group.MsisdnEnd)
		if startErr != nil || endErr != nil || prefix != endPrefix || width != endWidth || start > end {
			logger.SimappLog.Errorf("invalid MSISDN range %q-%q for group %s", group.MsisdnStart, group.MsisdnEnd, group.Name)
			return
		}
		diff := end - start
		if diff == ^uint64(0) {
			logger.SimappLog.Errorf("MSISDN range too large %q-%q for group %s", group.MsisdnStart, group.MsisdnEnd, group.Name)
			return
		}
		count := diff + 1
		if count > uint64(^uint(0)>>1) {
			logger.SimappLog.Errorf("MSISDN range too large %q-%q for group %s", group.MsisdnStart, group.MsisdnEnd, group.Name)
			return
		}

		group.Msisdns = make([]string, 0, int(count))
		for msisdn := start; ; msisdn++ {
			group.Msisdns = append(group.Msisdns, fmt.Sprintf("%s%0*d", prefix, width, msisdn))
			if msisdn == end {
				break
			}
		}
	}
	logger.SimappLog.Infoln("group name", group.Name)
	logger.SimappLog.Infoln("site name", group.SiteInfo)
	logger.SimappLog.Infoln("imsis", group.Imsis)
	for _, imsi := range group.Imsis {
		logger.SimappLog.Debugln("imsi", imsi)
	}
	logger.SimappLog.Infoln("IpDomainName", group.IpDomainName)
	// Check if IpDomains is nil or not
	if group.IpDomains != nil {
		for _, ipDomain := range group.IpDomains {
			logger.SimappLog.Infoln("  IpDomain Dnn:", ipDomain.Dnn)
			logger.SimappLog.Infoln("  IpDomain Dns Primary:", ipDomain.DnsPrimary)
			logger.SimappLog.Infoln("  IpDomain Pcscf Primary:", ipDomain.PcscfPrimary)
			logger.SimappLog.Infoln("  IpDomain Mtu:", ipDomain.Mtu)
			logger.SimappLog.Infoln("  IpDomain UePool:", ipDomain.UePool)

			// Check if UeDnnQos field is populated
			if ipDomain.UeDnnQos != nil {
				logger.SimappLog.Infoln("  UeDnnQos:", ipDomain.UeDnnQos)
			} else {
				logger.SimappLog.Warnln("  UeDnnQos is nil")
			}
		}
	} else {
		logger.SimappLog.Warnln("  IpDomains is nil")
	}
	b, err := json.Marshal(group)
	if err != nil {
		logger.SimappLog.Errorln("error in marshal", err)
		return
	}
	reqMsgBody := bytes.NewBuffer(b)
	var msg configMessage
	msg.msgPtr = reqMsgBody
	msg.msgType = device_group
	msg.name = group.Name
	msg.msgOp = msgOp
	if wg != nil {
		wg.Add(1)
		msg.wg = wg
	}
	configMsgChan <- msg
}

// DispatchAllGroups queues every configured entry for provisioning.
func DispatchAllGroups(configMsgChan chan configMessage, wg *sync.WaitGroup) {
	logger.SimappLog.Infoln("number of device groups", len(factory.SimappConfig.Configuration.DevGroup))
	for _, group := range factory.SimappConfig.Configuration.DevGroup {
		dispatchGroup(configMsgChan, group, add_op, wg)
	}
}

func dispatchNetworkSlice(configMsgChan chan configMessage, slice *factory.NetworkSlice, msgOp int, wg *sync.WaitGroup) {
	logger.SimappLog.Infoln("slice name:", slice.Name)
	logger.SimappLog.Infof("slice sst %v, sd %v", slice.SliceId.Sst, slice.SliceId.Sd)
	logger.SimappLog.Infof("slice site info: %+v", slice.SiteInfo)
	site := slice.SiteInfo
	logger.SimappLog.Infoln("slice site name", site.SiteName)
	logger.SimappLog.Infoln("slice gNB", len(site.Gnb))
	for i, gnb := range site.Gnb {
		logger.SimappLog.Infof("slice gNB[%v] = %s, tac: %d", i, gnb.Name, gnb.Tac)
	}
	logger.SimappLog.Infof("slice Plmn %+v", site.Plmn)
	logger.SimappLog.Infof("slice Upf %+v", site.Upf)

	logger.SimappLog.Infoln("slice device groups", slice.DevGroups)
	for _, devGroup := range slice.DevGroups {
		logger.SimappLog.Infoln("attached device group", devGroup)
	}

	b, err := json.Marshal(slice)
	if err != nil {
		logger.SimappLog.Errorln("error in marshal", err)
		return
	}
	reqMsgBody := bytes.NewBuffer(b)

	var msg configMessage
	msg.msgPtr = reqMsgBody
	msg.msgType = network_slice
	msg.name = slice.Name
	msg.msgOp = msgOp
	if wg != nil {
		wg.Add(1)
		msg.wg = wg
	}
	configMsgChan <- msg
}

// DispatchAllNetworkSlices queues every configured entry for provisioning.
func DispatchAllNetworkSlices(configMsgChan chan configMessage, wg *sync.WaitGroup) {
	logger.SimappLog.Infoln("number of network slices", len(factory.SimappConfig.Configuration.NetworkSlice))
	for _, slice := range factory.SimappConfig.Configuration.NetworkSlice {
		dispatchNetworkSlice(configMsgChan, slice, add_op, wg)
	}
}
