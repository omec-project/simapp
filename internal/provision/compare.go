// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package provision

import (
	"crypto/sha1"
	"encoding/hex"
	"reflect"
	"slices"

	"github.com/omec-project/simapp/internal/factory"
	"github.com/omec-project/simapp/internal/logger"
)

func compareSubscriber(subscriberNew *factory.Subscriber, subscriberOld *factory.Subscriber) bool {
	if subscriberNew.PlmnId != subscriberOld.PlmnId {
		logger.SimappLog.Infoln("plmn ID changed")
		return true
	}
	if subscriberNew.OPc != subscriberOld.OPc {
		logger.SimappLog.Infoln("OPc changed")
		return true
	}
	if subscriberNew.OP != subscriberOld.OP {
		logger.SimappLog.Infoln("OP changed")
		return true
	}
	if subscriberNew.Key != subscriberOld.Key {
		logger.SimappLog.Infoln("Key changed")
		return true
	}
	if subscriberNew.SequenceNumber != subscriberOld.SequenceNumber {
		logger.SimappLog.Infoln("SequenceNumber changed")
		return true
	}
	return false
}

func compareGroup(groupNew *factory.DevGroup, groupOld *factory.DevGroup) bool {
	if groupNew.IpDomainName != groupOld.IpDomainName {
		logger.SimappLog.Infoln("ip domain name changed")
		return true
	}

	if groupNew.SiteInfo != groupOld.SiteInfo {
		logger.SimappLog.Infoln("siteInfo name changed")
		return true
	}

	if groupNew.ImsiStart != groupOld.ImsiStart || groupNew.ImsiEnd != groupOld.ImsiEnd {
		logger.SimappLog.Infoln("imsi range changed")
		return true
	}

	if len(groupNew.Imsis) != len(groupOld.Imsis) {
		logger.SimappLog.Infoln("number of Imsis changed")
		return true
	}

	if groupNew.MsisdnStart != groupOld.MsisdnStart || groupNew.MsisdnEnd != groupOld.MsisdnEnd {
		logger.SimappLog.Infoln("msisdn range changed")
		return true
	}

	// Compare MSISDN lists
	if !reflect.DeepEqual(groupNew.Msisdns, groupOld.Msisdns) {
		logger.SimappLog.Infoln("msisdns list has changed")
		return true
	}

	var allimsiNew string
	for _, imsi := range groupNew.Imsis {
		allimsiNew = allimsiNew + imsi
	}
	h1 := sha1.New()
	h1.Write([]byte(allimsiNew))
	bs1 := h1.Sum(nil)
	strcode1 := hex.EncodeToString(bs1[:])

	var allimsiOld string
	for _, imsi := range groupOld.Imsis {
		allimsiOld = allimsiOld + imsi
	}
	h2 := sha1.New()
	h2.Write([]byte(allimsiOld))
	bs2 := h2.Sum(nil)
	strcode2 := hex.EncodeToString(bs2[:])

	logger.SimappLog.Infof("CODE1: %s and CODE2: %s", strcode1, strcode2)
	if strcode2 != strcode1 {
		return true
	}

	if len(groupNew.IpDomains) != len(groupOld.IpDomains) {
		logger.SimappLog.Infoln("Number of IpDomains changed")
		return true
	}

	for i := range groupNew.IpDomains {
		if i >= len(groupOld.IpDomains) {
			// If groupOld doesn't have enough IpDomains
			return true
		}

		oldIpDomain := groupOld.IpDomains[i]
		newIpDomain := groupNew.IpDomains[i]

		if oldIpDomain.Dnn != newIpDomain.Dnn {
			logger.SimappLog.Infoln("DNN changed")
			return true
		}

		if oldIpDomain.Mtu != newIpDomain.Mtu {
			logger.SimappLog.Infoln("MTU changed")
			return true
		}

		if oldIpDomain.PcscfPrimary != newIpDomain.PcscfPrimary {
			logger.SimappLog.Infoln("PcscfPrimary changed")
			return true
		}

		if oldIpDomain.UePool != newIpDomain.UePool {
			logger.SimappLog.Infoln("UePool changed")
			return true
		}

		// Detect addition or removal of UeDnnQos
		if (oldIpDomain.UeDnnQos == nil) != (newIpDomain.UeDnnQos == nil) {
			logger.SimappLog.Infoln("UeDnnQos presence changed")
			return true
		}

		if oldIpDomain.UeDnnQos != nil && newIpDomain.UeDnnQos != nil {
			// Detect addition or removal of TrafficClass
			if (oldIpDomain.UeDnnQos.TrafficClass == nil) !=
				(newIpDomain.UeDnnQos.TrafficClass == nil) {
				logger.SimappLog.Infoln("TrafficClass presence changed")
				return true
			}

			// Compare UeDnnQos bitrate fields
			if oldIpDomain.UeDnnQos.Uplink != newIpDomain.UeDnnQos.Uplink ||
				oldIpDomain.UeDnnQos.Downlink != newIpDomain.UeDnnQos.Downlink ||
				oldIpDomain.UeDnnQos.BitRateUnit != newIpDomain.UeDnnQos.BitRateUnit {
				logger.SimappLog.Infoln("UeDnnQos bitrate parameters changed")
				return true
			}

			// Compare TrafficClass fields when both exist
			if oldIpDomain.UeDnnQos.TrafficClass != nil &&
				newIpDomain.UeDnnQos.TrafficClass != nil {
				if oldIpDomain.UeDnnQos.TrafficClass.Name != newIpDomain.UeDnnQos.TrafficClass.Name ||
					oldIpDomain.UeDnnQos.TrafficClass.Qci != newIpDomain.UeDnnQos.TrafficClass.Qci ||
					oldIpDomain.UeDnnQos.TrafficClass.Arp != newIpDomain.UeDnnQos.TrafficClass.Arp ||
					oldIpDomain.UeDnnQos.TrafficClass.Pdb != newIpDomain.UeDnnQos.TrafficClass.Pdb ||
					oldIpDomain.UeDnnQos.TrafficClass.Pelr != newIpDomain.UeDnnQos.TrafficClass.Pelr {
					logger.SimappLog.Infoln("TrafficClass properties changed")
					return true
				}
			}
		}
	}
	return false
}

func compareNetworkSlice(sliceNew *factory.NetworkSlice, sliceOld *factory.NetworkSlice) bool {
	// slice Id should not change
	appFilteringRulesNew := sliceNew.ApplicationFilteringRules
	appFilteringRulesOld := sliceOld.ApplicationFilteringRules
	if !reflect.DeepEqual(appFilteringRulesNew, appFilteringRulesOld) {
		return true
	}
	for _, ng := range sliceNew.DevGroups {
		if !slices.Contains(sliceOld.DevGroups, ng) {
			logger.SimappLog.Infoln("new dev group added in slice")
			return true // 2 network slices have some difference
		}
	}
	for _, ng := range sliceOld.DevGroups {
		if !slices.Contains(sliceNew.DevGroups, ng) {
			logger.SimappLog.Infoln("dev group deleted in slice")
			return true // 2 network slices have some difference
		}
	}
	oldSite := sliceOld.SiteInfo
	newSite := sliceNew.SiteInfo
	if oldSite.SiteName != newSite.SiteName {
		logger.SimappLog.Infoln("site name changed")
		return true
	}
	oldUpf := oldSite.Upf
	newUpf := newSite.Upf
	if (oldUpf.UpfName != newUpf.UpfName) && (oldUpf.UpfPort != newUpf.UpfPort) {
		logger.SimappLog.Infoln("upf details changed")
		return true
	}

	for _, newgnb := range newSite.Gnb {
		if !slices.ContainsFunc(oldSite.Gnb, func(oldgnb *factory.Gnb) bool {
			return newgnb.Name == oldgnb.Name && newgnb.Tac == oldgnb.Tac
		}) {
			logger.SimappLog.Infoln("gnb changed in slice")
			return true // change in slice details
		}
	}

	logger.SimappLog.Warnln("no change in slices")
	return false
}
