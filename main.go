package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/bigdatadev/goryman"
	"github.com/golang/glog"
	"github.com/google/cadvisor/client"
	"github.com/google/cadvisor/info"
)

var riemannAddress = flag.String("riemann_address", "localhost:5555", "specify the riemann server location")
var cadvisorAddress = flag.String("cadvisor_address", "http://localhost:8080", "specify the cadvisor API server location")
var sampleInterval = flag.Duration("interval", 10*time.Second, "Interval between sampling (default: 10s)")

func pushToRiemann(r *goryman.GorymanClient, service string, metric int, ttl float32, tags []string) {
	err := r.SendEvent(&goryman.Event{
		Service: service,
		Metric:  metric,
		Ttl: 	 ttl,
		Tags:    tags,
	})
	if err != nil {
		glog.Fatalf("unable to write to riemann: %s", err)
	}
}

func main() {
	defer glog.Flush()
	flag.Parse()

	// Setting up the Riemann client
	r := goryman.NewGorymanClient(*riemannAddress)
	err := r.Connect()
	if err != nil {
		glog.Fatalf("unable to connect to riemann: %s", err)
	}
	//defer r.Close()

	// Setting up the cadvisor client
	c, err := client.NewClient(*cadvisorAddress)
	if err != nil {
		glog.Fatalf("unable to setup cadvisor client: %s", err)
	}

	// Setting up the ticker
	ticker := time.NewTicker(*sampleInterval).C
	for {
		select {
		case <-ticker:
			// Make the call to get all the possible data points
			request := info.ContainerInfoRequest{10}
			returned, err := c.AllDockerContainers(&request)
			if err != nil {
				glog.Fatalf("unable to retrieve machine data: %s", err)
			}

			//TODO client.MachineInfo();

			// Start dumping data into riemann
			// Loop into each ContainerInfo
			// Get stats
			// Push into riemann
			for _, container := range returned {
				pushToRiemann(r, fmt.Sprintf("Cpu.Load %s", container.Aliases[0]), int(container.Stats[0].Cpu.Load), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Cpu.Usage.Total %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.Total), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Memory.Usage %s", container.Aliases[0]), int(container.Stats[0].Memory.Usage), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.RxBytes %s", container.Aliases[0]), int(container.Stats[0].Network.RxBytes), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.RxPackets %s", container.Aliases[0]), int(container.Stats[0].Network.RxPackets), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.RxErrors %s", container.Aliases[0]), int(container.Stats[0].Network.RxErrors), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.RxDropped %s", container.Aliases[0]), int(container.Stats[0].Network.RxDropped), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.TxBytes %s", container.Aliases[0]), int(container.Stats[0].Network.TxBytes), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.TxPackets %s", container.Aliases[0]), int(container.Stats[0].Network.TxPackets), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.TxErrors %s", container.Aliases[0]), int(container.Stats[0].Network.TxErrors), float32(10), container.Aliases)
				pushToRiemann(r, fmt.Sprintf("Network.TxDropped %s", container.Aliases[0]), int(container.Stats[0].Network.TxDropped), float32(10), container.Aliases)
			}
		}
	}

}
