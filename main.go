package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Community   string       `json:"community"`
	Targets     []Target     `json:"targets"`
	Connections []Connection `json:"connections"`
}

func (config *Config) lookupFromIP(IP string) *Target {
	for _, target := range config.Targets {
		if target.IP == IP {
			return &target
		}
	}
	return nil
}

func (config *Config) generateTopology() *Topology {
	topology := &Topology{
		Targets:     map[string]TopologyTarget{},
		Connections: []TopologyConnection{},
	}
	for _, target := range config.Targets {
		topology.Targets[target.ID] = TopologyTarget{
			IP:   target.IP,
			Port: target.Port,
			X:    target.X,
			Y:    target.Y,
		}
	}
	for _, connection := range config.Connections {
		topology.Connections = append(topology.Connections, TopologyConnection{
			From:   connection.From,
			FromIf: connection.FromIf,
			To:     connection.To,
			ToIf:   connection.ToIf,
		})
	}
	return topology
}

type Topology struct {
	Targets     map[string]TopologyTarget `json:"targets"`
	Connections []TopologyConnection      `json:"connections"`
}

type TopologyTarget struct {
	IP   string `json:"ip"`
	Port uint16 `json:"port"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type TopologyConnection struct {
	From   string `json:"from"`
	FromIf string `json:"fromIf"`
	To     string `json:"to"`
	ToIf   string `json:"toIf"`
}

type Target struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	Port      uint16 `json:"port"`
	Community string `json:"community"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
}

type Connection struct {
	From   string `json:"from"`
	FromIf string `json:"fromIf"`
	To     string `json:"to"`
	ToIf   string `json:"toIf"`
}

type Event struct {
	Channel string      `json:"channel"`
	Payload interface{} `json:"payload"`
}

var confFile *string

func main() {
	confFile = flag.String("config", "./config.toml", "config file path")
	flag.Parse()

	config := &Config{}
	_, err := toml.DecodeFile(*confFile, config)
	if err != nil {
		panic(err)
	}

	eventCollector := make(chan Event)

	err = initSNMP(config, eventCollector)
	if err != nil {
		panic(err)
	}
	err = initAPI(config, eventCollector)
	if err != nil {
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	for {
		_, ok := <-signalChan
		if ok {
			reload()
		} else {
			panic(fmt.Errorf(""))
		}
	}
}

func reload() {
	fmt.Println("Not implemented")
}
