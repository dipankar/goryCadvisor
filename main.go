package main

import (
	"flag"
	"fmt"
	"github.com/bigdatadev/goryman"
	"github.com/golang/glog"
	cadvisor "github.com/google/cadvisor/client"
	"time"
)

var reimannAddress = flag.String("reimann_address", "localhost:5555", "specify the reimann server location")
var cadvisorAddress = flag.String("cadvisor_address", "http://localhost:8080", "specify the cadvisor API server location")

func main() {
	defer glog.Flush()
	flag.Parse()

	c := goryman.NewGorymanClient(*reimannAddress)
	err := c.Connect()
	if err != nil {
		glog.Fatalf("unable to connect to reimann: %s", err)
	}
	defer c.Close()

	client, err := cadvisor.NewClient(*cadvisorAddress)
	if err != nil {
		glog.Fatalf("unable to setup cadvisor client: %s", err)
	}
	ticker := time.NewTicker(1 * time.Second).C
	for {
		select {
		case <-ticker:
			returned, err := client.MachineInfo()
			if err != nil {
				glog.Fatalf("unable to retrieve machine data: %s", err)
			}
			fmt.Println(returned)
		}
	}
}
