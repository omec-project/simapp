// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package factory

type Config struct {
	Info          *Info          `yaml:"info"`
	Configuration *Configuration `yaml:"configuration"`
	Logger        *Logger        `yaml:"logger"`
}

type Logger struct {
	APP *LogSetting `yaml:"APP" valid:"optional"`
}

type LogSetting struct {
	DebugLevel string `yaml:"debugLevel" valid:"debugLevel"`
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	HttpVersion int    `yaml:"http-version,omitempty"`
}

type Configuration struct {
	SubProvisionEndpt *SubProvisionEndpt `yaml:"sub-provision-endpt,omitempty"`
	DevGroup          []*DevGroup        `yaml:"device-groups,omitempty"`
	NetworkSlice      []*NetworkSlice    `yaml:"network-slices,omitempty"`
	Subscriber        []*Subscriber      `yaml:"subscribers,omitempty"`
	MaxWorkers        int                `yaml:"max-workers,omitempty"`
}

type DevGroup struct {
	Name         string     `yaml:"name,omitempty"`
	SiteInfo     string     `yaml:"site-info,omitempty" json:"site-info,omitempty"`
	Imsis        []string   `yaml:"imsis,omitempty" json:"imsis,omitempty"`
	ImsiStart    string     `yaml:"imsi-start,omitempty" json:"imsi-start,omitempty"`
	ImsiEnd      string     `yaml:"imsi-end,omitempty" json:"imsi-end,omitempty"`
	Msisdns      []string   `yaml:"msisdns,omitempty" json:"msisdns,omitempty"`
	MsisdnStart  string     `yaml:"msisdn-start,omitempty" json:"msisdn-start,omitempty"`
	MsisdnEnd    string     `yaml:"msisdn-end,omitempty" json:"msisdn-end,omitempty"`
	IpDomainName string     `yaml:"ip-domain-name,omitempty" json:"ip-domain-name,omitempty"`
	IpDomains    []IpDomain `yaml:"ip-domains,omitempty" json:"ip-domains,omitempty"`
	Visited      bool       `yaml:"-" json:"-"`
}

type IpDomain struct {
	UeDnnQos     *UeDnnQosInfo `yaml:"ue-dnn-qos,omitempty" json:"ue-dnn-qos,omitempty"`
	Dnn          string        `yaml:"dnn,omitempty" json:"dnn,omitempty"`
	DnsPrimary   string        `yaml:"dns-primary,omitempty" json:"dns-primary,omitempty"`
	PcscfPrimary string        `yaml:"pcscf-primary,omitempty" json:"pcscf-primary,omitempty"`
	DnsSecondary string        `yaml:"dns-secondary,omitempty" json:"dns-secondary,omitempty"`
	UePool       string        `yaml:"ue-ip-pool,omitempty" json:"ue-ip-pool,omitempty"`
	Mtu          int           `yaml:"mtu,omitempty" json:"mtu,omitempty"`
}

type Subscriber struct {
	UeId           string
	UeIdStart      string `yaml:"ueId-start,omitempty" json:"-"`
	UeIdEnd        string `yaml:"ueId-end,omitempty" json:"-"`
	PlmnId         string `yaml:"plmnId,omitempty" json:"plmnId,omitempty"`
	OPc            string `yaml:"opc,omitempty" json:"opc,omitempty"`
	OP             string `yaml:"op,omitempty" json:"op,omitempty"`
	Key            string `yaml:"key,omitempty" json:"key,omitempty"`
	SequenceNumber string `yaml:"sequenceNumber,omitempty" json:"sequenceNumber,omitempty"`
}

type SubProvisionEndpt struct {
	Addr string `yaml:"addr,omitempty" json:"addr,omitempty"`
	Port string `yaml:"port,omitempty" json:"port,omitempty"`
}

type NetworkSlice struct {
	Name                      string                       `yaml:"name,omitempty" json:"name,omitempty"`
	SliceId                   *SliceId                     `yaml:"slice-id,omitempty" json:"slice-id,omitempty"`
	DevGroups                 []string                     `yaml:"site-device-group,omitempty" json:"site-device-group,omitempty"`
	SiteInfo                  *SiteInfo                    `yaml:"site-info,omitempty" json:"site-info,omitempty"`
	ApplicationFilteringRules []*ApplicationFilteringRules `yaml:"application-filtering-rules,omitempty" json:"application-filtering-rules,omitempty"`
	Visited                   bool                         `yaml:"-" json:"-"`
	Modified                  bool                         `yaml:"-" json:"-"`
}

type SliceId struct {
	Sst string `yaml:"sst,omitempty" json:"sst,omitempty"`
	Sd  string `yaml:"sd,omitempty" json:"sd,omitempty"`
}

type UeDnnQosInfo struct {
	TrafficClass *TrafficClassInfo `yaml:"traffic-class,omitempty" json:"traffic-class,omitempty"`
	BitRateUnit  string            `yaml:"bitrate-unit,omitempty" json:"bitrate-unit,omitempty"`
	Uplink       int               `yaml:"dnn-mbr-uplink,omitempty" json:"dnn-mbr-uplink,omitempty"`
	Downlink     int               `yaml:"dnn-mbr-downlink,omitempty" json:"dnn-mbr-downlink,omitempty"`
}

type TrafficClassInfo struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Qci  int    `yaml:"qci,omitempty" json:"qci,omitempty"`
	Arp  int    `yaml:"arp,omitempty" json:"arp,omitempty"`
	Pdb  int    `yaml:"pdb,omitempty" json:"pdb,omitempty"`
	Pelr int    `yaml:"pelr,omitempty" json:"pelr,omitempty"`
}

type SiteInfo struct {
	Plmn     *Plmn  `yaml:"plmn,omitempty"   json:"plmn,omitempty"`
	Upf      *Upf   `yaml:"upf,omitempty" json:"upf,omitempty"`
	SiteName string `yaml:"site-name,omitempty" json:"site-name,omitempty"`
	Gnb      []*Gnb `yaml:"gNodeBs,omitempty" json:"gNodeBs,omitempty"`
}

type Gnb struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Tac  int    `yaml:"tac,omitempty" json:"tac,omitempty"`
}

type Plmn struct {
	Mcc string `yaml:"mcc,omitempty" json:"mcc,omitempty"`
	Mnc string `yaml:"mnc,omitempty" json:"mnc,omitempty"`
}

type Upf struct {
	UpfName string `yaml:"upf-name,omitempty" json:"upf-name,omitempty"`
	UpfPort int    `yaml:"upf-port,omitempty" json:"upf-port,omitempty"`
}

type ApplicationFilteringRules struct {
	TrafficClass   *TrafficClassInfo `yaml:"traffic-class,omitempty" json:"traffic-class,omitempty"`
	RuleName       string            `yaml:"rule-name,omitempty" json:"rule-name,omitempty"`
	Action         string            `yaml:"action,omitempty" json:"action,omitempty"`
	Endpoint       string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	BitRateUnit    string            `yaml:"bitrate-unit,omitempty" json:"bitrate-unit,omitempty"`
	RuleTrigger    string            `yaml:"rule-trigger,omitempty" json:"rule-trigger,omitempty"`
	Priority       int32             `yaml:"priority,omitempty" json:"priority,omitempty"`
	Protocol       int32             `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	StartPort      int32             `yaml:"dest-port-start,omitempty" json:"dest-port-start,omitempty"`
	EndPort        int32             `yaml:"dest-port-end,omitempty" json:"dest-port-end,omitempty"`
	AppMbrUplink   int32             `yaml:"app-mbr-uplink,omitempty" json:"app-mbr-uplink,omitempty"`
	AppMbrDownlink int32             `yaml:"app-mbr-downlink,omitempty" json:"app-mbr-downlink,omitempty"`
}
