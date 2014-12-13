package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/bigdatadev/goryman"
	"github.com/golang/glog"
	"github.com/google/cadvisor/info"
	"github.com/google/cadvisor/client"
)

var reimannAddress = flag.String("reimann_address", "localhost:5555", "specify the reimann server location")
var cadvisorAddress = flag.String("cadvisor_address", "http://localhost:8080", "specify the cadvisor API server location")
var sampleInterval = flag.Int("sample_interval",1000,"specify the sampling interval")

func main() {
	defer glog.Flush()
	flag.Parse()
	
	// Setting up the Reimann client
	r := goryman.NewGorymanClient(*reimannAddress)
	err := r.Connect()
	if err != nil {
		glog.Fatalf("unable to connect to reimann: %s", err)
	}
	//defer r.Close()

	// Setting up the cadvisor client
	c, err := client.NewClient(*cadvisorAddress)
	if err != nil {
		glog.Fatalf("unable to setup cadvisor client: %s", err)
	}
	
	// Setting up the 1 second ticker 
	ticker := time.NewTicker(1 * time.Second).C
	for {
		select {
		case <-ticker:
			// Make the call to get all the possible data points
			request := info.ContainerInfoRequest{10}
			returned, err := c.AllDockerContainers(&request)
			if err != nil {
				glog.Fatalf("unable to retrieve machine data: %s", err)
			}
			// Start dumping data into reimann
			// Loop into each ContainerInfo
			// Get stats
			// Push into reimann
			fmt.Println(returned[0].Stats[0])
		}
	}
	
}
