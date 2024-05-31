// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/omec-project/util/logger"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Info          *Info          `yaml:"info"`
	Configuration *Configuration `yaml:"configuration"`
	Logger        *logger.Logger `yaml:"logger"`
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	HttpVersion int    `yaml:"http-version,omitempty"`
}

type Configuration struct {
	ConfigSlice       bool               `yaml:"provision-network-slice,omitempty"`
	DevGroup          []*DevGroup        `yaml:"device-groups,omitempty"`
	NetworkSlice      []*NetworkSlice    `yaml:"network-slices,omitempty"`
	Subscriber        []*Subscriber      `yaml:"subscribers,omitempty"`
	SubProvisionEndpt *SubProvisionEndpt `yaml:"sub-provision-endpt,omitempty"`
	SubProxyEndpt     *SubProxyEndpt     `yaml:"sub-proxy-endpt,omitempty"`
}

type DevGroup struct {
	Name         string    `yaml:"name,omitempty"`
	SiteInfo     string    `yaml:"site-info,omitempty" json:"site-info,omitempty"`
	Imsis        []string  `yaml:"imsis,omitempty" json:"imsis,omitempty"`
	IpDomainName string    `yaml:"ip-domain-name,omitempty" json:"ip-domain-name,omitempty"`
	IpDomain     *IpDomain `yaml:"ip-domain-expanded,omitempty" json:"ip-domain-expanded,omitempty"`
	visited      bool
}

type IpDomain struct {
	Dnn          string        `yaml:"dnn,omitempty" json:"dnn,omitempty"`
	DnsPrimary   string        `yaml:"dns-primary,omitempty" json:"dns-primary,omitempty"`
	DnsSecondary string        `yaml:"dns-secondary,omitempty" json:"dns-secondary,omitempty"`
	Mtu          int           `yaml:"mtu,omitempty" json:"mtu,omitempty"`
	UePool       string        `yaml:"ue-ip-pool,omitempty" json:"ue-ip-pool,omitempty"`
	UeDnnQos     *UeDnnQosInfo `yaml:"ue-dnn-qos,omitempty" json:"ue-dnn-qos,omitempty"`
}

type Subscriber struct {
	UeId           string
	UeIdStart      string `yaml:"ueId-start,omitempty" json:"-",omitempty`
	UeIdEnd        string `yaml:"ueId-end,omitempty" json:"-", omitempty`
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

type SubProxyEndpt struct {
	Addr string `yaml:"addr,omitempty" json:"addr,omitempty"`
	Port string `yaml:"port,omitempty" json:"port,omitempty"`
}

type NetworkSlice struct {
	Name                      string                       `yaml:"name,omitempty" json:"name,omitempty"`
	SliceId                   *SliceId                     `yaml:"slice-id,omitempty" json:"slice-id,omitempty"`
	DevGroups                 []string                     `yaml:"site-device-group,omitempty" json:"site-device-group,omitempty"`
	SiteInfo                  *SiteInfo                    `yaml:"site-info,omitempty" json:"site-info,omitempty"`
	ApplicationFilteringRules []*ApplicationFilteringRules `yaml:"application-filtering-rules,omitempty" json:"application-filtering-rules,omitempty"`
	visited                   bool
	modified                  bool
}

type SliceId struct {
	Sst string `yaml:"sst,omitempty" json:"sst,omitempty"`
	Sd  string `yaml:"sd,omitempty" json:"sd,omitempty"`
}

type UeDnnQosInfo struct {
	Uplink       int               `yaml:"dnn-mbr-uplink,omitempty" json:"dnn-mbr-uplink,omitempty"`
	Downlink     int               `yaml:"dnn-mbr-downlink,omitempty" json:"dnn-mbr-downlink,omitempty"`
	BitRateUnit  string            `yaml:"bitrate-unit,omitempty" json:"bitrate-unit,omitempty"`
	TrafficClass *TrafficClassInfo `yaml:"traffic-class,omitempty" json:"traffic-class,omitempty"`
}

type TrafficClassInfo struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Qci  int    `yaml:"qci,omitempty" json:"qci,omitempty"`
	Arp  int    `yaml:"arp,omitempty" json:"arp,omitempty"`
	Pdb  int    `yaml:"pdb,omitempty" json:"pdb,omitempty"`
	Pelr int    `yaml:"pelr,omitempty" json:"pelr,omitempty"`
}

type SiteInfo struct {
	SiteName string `yaml:"site-name,omitempty" json:"site-name,omitempty"`
	Gnb      []*Gnb `yaml:"gNodeBs,omitempty" json:"gNodeBs,omitempty"`
	Plmn     *Plmn  `yaml:"plmn,omitempty"   json:"plmn,omitempty"`
	Upf      *Upf   `yaml:"upf,omitempty" json:"upf,omitempty"`
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
	// Rule name
	RuleName string `yaml:"rule-name,omitempty" json:"rule-name,omitempty"`
	// priority
	Priority int32 `yaml:"priority,omitempty" json:"priority,omitempty"`
	// action
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
	// Application Desination IP or network
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	// protocol
	Protocol int32 `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	// port range start
	StartPort int32 `yaml:"dest-port-start,omitempty" json:"dest-port-start,omitempty"`
	// port range end
	EndPort int32 `yaml:"dest-port-end,omitempty" json:"dest-port-end,omitempty"`

	AppMbrUplink int32 `yaml:"app-mbr-uplink,omitempty" json:"app-mbr-uplink,omitempty"`

	AppMbrDownlink int32 `yaml:"app-mbr-downlink,omitempty" json:"app-mbr-downlink,omitempty"`

	BitRateUnit string `yaml:"bitrate-unit,omitempty" json:"bitrate-unit,omitempty"`

	TrafficClass *TrafficClassInfo `yaml:"traffic-class,omitempty" json:"traffic-class,omitempty"`

	RuleTrigger string `yaml:"rule-trigger,omitempty" json:"rule-trigger,omitempty"`
}

const (
	add_op = iota
	modify_op
	delete_op
)

const (
	device_group = iota
	network_slice
	subscriber
)

const httpProtocol = "http://"

type configMessage struct {
	msgPtr  *bytes.Buffer
	msgType int
	name    string
	msgOp   int
}

func (msg configMessage) String() string {
	var msgType, msgOp string
	switch msg.msgOp {
	case add_op:
		msgOp = "add_op"
	case modify_op:
		msgOp = "modify_op"
	case delete_op:
		msgOp = "delete_op"
	}

	switch msg.msgType {
	case device_group:
		msgType = "device-group"
	case network_slice:
		msgType = "network-slice"
	case subscriber:
		msgType = "subscriber"
	}
	return fmt.Sprintf("Config msg name [%v], type [%v], op [%v]", msg.name, msgType, msgOp)
}

var (
	SimappConfig  Config
	configMsgChan chan configMessage
	client        *http.Client
)

func InitConfigFactory(f string, configMsgChan chan configMessage, subProvisionEndpt *SubProvisionEndpt, subProxyEndpt *SubProxyEndpt) error {
	log.Println("Function called ", f)
	if content, err := os.ReadFile(f); err != nil {
		log.Println("Readfile failed called ", err)
		return err
	} else {
		SimappConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &SimappConfig); yamlErr != nil {
			log.Println("yaml parsing failed ", yamlErr)
			return yamlErr
		}
	}
	if SimappConfig.Configuration == nil {
		log.Println("Configuration Parsing Failed ", SimappConfig.Configuration)
		return nil
	}

	// set http client
	if SimappConfig.Info.HttpVersion == 2 {
		client = &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
			Timeout: 5 * time.Second,
		}
	} else {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	log.Println("Subscriber Provision Endpoint:")
	log.Println("Address ", SimappConfig.Configuration.SubProvisionEndpt.Addr)
	log.Println("Port ", SimappConfig.Configuration.SubProvisionEndpt.Port)
	subProvisionEndpt.Addr = SimappConfig.Configuration.SubProvisionEndpt.Addr
	subProvisionEndpt.Port = SimappConfig.Configuration.SubProvisionEndpt.Port

	if SimappConfig.Configuration.SubProxyEndpt != nil && SimappConfig.Configuration.SubProxyEndpt.Addr != "" {
		log.Println("Subscriber Proxy Endpoint:")
		log.Println("Address ", SimappConfig.Configuration.SubProxyEndpt.Addr)
		log.Println("Port ", SimappConfig.Configuration.SubProxyEndpt.Port)
		subProxyEndpt.Addr = SimappConfig.Configuration.SubProxyEndpt.Addr
		subProxyEndpt.Port = SimappConfig.Configuration.SubProxyEndpt.Port
	}

	viper.SetConfigName("simapp.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/simapp/config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return err
	}
	return nil
}

func syncConfig(w http.ResponseWriter, req *http.Request) {
	_, err := fmt.Fprintf(w, "OK\n")
	if err != nil {
		log.Println(err)
	}
	dispatchAllGroups(configMsgChan)
	dispatchAllNetworkSlices(configMsgChan)
}

func main() {
	log.Println("SimApp started")
	configMsgChan = make(chan configMessage, 100)
	var subProvisionEndpt SubProvisionEndpt
	var subProxyEndpt SubProxyEndpt

	err := InitConfigFactory("./config/simapp.yaml", configMsgChan, &subProvisionEndpt, &subProxyEndpt)
	if err != nil {
		log.Println(err)
	}

	go sendMessage(configMsgChan, subProvisionEndpt, subProxyEndpt)
	go WatchConfig()

	dispatchAllSubscribers(configMsgChan)
	dispatchAllGroups(configMsgChan)
	dispatchAllNetworkSlices(configMsgChan)

	http.HandleFunc("/synchronize", syncConfig)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		// Note: as per the `ListenAndServe` documentation: "ListenAndServe always returns a non-nil error."
		log.Println(err)
	}
	for {
		time.Sleep(100 * time.Second)
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
		log.Println(err)
	}
	for {
		cloneReq := req.Clone(context.Background())
		req.Body = io.NopCloser(bytes.NewReader(body))
		cloneReq.Body = io.NopCloser(bytes.NewReader(body))
		rsp, err := client.Do(cloneReq)
		retries += 1
		if err != nil {
			nextInterval := getNextBackoffInterval(retries, 2)
			log.Printf("http req send error [%v], retrying after %v sec...", err.Error(), nextInterval)
			time.Sleep(time.Second * time.Duration(nextInterval))
			continue
		}

		if rsp.StatusCode == http.StatusAccepted ||
			rsp.StatusCode == http.StatusOK || rsp.StatusCode == http.StatusNoContent ||
			rsp.StatusCode == http.StatusCreated {
			log.Println("config push success")
			err = req.Body.Close()
			if err != nil {
				log.Println(err)
			}
			return rsp, nil
		} else {
			nextInterval := getNextBackoffInterval(retries, 2)
			log.Printf("http rsp error [%v], retrying after [%v] sec...", http.StatusText(rsp.StatusCode), nextInterval)
			err = rsp.Body.Close()
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second * time.Duration(nextInterval))
		}
	}
}

func sendMessage(msgChan chan configMessage, subProvisionEndpt SubProvisionEndpt, subProxyEndpt SubProxyEndpt) {
	var devGroupHttpend string
	var networkSliceHttpend string
	var subscriberHttpend string

	log.Println("Subscriber Provision Endpoint in sendMessage:")
	log.Println("Address ", subProvisionEndpt.Addr)
	log.Println("Port ", subProvisionEndpt.Port)

	ip := strings.TrimSpace(subProvisionEndpt.Addr)

	log.Println("webui running at ", ip)
	devGroupHttpend = httpProtocol + ip + ":" + subProvisionEndpt.Port + "/config/v1/device-group/"
	log.Println("device trigger  http endpoint ", devGroupHttpend)
	networkSliceHttpend = httpProtocol + ip + ":" + subProvisionEndpt.Port + "/config/v1/network-slice/"
	log.Println("network slice http endpoint ", devGroupHttpend)
	subscriberHttpend = httpProtocol + ip + ":" + subProvisionEndpt.Port + "/api/subscriber/imsi-"
	log.Println("subscriber http endpoint ", subscriberHttpend)
	baseDestUrl := subscriberHttpend
	if subProxyEndpt.Port != "" {
		ip := strings.TrimSpace(subProxyEndpt.Addr)
		devGroupHttpend = httpProtocol + ip + ":" + subProxyEndpt.Port + "/config/v1/device-group/"
		log.Println("device trigger Proxy http endpoint ", devGroupHttpend)
		networkSliceHttpend = httpProtocol + ip + ":" + subProxyEndpt.Port + "/config/v1/network-slice/"
		log.Println("network slice Proxy http endpoint ", devGroupHttpend)
		subscriberHttpend = httpProtocol + ip + ":" + subProxyEndpt.Port + "/api/subscriber/imsi-"
		log.Println("subscriber Proxy http endpoint ", subscriberHttpend)
	}

	for msg := range msgChan {
		var httpend string
		var destUrl string
		log.Println("Received Message from Channel", msgChan, msg)
		switch msg.msgType {
		case device_group:
			httpend = devGroupHttpend + msg.name
		case network_slice:
			httpend = networkSliceHttpend + msg.name
		case subscriber:
			httpend = subscriberHttpend + msg.name
			destUrl = baseDestUrl + msg.name
		}
		var rsp *http.Response
		var httpErr error
		for {
			if msg.msgOp == add_op {
				log.Printf("Post Message [%v] to %v", msg.String(), httpend)
				req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, httpend, msg.msgPtr)
				if err != nil {
					fmt.Printf("An Error Occurred %v", err)
					time.Sleep(1 * time.Second)
					continue
				}

				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				if subProxyEndpt.Port != "" {
					req.Header.Add("Dest-Url", destUrl)
				}
				rsp, httpErr = sendHttpReqMsg(req)
				if httpErr != nil {
					log.Printf("Post Message [%v] returned error [%v] ", httpend, httpErr.Error())
				}

				fmt.Printf("Message POST %v Success\n", rsp.StatusCode)
			} else if msg.msgOp == modify_op {
				log.Printf("Put Message [%v] to %v", msg.String(), httpend)

				req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, httpend, msg.msgPtr)
				// Handle Error
				if err != nil {
					fmt.Printf("An Error Occurred %v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				// set the request header Content-Type for json
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				if subProxyEndpt.Port != "" {
					req.Header.Add("Dest-Url", destUrl)
				}
				rsp, httpErr = sendHttpReqMsg(req)
				if httpErr != nil {
					log.Printf("Put Message [%v] returned error [%v] ", httpend, httpErr.Error())
				}

				fmt.Printf("Message PUT %v Success\n", rsp.StatusCode)
			} else if msg.msgOp == delete_op {
				log.Printf("Delete Message [%v] to %v", msg.String(), httpend)

				req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, httpend, msg.msgPtr)
				// Handle Error
				if err != nil {
					fmt.Printf("An Error Occurred %v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				// set the request header Content-Type for json
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				if subProxyEndpt.Port != "" {
					req.Header.Add("Dest-Url", destUrl)
				}
				rsp, httpErr = sendHttpReqMsg(req)
				if httpErr != nil {
					log.Printf("Delete Message [%v] returned error [%v] ", httpend, httpErr.Error())
				}
				fmt.Printf("Message DEL %v Success\n", rsp.StatusCode)
			}
			err := rsp.Body.Close()
			if err != nil {
				log.Println(err)
			}
			break
		}
	}
}

func compareSubscriber(subscriberNew *Subscriber, subscriberOld *Subscriber) bool {
	if subscriberNew.PlmnId != subscriberOld.PlmnId {
		log.Println("Plmn ID changed.")
		return true
	}
	if subscriberNew.OPc != subscriberOld.OPc {
		log.Println("OPc changed.")
		return true
	}
	if subscriberNew.OP != subscriberOld.OP {
		log.Println("OP changed.")
		return true
	}
	if subscriberNew.Key != subscriberOld.Key {
		log.Println("Key changed.")
		return true
	}
	if subscriberNew.SequenceNumber != subscriberOld.SequenceNumber {
		log.Println("SequenceNumber changed.")
		return true
	}
	return false
}

func compareGroup(groupNew *DevGroup, groupOld *DevGroup) bool {
	if groupNew.IpDomainName != groupOld.IpDomainName {
		log.Println("IP domain name changed.")
		return true
	}

	if groupNew.SiteInfo != groupOld.SiteInfo {
		log.Println("SIteInfo name changed.")
		return true
	}

	if len(groupNew.Imsis) != len(groupOld.Imsis) {
		log.Println("number of Imsis changed.")
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

	log.Println("CODE1 and CODE2 ", strcode1, strcode2)
	if strcode2 != strcode1 {
		return true
	}

	oldipdomain := groupOld.IpDomain
	newipdomain := groupNew.IpDomain
	if oldipdomain.Dnn != newipdomain.Dnn {
		return true
	}
	if oldipdomain.Mtu != newipdomain.Mtu {
		return true
	}
	if oldipdomain.UePool != newipdomain.UePool {
		return true
	}
	if oldipdomain.UeDnnQos != nil && newipdomain.UeDnnQos != nil {
		if oldipdomain.UeDnnQos.TrafficClass != nil &&
			newipdomain.UeDnnQos.TrafficClass != nil {
			if (oldipdomain.UeDnnQos.TrafficClass.Name != newipdomain.UeDnnQos.TrafficClass.Name) ||
				(oldipdomain.UeDnnQos.TrafficClass.Qci != newipdomain.UeDnnQos.TrafficClass.Qci) ||
				(oldipdomain.UeDnnQos.TrafficClass.Arp != newipdomain.UeDnnQos.TrafficClass.Arp) ||
				(oldipdomain.UeDnnQos.TrafficClass.Pdb != newipdomain.UeDnnQos.TrafficClass.Pdb) ||
				(oldipdomain.UeDnnQos.TrafficClass.Pelr != newipdomain.UeDnnQos.TrafficClass.Pelr) {
				return true
			}
		}
	}

	return false
}

func compareNetworkSlice(sliceNew *NetworkSlice, sliceOld *NetworkSlice) bool {
	// slice Id should not change
	appFilteringRulesNew := sliceNew.ApplicationFilteringRules
	appFilteringRulesOld := sliceOld.ApplicationFilteringRules
	if !reflect.DeepEqual(appFilteringRulesNew, appFilteringRulesOld) {
		return true
	}
	for _, ng := range sliceNew.DevGroups {
		found := false
		for _, og := range sliceOld.DevGroups {
			if ng == og {
				found = true
				break
			}
		}
		if !found {
			log.Println("new Dev Group added in slice ")
			return true // 2 network slices have some difference
		}
	}
	for _, ng := range sliceOld.DevGroups {
		found := false
		for _, og := range sliceNew.DevGroups {
			if ng == og {
				found = true
				break
			}
		}
		if !found {
			log.Println("Dev Group Deleted in slice ")
			return true // 2 network slices have some difference
		}
	}
	oldSite := sliceOld.SiteInfo
	newSite := sliceNew.SiteInfo
	if oldSite.SiteName != newSite.SiteName {
		log.Println("site name changed ")
		return true
	}
	oldUpf := oldSite.Upf
	newUpf := newSite.Upf
	if (oldUpf.UpfName != newUpf.UpfName) && (oldUpf.UpfPort != newUpf.UpfPort) {
		log.Println("Upf details changed")
		return true
	}

	for _, newgnb := range newSite.Gnb {
		found := false
		for _, oldgnb := range oldSite.Gnb {
			if newgnb.Name == oldgnb.Name && newgnb.Tac == oldgnb.Tac {
				found = true
				break
			}
		}
		if !found {
			log.Println("gnb changed in slice ")
			return true // change in slice details
		}
	}

	log.Println("No change in slices ")
	return false
}

func UpdateConfig(f string) error {
	if content, err := os.ReadFile(f); err != nil {
		return err
	} else {
		NewSimappConfig := Config{}

		if yamlErr := yaml.Unmarshal(content, &NewSimappConfig); yamlErr != nil {
			return yamlErr
		}
		if NewSimappConfig.Configuration == nil {
			log.Println("Configuration Parsing Failed ", NewSimappConfig.Configuration)
			return nil
		}

		log.Println("Number of subscriber ranges in updated config", len(SimappConfig.Configuration.Subscriber))
		var newImsiList []uint64
		for o := 0; o < len(NewSimappConfig.Configuration.Subscriber); o++ {
			newSubscribers := NewSimappConfig.Configuration.Subscriber[o]
			log.Println("Subscribers:")
			log.Println("    UeIdStart", newSubscribers.UeIdStart)
			log.Println("    UeIdEnd", newSubscribers.UeIdEnd)
			log.Println("    PlmnId", newSubscribers.PlmnId)
			log.Println("    OPc", newSubscribers.OPc)
			log.Println("    OP", newSubscribers.OP)
			log.Println("    Key", newSubscribers.Key)
			log.Println("    SequenceNumber", newSubscribers.SequenceNumber)

			newStart, err := strconv.Atoi(newSubscribers.UeIdStart)
			if err != nil {
				log.Println("error in Atoi with UeIdStart", err)
				continue
			}
			newEnd, err := strconv.Atoi(newSubscribers.UeIdEnd)
			if err != nil {
				log.Println("error in Atoi with UeIdEnd", err)
				continue
			}
			for i := newStart; i <= newEnd; i++ {
				found := false
				newImsiList = append(newImsiList, uint64(i))
				for s := 0; s < len(SimappConfig.Configuration.Subscriber); s++ {
					subscribers := SimappConfig.Configuration.Subscriber[s]
					start, err := strconv.Atoi(subscribers.UeIdStart)
					if err != nil {
						log.Println("error in Atoi with UeIdStart", err)
						continue
					}
					end, err := strconv.Atoi(subscribers.UeIdEnd)
					if err != nil {
						log.Println("error in Atoi with UeIdEnd", err)
						continue
					}
					for j := start; j <= end; j++ {
						if i == j { // two subcribers' imsi are same
							found = true
							if compareSubscriber(newSubscribers, subscribers) {
								log.Println("WARNING: subscriber provision not support modify yet!")
							}
							break
						}
					}
				}
				if found {
					continue
				}
				// add subscriber to chan
				newSubscribers.UeId = fmt.Sprintf("%015d", i)

				b, err := json.Marshal(newSubscribers)
				if err != nil {
					log.Println("error in marshal with newSubscriber", err)
					continue
				}
				reqMsgBody := bytes.NewBuffer(b)
				var msg configMessage
				msg.msgPtr = reqMsgBody
				msg.msgType = subscriber
				msg.name = newSubscribers.UeId
				msg.msgOp = add_op
				configMsgChan <- msg
			}
		}
		// delete all the existing subscribers not show up in new config.
		for o := 0; o < len(SimappConfig.Configuration.Subscriber); o++ {
			subscribers := SimappConfig.Configuration.Subscriber[o]
			start, err := strconv.Atoi(subscribers.UeIdStart)
			if err != nil {
				log.Println("error in Atoi with UeIdStart", err)
				continue
			}
			end, err := strconv.Atoi(subscribers.UeIdEnd)
			if err != nil {
				log.Println("error in Atoi with UeIdEnd", err)
				continue
			}
			for k := start; k <= end; k++ {
				has := false
				for _, v := range newImsiList {
					if v == uint64(k) {
						has = true
					}
				}
				if !has {
					log.Println("going to delete subscriber: ", k)
					b, err := json.Marshal("")
					if err != nil {
						log.Println("error in marshal with subscriber", err)
						continue
					}
					reqMsgBody := bytes.NewBuffer(b)
					var msg configMessage
					msg.msgPtr = reqMsgBody
					msg.msgType = subscriber
					msg.name = fmt.Sprintf("%015d", k)
					msg.msgOp = delete_op
					configMsgChan <- msg
				}
			}
		}
		SimappConfig.Configuration.Subscriber = NewSimappConfig.Configuration.Subscriber
		// end process subscriber update

		for _, group := range SimappConfig.Configuration.DevGroup {
			group.visited = false
		}
		for _, groupNew := range NewSimappConfig.Configuration.DevGroup {
			found := false
			for _, groupOld := range SimappConfig.Configuration.DevGroup {
				if groupNew.Name == groupOld.Name {
					configChange := compareGroup(groupNew, groupOld)
					if configChange {
						// send Group Put
						log.Println("Updated group config ", groupNew.Name)
						dispatchGroup(configMsgChan, groupNew, modify_op)
						// find all slices which are using this device group and mark them modified
						for _, slice := range SimappConfig.Configuration.NetworkSlice {
							for _, dg := range slice.DevGroups {
								if groupOld.Name == dg {
									slice.modified = true
									break
								}
							}
						}
					} else {
						log.Println("Config not updated for group ", groupNew.Name)
					}
					found = true
					groupOld.visited = true
					break
				}
			}
			if !found {
				// new Group - Send Post
				log.Println("New group config ", groupNew.Name)
				dispatchGroup(configMsgChan, groupNew, add_op)
			}
		}
		// visit all groups see if slice is deleted...if found = false
		for _, group := range SimappConfig.Configuration.DevGroup {
			if !group.visited {
				log.Println("Group deleted ", group.Name)
				dispatchGroup(configMsgChan, group, delete_op)
				// find all slices which are using this device group and mark them modified
				for _, slice := range SimappConfig.Configuration.NetworkSlice {
					for _, dg := range slice.DevGroups {
						if group.Name == dg {
							slice.modified = true
							break
						}
					}
				}
			}
		}

		SimappConfig.Configuration.DevGroup = NewSimappConfig.Configuration.DevGroup

		// visit all sliceOld see if slice is deleted...if found = false
		for _, slice := range SimappConfig.Configuration.NetworkSlice {
			slice.visited = false
		}

		for _, sliceNew := range NewSimappConfig.Configuration.NetworkSlice {
			found := false
			for _, sliceOld := range SimappConfig.Configuration.NetworkSlice {
				if sliceNew.Name == sliceOld.Name {
					configChange := compareNetworkSlice(sliceNew, sliceOld)
					if sliceOld.modified {
						log.Println("Updated slice config ", sliceNew.Name)
						sliceOld.modified = false
						dispatchNetworkSlice(configMsgChan, sliceNew, modify_op)
					} else if configChange {
						// send Slice Put
						log.Println("Updated slice config ", sliceNew.Name)
						dispatchNetworkSlice(configMsgChan, sliceNew, modify_op)
					} else {
						log.Println("Config not updated for slice ", sliceNew.Name)
					}
					found = true
					sliceOld.visited = true
					break
				}
			}
			if !found {
				// new Slice - Send Post
				log.Println("New slice config ", sliceNew.Name)
				dispatchNetworkSlice(configMsgChan, sliceNew, add_op)
			}
		}
		// visit all sliceOld see if slice is deleted...if found = false
		for _, slice := range SimappConfig.Configuration.NetworkSlice {
			if !slice.visited {
				log.Println("Slice deleted ", slice.Name)
				dispatchNetworkSlice(configMsgChan, slice, delete_op)
			}
		}
		SimappConfig.Configuration.NetworkSlice = NewSimappConfig.Configuration.NetworkSlice
	}
	return nil
}

func WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("****Config file changed:", e.Name)
		if err := UpdateConfig("config/simapp.yaml"); err != nil {
			log.Println("error in loading updated configuration ", err)
		} else {
			log.Println("****Successfully updated configuration****")
		}
	})
	log.Println("WatchConfig done")
}

func dispatchAllSubscribers(configMsgChan chan configMessage) {
	log.Println("Number of subscriber ranges", len(SimappConfig.Configuration.Subscriber))
	for o := 0; o < len(SimappConfig.Configuration.Subscriber); o++ {
		subscribers := SimappConfig.Configuration.Subscriber[o]
		log.Println("Subscribers:")
		log.Println("    UeIdStart", subscribers.UeIdStart)
		log.Println("    UeIdEnd", subscribers.UeIdEnd)
		log.Println("    PlmnId", subscribers.PlmnId)
		log.Println("    OPc", subscribers.OPc)
		log.Println("    OP", subscribers.OP)
		log.Println("    Key", subscribers.Key)
		log.Println("    SequenceNumber", subscribers.SequenceNumber)

		start, err := strconv.Atoi(subscribers.UeIdStart)
		if err != nil {
			log.Println("error in Atoi with UeIdStart", err)
			continue
		}
		end, err := strconv.Atoi(subscribers.UeIdEnd)
		if err != nil {
			log.Println("error in Atoi with UeIdEnd", err)
			continue
		}
		for i := start; i <= end; i++ {
			subscribers.UeId = fmt.Sprintf("%015d", i)
			log.Println("    UeId", subscribers.UeId)
			//			subscribers.UeIdStart = ""
			//			subscribers.UeIdEnd = ""
			b, err := json.Marshal(subscribers)
			if err != nil {
				log.Println("error in marshal with subscribers", err)
				continue
			}
			reqMsgBody := bytes.NewBuffer(b)
			var msg configMessage
			msg.msgPtr = reqMsgBody
			msg.msgType = subscriber
			msg.name = subscribers.UeId
			msg.msgOp = add_op
			configMsgChan <- msg
		}
	}
}

func dispatchGroup(configMsgChan chan configMessage, group *DevGroup, msgOp int) {
	log.Println("Group Name ", group.Name)
	log.Println("  Site Name ", group.SiteInfo)
	log.Println("  Imsis ", group.Imsis)
	for im := 0; im < len(group.Imsis); im++ {
		log.Println("  IMSI ", group.Imsis[im])
	}
	log.Println("  IpDomainName ", group.IpDomainName)
	ipDomain := group.IpDomain
	if group.IpDomain != nil {
		log.Println("  IpDomain Dnn ", ipDomain.Dnn)
		log.Println("  IpDomain Dns Primary ", ipDomain.DnsPrimary)
		log.Println("  IpDomain Mtu ", ipDomain.Mtu)
		log.Println("  IpDomain UePool ", ipDomain.UePool)
	}
	b, err := json.Marshal(group)
	if err != nil {
		log.Println("error in marshal ", err)
		return
	}
	reqMsgBody := bytes.NewBuffer(b)
	if !SimappConfig.Configuration.ConfigSlice {
		log.Println("Don't configure network slice ")
		return
	}
	var msg configMessage
	msg.msgPtr = reqMsgBody
	msg.msgType = device_group
	msg.name = group.Name
	msg.msgOp = msgOp
	configMsgChan <- msg
}

func dispatchAllGroups(configMsgChan chan configMessage) {
	log.Println("Number of device Groups ", len(SimappConfig.Configuration.DevGroup))
	for _, group := range SimappConfig.Configuration.DevGroup {
		dispatchGroup(configMsgChan, group, add_op)
	}
}

func dispatchNetworkSlice(configMsgChan chan configMessage, slice *NetworkSlice, msgOp int) {
	log.Println("  Slice Name : ", slice.Name)
	fmt.Printf("  Slice sst %v, sd %v", slice.SliceId.Sst, slice.SliceId.Sd)
	log.Println("  Slice site info ", slice.SiteInfo)
	site := slice.SiteInfo
	log.Println("  Slice site name ", site.SiteName)
	log.Println("  Slice gNB ", len(site.Gnb))
	for e := 0; e < len(site.Gnb); e++ {
		fmt.Printf("  Slice gNB[%v] = %v  \n", e, site.Gnb[e])
	}
	log.Println("  Slice Plmn ", site.Plmn)
	log.Println("  Slice Upf ", site.Upf)

	log.Println("  Slice Device Groups ", slice.DevGroups)
	for im := 0; im < len(slice.DevGroups); im++ {
		log.Println("  Attached Device Groups  ", slice.DevGroups[im])
	}

	b, err := json.Marshal(slice)
	if err != nil {
		log.Println("error in marshal ", err)
		return
	}
	reqMsgBody := bytes.NewBuffer(b)

	if !SimappConfig.Configuration.ConfigSlice {
		log.Println("Don't configure network slice ")
		return
	}
	var msg configMessage
	msg.msgPtr = reqMsgBody
	msg.msgType = network_slice
	msg.name = slice.Name
	msg.msgOp = msgOp
	configMsgChan <- msg
}

func dispatchAllNetworkSlices(configMsgChan chan configMessage) {
	log.Println("Number of network Slices ", len(SimappConfig.Configuration.NetworkSlice))
	for _, slice := range SimappConfig.Configuration.NetworkSlice {
		dispatchNetworkSlice(configMsgChan, slice, add_op)
	}
}
