package main

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/soniah/gosnmp"
)

type TargetStatus struct {
	PollSuccess bool                     `json:"pollSuccess"`
	HostName    string                   `json:"hostname"`
	Interfaces  map[int]*InterfaceStatus `json:"interfaces"`
}

type InterfaceStatusUpdate struct {
	ID         string `json:"id"`
	IfIndex    int    `json:"ifIndex"`
	OperStatus int    `json:"operStatus"`
}

type InterfaceStatus struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        int              `json:"type"`
	OperStatus  int              `json:"operStatus"`
	HighSpeed   uint             `json:"highSpeed"`
	Traffic     InterfaceTraffic `json:"traffic"`
}

type InterfaceTraffic struct {
	InOctets    uint   `json:"inOctets"`
	OutOctets   uint   `json:"outOctets"`
	HCInOctets  uint64 `json:"hcInOctets"`
	HCOutOctets uint64 `json:"hcOutOctets"`
	InDiscards  uint   `json:"inDiscards"`
	OutDiscards uint   `json:"outDiscards"`
	InErrors    uint   `json:"inErrors"`
	OutErrors   uint   `json:"outErrors"`
}

var (
	sysNameOIB             = ".1.3.6.1.2.1.1.5.0"       //poll
	ifIndexOIB             = ".1.3.6.1.2.1.2.2.1.1"     //walk
	ifDescrOIBPrefix       = ".1.3.6.1.2.1.2.2.1.2"     //poll
	ififAliasOIBPrefix     = ".1.3.6.1.2.1.31.1.1.1.18" //poll
	ifTypeOIBPrefix        = ".1.3.6.1.2.1.2.2.1.3"     //poll
	ifAdminStatusOIBPrefix = ".1.3.6.1.2.1.2.2.1.7"     //poll
	ifOperStatusOIBPrefix  = ".1.3.6.1.2.1.2.2.1.8"     //poll,trap
	ifHighSpeedOIBPrefix   = ".1.3.6.1.2.1.31.1.1.1.15" //poll
	ifInOctetsOIBPrefix    = ".1.3.6.1.2.1.2.2.1.10"    //poll
	ifOutOctetsOIBPrefix   = ".1.3.6.1.2.1.2.2.1.16"    //poll
	ifHCInOctetsOIBPrefix  = ".1.3.6.1.2.1.31.1.1.1.6"  //poll
	ifHCOutOctetsOIBPrefix = ".1.3.6.1.2.1.31.1.1.1.10" //poll
	ifInDiscardsOIBPrefix  = ".1.3.6.1.2.1.2.2.1.13"    //poll
	ifOutDiscardsOIBPrefix = ".1.3.6.1.2.1.2.2.1.19"    //poll
	ifInErrorsOIBPrefix    = ".1.3.6.1.2.1.2.2.1.14"    //poll
	ifOutErrorsOIBPrefix   = ".1.3.6.1.2.1.2.2.1.20"    //poll
)

var trafficCount = map[string]*TargetStatus{}

func initSNMP(config *Config, eventCollector chan Event) error {
	go startPolling(config, eventCollector)
	go startTrap(config, eventCollector)
	return nil
}

func startTrap(config *Config, eventCollector chan Event) {
	tl := gosnmp.NewTrapListener()
	regex := regexp.MustCompile(fmt.Sprintf("%s\\.([0-9]+)", ifOperStatusOIBPrefix))
	tl.OnNewTrap = func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		for _, v := range packet.Variables {
			if strings.HasPrefix(v.Name, ifOperStatusOIBPrefix) {
				ifIndex, err := strconv.Atoi(string(regex.FindSubmatch([]byte(v.Name))[1]))
				if err != nil {
					continue
					//TODO: error log
				}
				target := config.lookupFromIP(addr.IP.String())
				if target == nil {
					continue
				}
				eventCollector <- Event{"trap_interface_state", InterfaceStatusUpdate{
					ID:         target.ID,
					IfIndex:    ifIndex,
					OperStatus: v.Value.(int),
				}}
			}
		}
	}
	tl.Params = gosnmp.Default

	err := tl.Listen("0.0.0.0:162")
	if err != nil {
		panic(err)
	}
}

func startPolling(config *Config, eventCollector chan Event) {
	for {
		go func() {
			targets := map[string]*TargetStatus{}
			for _, target := range config.Targets {
				var community string
				if len(target.Community) == 0 {
					community = config.Community
				} else {
					community = target.Community
				}
				status, err := getTargetStatus(target, community)
				if err != nil {
					targets[target.ID] = &TargetStatus{
						PollSuccess: false,
					}
					continue
				}
				targets[target.ID] = status
			}
			eventCollector <- Event{
				Channel: "poll_target",
				Payload: targets,
			}
		}()
		time.Sleep(1 * time.Second)
	}
}

func getTargetStatus(target Target, community string) (*TargetStatus, error) {
	status := TargetStatus{
		PollSuccess: true,
	}
	params := &gosnmp.GoSNMP{
		Target:    target.IP,
		Port:      target.Port,
		Community: community,
		Version:   gosnmp.Version2c,
		Retries:   0,
		Timeout:   time.Duration(1) * time.Second,
	}
	err := params.Connect()
	if err != nil {
		return nil, err
	}
	defer params.Conn.Close()
	result, err := params.Get([]string{sysNameOIB})
	if err != nil {
		return nil, err
	}

	status.HostName = string(result.Variables[0].Value.([]byte))
	status.Interfaces = map[int]*InterfaceStatus{}

	if _, ok := trafficCount[target.ID]; !ok {
		trafficCount[target.ID] = &TargetStatus{
			Interfaces: map[int]*InterfaceStatus{},
		}
	}

	err = params.BulkWalk(ifIndexOIB, func(pdu gosnmp.SnmpPDU) error {
		ifIndex := pdu.Value.(int)
		if _, ok := trafficCount[target.ID].Interfaces[ifIndex]; !ok {
			trafficCount[target.ID].Interfaces[ifIndex] = &InterfaceStatus{Traffic: InterfaceTraffic{}}
		}
		interfaceStatus, err := getInterfaceStatus(target.ID, params, ifIndex)
		if err != nil {
			return err
		}
		status.Interfaces[ifIndex] = interfaceStatus
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func getInterfaceStatus(id string, params *gosnmp.GoSNMP, ifIndex int) (*InterfaceStatus, error) {
	getOIDs := []string{
		fmt.Sprintf("%s.%d", ifDescrOIBPrefix, ifIndex),       // 0
		fmt.Sprintf("%s.%d", ififAliasOIBPrefix, ifIndex),     // 1
		fmt.Sprintf("%s.%d", ifTypeOIBPrefix, ifIndex),        // 2
		fmt.Sprintf("%s.%d", ifAdminStatusOIBPrefix, ifIndex), // 3
		fmt.Sprintf("%s.%d", ifOperStatusOIBPrefix, ifIndex),  // 4
		fmt.Sprintf("%s.%d", ifHighSpeedOIBPrefix, ifIndex),   // 5
		fmt.Sprintf("%s.%d", ifInOctetsOIBPrefix, ifIndex),    // 6
		fmt.Sprintf("%s.%d", ifOutOctetsOIBPrefix, ifIndex),   // 7
		fmt.Sprintf("%s.%d", ifHCInOctetsOIBPrefix, ifIndex),  // 8
		fmt.Sprintf("%s.%d", ifHCOutOctetsOIBPrefix, ifIndex), // 9
		fmt.Sprintf("%s.%d", ifInDiscardsOIBPrefix, ifIndex),  // 10
		fmt.Sprintf("%s.%d", ifOutDiscardsOIBPrefix, ifIndex), // 11
		fmt.Sprintf("%s.%d", ifInErrorsOIBPrefix, ifIndex),    // 12
		fmt.Sprintf("%s.%d", ifOutErrorsOIBPrefix, ifIndex),   // 13
	}
	result, err := params.Get(getOIDs)
	if err != nil {
		return nil, err
	}
	interfaceStatus := InterfaceStatus{}
	interfaceStatus.Name = string(result.Variables[0].Value.([]byte))
	interfaceStatus.Description = string(result.Variables[1].Value.([]byte))
	interfaceStatus.Type = result.Variables[2].Value.(int)
	interfaceStatus.OperStatus = result.Variables[4].Value.(int)
	interfaceStatus.HighSpeed = result.Variables[5].Value.(uint)

	interfaceStatus.Traffic = InterfaceTraffic{}

	if result.Variables[6].Value != nil {
		currentCount := result.Variables[6].Value.(uint)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.InOctets = currentCount - interfaceTrafficCount.Traffic.InOctets
		interfaceTrafficCount.Traffic.InOctets = currentCount
	}

	if result.Variables[7].Value != nil {
		currentCount := result.Variables[7].Value.(uint)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.OutOctets = currentCount - interfaceTrafficCount.Traffic.OutOctets
		interfaceTrafficCount.Traffic.OutOctets = currentCount
	}

	if result.Variables[8].Value != nil {
		currentCount := result.Variables[8].Value.(uint64)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.HCInOctets = currentCount - interfaceTrafficCount.Traffic.HCInOctets
		interfaceTrafficCount.Traffic.HCInOctets = currentCount
	}

	if result.Variables[9].Value != nil {
		currentCount := result.Variables[9].Value.(uint64)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.HCOutOctets = currentCount - interfaceTrafficCount.Traffic.HCOutOctets
		interfaceTrafficCount.Traffic.HCOutOctets = currentCount
	}

	if result.Variables[10].Value != nil {
		currentCount := result.Variables[10].Value.(uint)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.InDiscards = currentCount - interfaceTrafficCount.Traffic.InDiscards
		interfaceTrafficCount.Traffic.InDiscards = currentCount
	}

	if result.Variables[11].Value != nil {
		currentCount := result.Variables[11].Value.(uint)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.OutDiscards = currentCount - interfaceTrafficCount.Traffic.OutDiscards
		interfaceTrafficCount.Traffic.OutDiscards = currentCount
	}

	if result.Variables[12].Value != nil {
		currentCount := result.Variables[12].Value.(uint)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.InErrors = currentCount - interfaceTrafficCount.Traffic.InErrors
		interfaceTrafficCount.Traffic.InErrors = currentCount
	}

	if result.Variables[13].Value != nil {
		currentCount := result.Variables[13].Value.(uint)
		interfaceTrafficCount := trafficCount[id].Interfaces[ifIndex]
		interfaceStatus.Traffic.OutErrors = currentCount - interfaceTrafficCount.Traffic.OutErrors
		interfaceTrafficCount.Traffic.OutErrors = currentCount
	}
	return &interfaceStatus, nil
}
