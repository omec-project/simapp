// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/free5gc/logger_util"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Info          *Info               `yaml:"info"`
	Configuration *Configuration      `yaml:"configuration"`
	Logger        *logger_util.Logger `yaml:"logger"`
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	ConfigSlice       bool               `yaml:"provision-network-slice,omitempty"`
	DevGroup          []*DevGroup        `yaml:"device-groups,omitempty"`
	NetworkSlice      []*NetworkSlice    `yaml:"network-slices,omitempty"`
	Subscriber        []*Subscriber      `yaml:"subscribers,omitempty"`
	SubProvisionEndpt *SubProvisionEndpt `yaml:"sub-provision-endpt,omitempty"`
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
	UeIdStart      string `yaml:"ueId-start,omitempty" json:"ueId-start,omitempty"`
	UeIdEnd        string `yaml:"ueId-end,omitempty" json:"ueId-end,omitempty"`
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
	Qos                       *QosInfo                     `yaml:"qos,omitempty" json:"qos,omitempty"`
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

type QosInfo struct {
	Uplink       int    `yaml:"uplink,omitempty" json:"uplink,omitempty"`
	Downlink     int    `yaml:"downlink,omitempty" json:"downlink,omitempty"`
	TrafficClass string `yaml:"traffic-class,omitempty" json:"traffic-class,omitempty"`
}

type UeDnnQosInfo struct {
	Uplink       int               `yaml:"dnn-mbr-uplink,omitempty" json:"dnn-mbr-uplink,omitempty"`
	Downlink     int               `yaml:"dnn-mbr-downlink,omitempty" json:"dnn-mbr-downlink,omitempty"`
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

type AppInfo struct {
	AppName   string `yaml:"app-name,omitempty" json:"app-name,omitempty"`
	EndPort   int    `yaml:"end-port,omitempty" json:"end-port,omitempty"`
	EndPoint  string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Protocol  int    `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	StartPort int    `yaml:"start-port,omitempty" json:"start-port,omitempty"`
}

type ApplicationFilteringRules struct {
	// Rule name
	RuleName string `yaml:"rule-name,omitempty" json:"rule-name,omitempty"`
	//priority
	Priority int32 `yaml:"priority,omitempty" json:"priority,omitempty"`
	//action
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
	// Application Desination IP or network
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	//protocol
	Protocol int32 `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	// port range start
	StartPort int32 `yaml:"start-port,omitempty" json:"start-port,omitempty"`
	// port range end
	EndPort int32 `yaml:"end-port,omitempty" json:"end-port,omitempty"`

	AppMbrUplink int32 `yaml:"app-mbr-uplink,omitempty" json:"app-mbr-uplink,omitempty"`

	AppMbrDownlink int32 `yaml:"app-mbr-downlink,omitempty" json:"app-mbr-downlink,omitempty"`

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

type configMessage struct {
	msgPtr  *bytes.Buffer
	msgType int
	name    string
	msgOp   int
}

var SimappConfig Config
var configMsgChan chan configMessage

func InitConfigFactory(f string, configMsgChan chan configMessage, subProvisionEndpt *SubProvisionEndpt) error {
	fmt.Println("Function called ", f)
	if content, err := ioutil.ReadFile(f); err != nil {
		fmt.Println("Readfile failed called ", err)
		return err
	} else {
		SimappConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &SimappConfig); yamlErr != nil {
			fmt.Println("yaml parsing failed ", yamlErr)
			return yamlErr
		}
	}
	if SimappConfig.Configuration == nil {
		fmt.Println("Configuration Parsing Failed ", SimappConfig.Configuration)
		return nil
	}

	fmt.Println("Subscriber Provision Endpoint:")
	fmt.Println("Address ", SimappConfig.Configuration.SubProvisionEndpt.Addr)
	fmt.Println("Port ", SimappConfig.Configuration.SubProvisionEndpt.Port)
	subProvisionEndpt.Addr = SimappConfig.Configuration.SubProvisionEndpt.Addr
	subProvisionEndpt.Port = SimappConfig.Configuration.SubProvisionEndpt.Port

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
	fmt.Fprintf(w, "OK\n")
	dispatchAllGroups(configMsgChan)
	dispatchAllNetworkSlices(configMsgChan)
}

func main() {
	fmt.Println("SimApp started")
	configMsgChan = make(chan configMessage, 100)
	var subProvisionEndpt SubProvisionEndpt

	InitConfigFactory("./config/simapp.yaml", configMsgChan, &subProvisionEndpt)

	go sendMessage(configMsgChan, subProvisionEndpt)
	go WatchConfig()

	dispatchAllSubscribers(configMsgChan)
	dispatchAllGroups(configMsgChan)
	dispatchAllNetworkSlices(configMsgChan)

	http.HandleFunc("/synchronize", syncConfig)
	http.ListenAndServe(":8080", nil)
	for {
		time.Sleep(100 * time.Second)
	}
}

func sendMessage(msgChan chan configMessage, subProvisionEndpt SubProvisionEndpt) {
	var devGroupHttpend string
	var networkSliceHttpend string
	var subscriberHttpend string

	fmt.Println("Subscriber Provision Endpoint in sendMessage:")
	fmt.Println("Address ", subProvisionEndpt.Addr)
	fmt.Println("Port ", subProvisionEndpt.Port)

	ip := strings.TrimSpace(subProvisionEndpt.Addr)

	fmt.Println("webui running at ", ip)
	devGroupHttpend = "http://" + ip + ":" + subProvisionEndpt.Port + "/config/v1/device-group/"
	fmt.Println("device trigger  http endpoint ", devGroupHttpend)
	networkSliceHttpend = "http://" + ip + ":" + subProvisionEndpt.Port + "/config/v1/network-slice/"
	fmt.Println("network slice http endpoint ", devGroupHttpend)
	subscriberHttpend = "http://" + ip + ":" + subProvisionEndpt.Port + "/api/subscriber/imsi-"
	fmt.Println("subscriber http endpoint ", subscriberHttpend)

	for msg := range msgChan {
		var httpend string
		fmt.Println("Received Message from Channel", msgChan, msg)
		switch msg.msgType {
		case device_group:
			httpend = devGroupHttpend + msg.name
		case network_slice:
			httpend = networkSliceHttpend + msg.name
		case subscriber:
			httpend = subscriberHttpend + msg.name
		}

		for {
			if msg.msgOp == add_op {
				fmt.Println("Post Message to ", httpend)
				client := &http.Client{Timeout: 5 * time.Second}
				req, err := http.NewRequest(http.MethodPost, httpend, msg.msgPtr)
				//resp, err := http.Post(httpend, "application/json", msg.msgPtr)
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				resp, err := client.Do(req)
				fmt.Println("Post Message returned ", httpend)
				//Handle Error
				if err != nil {
					fmt.Printf("An Error Occured %v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				defer resp.Body.Close()
				//Read the response body
				_, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					time.Sleep(1 * time.Second)
					continue
				}
				fmt.Printf("Message Post %v Success\n", httpend)
			} else if msg.msgOp == modify_op {
				fmt.Println("PUT Message to ", httpend)
				// initialize http client
				client := &http.Client{Timeout: 5 * time.Second}
				req, err := http.NewRequest(http.MethodPut, httpend, msg.msgPtr)
				//Handle Error
				if err != nil {
					fmt.Printf("An Error Occured %v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				// set the request header Content-Type for json
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Message PUT %v Success\n", resp.StatusCode)
			} else if msg.msgOp == delete_op {
				fmt.Println("DELETE Message to ", httpend)
				// initialize http client
				client := &http.Client{Timeout: 5 * time.Second}
				req, err := http.NewRequest(http.MethodDelete, httpend, msg.msgPtr)
				//Handle Error
				if err != nil {
					fmt.Printf("An Error Occured %v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				// set the request header Content-Type for json
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Message DEL %v Success\n", resp.StatusCode)
			}
			break
		}
	}
}

func compareSubscriber(subscriberNew *Subscriber, subscriberOld *Subscriber) bool {

	if subscriberNew.PlmnId != subscriberOld.PlmnId {
		fmt.Println("Plmn ID changed.")
		return true
	}
	if subscriberNew.OPc != subscriberOld.OPc {
		fmt.Println("OPc changed.")
		return true
	}
	if subscriberNew.OP != subscriberOld.OP {
		fmt.Println("OP changed.")
		return true
	}
	if subscriberNew.Key != subscriberOld.Key {
		fmt.Println("Key changed.")
		return true
	}
	if subscriberNew.SequenceNumber != subscriberOld.SequenceNumber {
		fmt.Println("SequenceNumber changed.")
		return true
	}
	return false
}

func compareGroup(groupNew *DevGroup, groupOld *DevGroup) bool {
	if groupNew.IpDomainName != groupOld.IpDomainName {
		fmt.Println("IP domain name changed.")
		return true
	}

	if groupNew.SiteInfo != groupOld.SiteInfo {
		fmt.Println("SIteInfo name changed.")
		return true
	}

	if len(groupNew.Imsis) != len(groupOld.Imsis) {
		fmt.Println("number of Imsis changed.")
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

	fmt.Println("CODE1 and CODE2 ", strcode1, strcode2)
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
	//slice Id should not change
	qosNew := sliceNew.Qos
	qosOld := sliceOld.Qos
	if qosNew.Uplink != qosOld.Uplink {
		fmt.Println("Uplink Rate changed ")
		return true
	}
	if qosNew.Downlink != qosOld.Downlink {
		fmt.Println("Downlink Rate changed ")
		return true
	}
	if qosNew.TrafficClass != qosOld.TrafficClass {
		fmt.Println("Traffic Class changed ")
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
		if found == false {
			fmt.Println("new Dev Group added in slice ")
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
		if found == false {
			fmt.Println("Dev Group Deleted in slice ")
			return true // 2 network slices have some difference
		}
	}
	oldSite := sliceOld.SiteInfo
	newSite := sliceNew.SiteInfo
	if oldSite.SiteName != newSite.SiteName {
		fmt.Println("site name changed ")
		return true
	}
	oldUpf := oldSite.Upf
	newUpf := newSite.Upf
	if (oldUpf.UpfName != newUpf.UpfName) && (oldUpf.UpfPort != newUpf.UpfPort) {
		fmt.Println("Upf details changed")
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
		if found == false {
			fmt.Println("gnb changed in slice ")
			return true // change in slice details
		}
	}

	fmt.Println("No change in slices ")
	return false
}

func UpdateConfig(f string) error {
	if content, err := ioutil.ReadFile(f); err != nil {
		return err
	} else {
		var NewSimappConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &NewSimappConfig); yamlErr != nil {
			return yamlErr
		}
		if NewSimappConfig.Configuration == nil {
			fmt.Println("Configuration Parsing Failed ", NewSimappConfig.Configuration)
			return nil
		}

		fmt.Println("Number of subscriber ranges in updated config", len(SimappConfig.Configuration.Subscriber))
		var newImsiList []uint64
		for o := 0; o < len(NewSimappConfig.Configuration.Subscriber); o++ {
			newSubscribers := NewSimappConfig.Configuration.Subscriber[o]
			fmt.Println("Subscribers:")
			fmt.Println("    UeIdStart", newSubscribers.UeIdStart)
			fmt.Println("    UeIdEnd", newSubscribers.UeIdEnd)
			fmt.Println("    PlmnId", newSubscribers.PlmnId)
			fmt.Println("    OPc", newSubscribers.OPc)
			fmt.Println("    OP", newSubscribers.OP)
			fmt.Println("    Key", newSubscribers.Key)
			fmt.Println("    SequenceNumber", newSubscribers.SequenceNumber)

			newStart, err := strconv.ParseUint(newSubscribers.UeIdStart, 0, 64)
			if err != nil {
				fmt.Println("error in ParseUint with UeIdStart", err)
				continue
			}
			newEnd, err := strconv.ParseUint(newSubscribers.UeIdEnd, 0, 64)
			if err != nil {
				fmt.Println("error in ParseUint with UeIdEnd", err)
				continue
			}
			for i := newStart; i <= newEnd; i++ {
				found := false
				newImsiList = append(newImsiList, i)
				for s := 0; s < len(SimappConfig.Configuration.Subscriber); s++ {
					subscribers := SimappConfig.Configuration.Subscriber[s]
					start, err := strconv.ParseUint(subscribers.UeIdStart, 0, 64)
					if err != nil {
						fmt.Println("error in ParseUint with UeIdStart", err)
						continue
					}
					end, err := strconv.ParseUint(subscribers.UeIdEnd, 0, 64)
					if err != nil {
						fmt.Println("error in ParseUint with UeIdEnd", err)
						continue
					}
					for j := start; j <= end; j++ {
						if i == j { // two subcribers' imsi are same
							found = true
							if compareSubscriber(newSubscribers, subscribers) == true {
								fmt.Println("WARNING: subscriber provision not support modify yet!")
							}
							break
						}
					}
				}
				if found == true {
					continue
				}
				// add subscriber to chan
				newSubscribers.UeId = strconv.FormatUint(i, 10)
				if err != nil {
					fmt.Println("error in FormatUint with UeId", err)
					continue
				}

				b, err := json.Marshal(newSubscribers)
				if err != nil {
					fmt.Println("error in marshal with newSubscriber", err)
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
		//delete all the exsiting subscribers not show up in new config.
		for o := 0; o < len(SimappConfig.Configuration.Subscriber); o++ {
			subscribers := SimappConfig.Configuration.Subscriber[o]
			start, err := strconv.ParseUint(subscribers.UeIdStart, 0, 64)
			if err != nil {
				fmt.Println("error in ParseUint with UeIdStart", err)
				continue
			}
			end, err := strconv.ParseUint(subscribers.UeIdEnd, 0, 64)
			if err != nil {
				fmt.Println("error in ParseUint with UeIdEnd", err)
				continue
			}
			for k := start; k <= end; k++ {
				has := false
				for _, v := range newImsiList {
					if v == k {
						has = true
					}
				}
				if has == false {
					fmt.Println("going to delete subscriber: ", k)
					b, err := json.Marshal("")
					if err != nil {
						fmt.Println("error in marshal with subscriber", err)
						continue
					}
					reqMsgBody := bytes.NewBuffer(b)
					var msg configMessage
					msg.msgPtr = reqMsgBody
					msg.msgType = subscriber
					msg.name = strconv.FormatUint(k, 10)
					msg.msgOp = delete_op
					configMsgChan <- msg
				}
			}
		}
		SimappConfig.Configuration.Subscriber = NewSimappConfig.Configuration.Subscriber
		//end process subscriber update

		for _, group := range SimappConfig.Configuration.DevGroup {
			group.visited = false
		}
		for _, groupNew := range NewSimappConfig.Configuration.DevGroup {
			found := false
			for _, groupOld := range SimappConfig.Configuration.DevGroup {
				if groupNew.Name == groupOld.Name {
					configChange := compareGroup(groupNew, groupOld)
					if configChange == true {
						// send Group Put
						fmt.Println("Updated group config ", groupNew.Name)
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
						fmt.Println("Config not updated for group ", groupNew.Name)
					}
					found = true
					groupOld.visited = true
					break
				}
			}
			if found == false {
				// new Group - Send Post
				fmt.Println("New group config ", groupNew.Name)
				dispatchGroup(configMsgChan, groupNew, add_op)
			}
		}
		// visit all groups see if slice is deleted...if found = false
		for _, group := range SimappConfig.Configuration.DevGroup {
			if group.visited == false {
				fmt.Println("Group deleted ", group.Name)
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
					if sliceOld.modified == true {
						fmt.Println("Updated slice config ", sliceNew.Name)
						sliceOld.modified = false
						dispatchNetworkSlice(configMsgChan, sliceNew, modify_op)
					} else if configChange == true {
						// send Slice Put
						fmt.Println("Updated slice config ", sliceNew.Name)
						dispatchNetworkSlice(configMsgChan, sliceNew, modify_op)
					} else {
						fmt.Println("Config not updated for slice ", sliceNew.Name)
					}
					found = true
					sliceOld.visited = true
					break
				}
			}
			if found == false {
				// new Slice - Send Post
				fmt.Println("New slice config ", sliceNew.Name)
				dispatchNetworkSlice(configMsgChan, sliceNew, add_op)
			}
		}
		// visit all sliceOld see if slice is deleted...if found = false
		for _, slice := range SimappConfig.Configuration.NetworkSlice {
			if slice.visited == false {
				fmt.Println("Slice deleted ", slice.Name)
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
		fmt.Println("****Config file changed:", e.Name)
		if err := UpdateConfig("config/simapp.yaml"); err != nil {
			fmt.Println("error in loading updated configuration ", err)
		} else {
			fmt.Println("****Successfully updated configuration****")
		}
	})
	fmt.Println("WatchConfig done")
}

func dispatchAllSubscribers(configMsgChan chan configMessage) {
	fmt.Println("Number of subscriber ranges", len(SimappConfig.Configuration.Subscriber))
	for o := 0; o < len(SimappConfig.Configuration.Subscriber); o++ {
		subscribers := SimappConfig.Configuration.Subscriber[o]
		fmt.Println("Subscribers:")
		fmt.Println("    UeIdStart", subscribers.UeIdStart)
		fmt.Println("    UeIdEnd", subscribers.UeIdEnd)
		fmt.Println("    PlmnId", subscribers.PlmnId)
		fmt.Println("    OPc", subscribers.OPc)
		fmt.Println("    OP", subscribers.OP)
		fmt.Println("    Key", subscribers.Key)
		fmt.Println("    SequenceNumber", subscribers.SequenceNumber)

		start, err := strconv.ParseUint(subscribers.UeIdStart, 0, 64)
		if err != nil {
			fmt.Println("error in ParseUint with UeIdStart", err)
			continue
		}
		end, err := strconv.ParseUint(subscribers.UeIdEnd, 0, 64)
		if err != nil {
			fmt.Println("error in ParseUint with UeIdEnd", err)
			continue
		}
		for i := start; i <= end; i++ {
			subscribers.UeId = strconv.FormatUint(i, 10)
			fmt.Println("    UeId", subscribers.UeId)
			if err != nil {
				fmt.Println("error in FormatUint with UeId", err)
				continue
			}
			//			subscribers.UeIdStart = ""
			//			subscribers.UeIdEnd = ""
			b, err := json.Marshal(subscribers)
			if err != nil {
				fmt.Println("error in marshal with subscribers", err)
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
	fmt.Println("Group Name ", group.Name)
	fmt.Println("  Site Name ", group.SiteInfo)
	fmt.Println("  Imsis ", group.Imsis)
	for im := 0; im < len(group.Imsis); im++ {
		fmt.Println("  IMSI ", group.Imsis[im])
	}
	fmt.Println("  IpDomainName ", group.IpDomainName)
	ipDomain := group.IpDomain
	if group.IpDomain != nil {
		fmt.Println("  IpDomain Dnn ", ipDomain.Dnn)
		fmt.Println("  IpDomain Dns Primary ", ipDomain.DnsPrimary)
		fmt.Println("  IpDomain Mtu ", ipDomain.Mtu)
		fmt.Println("  IpDomain UePool ", ipDomain.UePool)
	}
	b, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error in marshal ", err)
		return
	}
	reqMsgBody := bytes.NewBuffer(b)
	if SimappConfig.Configuration.ConfigSlice == false {
		fmt.Println("Don't configure network slice ")
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
	fmt.Println("Number of device Groups ", len(SimappConfig.Configuration.DevGroup))
	for _, group := range SimappConfig.Configuration.DevGroup {
		dispatchGroup(configMsgChan, group, add_op)
	}
}

func dispatchNetworkSlice(configMsgChan chan configMessage, slice *NetworkSlice, msgOp int) {
	fmt.Println("  Slice Name : ", slice.Name)
	fmt.Printf("  Slice sst %v, sd %v", slice.SliceId.Sst, slice.SliceId.Sd)
	fmt.Println("  QoS information ", slice.Qos)
	fmt.Println("  Slice site info ", slice.SiteInfo)
	site := slice.SiteInfo
	fmt.Println("  Slice site name ", site.SiteName)
	fmt.Println("  Slice gNB ", len(site.Gnb))
	for e := 0; e < len(site.Gnb); e++ {
		fmt.Printf("  Slice gNB[%v] = %v  \n", e, site.Gnb[e])
	}
	fmt.Println("  Slice Plmn ", site.Plmn)
	fmt.Println("  Slice Upf ", site.Upf)

	fmt.Println("  Slice Device Groups ", slice.DevGroups)
	for im := 0; im < len(slice.DevGroups); im++ {
		fmt.Println("  Attached Device Groups  ", slice.DevGroups[im])
	}

	b, err := json.Marshal(slice)
	if err != nil {
		fmt.Println("error in marshal ", err)
		return
	}
	reqMsgBody := bytes.NewBuffer(b)

	if SimappConfig.Configuration.ConfigSlice == false {
		fmt.Println("Don't configure network slice ")
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
	fmt.Println("Number of network Slices ", len(SimappConfig.Configuration.NetworkSlice))
	for _, slice := range SimappConfig.Configuration.NetworkSlice {
		dispatchNetworkSlice(configMsgChan, slice, add_op)
	}
}
