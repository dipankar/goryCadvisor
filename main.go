package main

import (
	"flag"
	"fmt"
	"time"
	"strconv"

	"github.com/bigdatadev/goryman"
	"github.com/golang/glog"
	"github.com/google/cadvisor/client"
	"github.com/google/cadvisor/info"
)

var riemannAddress = flag.String("riemann_address", "localhost:5555", "specify the riemann server location")
var cadvisorAddress = flag.String("cadvisor_address", "http://localhost:8080", "specify the cadvisor API server location")
var sampleInterval = flag.Duration("interval", 10*time.Second, "Interval between sampling (default: 10s)")
var hostEventRiemann = flag.String("riemann_host_event", "", "specify host in riemann event (default '')")

func pushToRiemann(r *goryman.GorymanClient, host string, service string, metric interface{}, ttl float32, tags []string) {
	err := r.SendEvent(&goryman.Event{
		Host:    host,
		Service: service,
		Metric:  metric,
		Ttl:     ttl,
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

			machineInfo, err := c.MachineInfo()
			if err != nil {
				glog.Fatal("unable to getMachineInfo: %s", err)
			}

			// Start dumping data into riemann
			// Loop into each ContainerInfo
			// Get stats
			// Push into riemann
			for _, container := range returned {
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Load %s", container.Aliases[0]), int(container.Stats[0].Cpu.Load), float32(10), container.Aliases)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.Total %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.Total), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.TotalPercent %s", container.Aliases[0]), getCpuTotalPercent(&container.Spec, container.Stats, machineInfo), float32(10), container.Aliases)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.User %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.User), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.System %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.System), float32(10), container.Aliases)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsageMB %s", container.Aliases[0]), getMemoryUsage(container.Stats), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsagePercent %s", container.Aliases[0]), getMemoryUsagePercent(&container.Spec, container.Stats, machineInfo), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsageHotPercent %s", container.Aliases[0]), getHotMemoryPercent(&container.Spec, container.Stats, machineInfo), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsageColdPercent %s", container.Aliases[0]), getColdMemoryPercent(&container.Spec, container.Stats, machineInfo), float32(10), container.Aliases)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxBytes %s", container.Aliases[0]), int(container.Stats[0].Network.RxBytes), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxPackets %s", container.Aliases[0]), int(container.Stats[0].Network.RxPackets), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxErrors %s", container.Aliases[0]), int(container.Stats[0].Network.RxErrors), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxDropped %s", container.Aliases[0]), int(container.Stats[0].Network.RxDropped), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxBytes %s", container.Aliases[0]), int(container.Stats[0].Network.TxBytes), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxPackets %s", container.Aliases[0]), int(container.Stats[0].Network.TxPackets), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxErrors %s", container.Aliases[0]), int(container.Stats[0].Network.TxErrors), float32(10), container.Aliases)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxDropped %s", container.Aliases[0]), int(container.Stats[0].Network.TxDropped), float32(10), container.Aliases)
			}
		}
	}
}

func RoundFloat(x float64, prec int) float64 {
	frep := strconv.FormatFloat(x, 'g', prec, 64)
	f, _ := strconv.ParseFloat(frep, 64)
	return f
}

func getCpuTotalPercent(spec *info.ContainerSpec, stats []*info.ContainerStats, machine *info.MachineInfo) float64 {

	cpuUsage := float64(0)
	if (spec.HasCpu && len(stats) >= 2) {
		cur := stats[len(stats) - 1];
		prev := stats[len(stats) - 2];
		rawUsage := float64(cur.Cpu.Usage.Total - prev.Cpu.Usage.Total);
		intervalInNs := float64(cur.Timestamp.Sub(prev.Timestamp).Nanoseconds());
		// Convert to millicores and take the percentage
		cpuUsage = RoundFloat(((rawUsage / intervalInNs) / float64(machine.NumCores)) * float64(100), 2);
		if (cpuUsage > float64(100)) {
			cpuUsage = float64(100)
		}
	}
	return cpuUsage
}

func toMegabytes(bytes uint64) float64 {
	return float64(bytes) / (1 << 20)
}

func getMemoryUsage(stats []*info.ContainerStats) float64 {
	if len(stats) == 0 {
		return float64(0.0)
	}
	return toMegabytes((stats[len(stats)-1].Memory.Usage))
}

func toMemoryPercent(usage uint64, spec *info.ContainerSpec, machine *info.MachineInfo) int {
	// Saturate limit to the machine size.
	limit := uint64(spec.Memory.Limit)
	if limit > uint64(machine.MemoryCapacity) {
		limit = uint64(machine.MemoryCapacity)
	}

	return int((usage * 100) / limit)
}

func getMemoryUsagePercent(spec *info.ContainerSpec, stats []*info.ContainerStats, machine *info.MachineInfo) int {
	if len(stats) == 0 {
		return 0
	}
	return toMemoryPercent((stats[len(stats)-1].Memory.Usage), spec, machine)
}

func getHotMemoryPercent(spec *info.ContainerSpec, stats []*info.ContainerStats, machine *info.MachineInfo) int {
	if len(stats) == 0 {
		return 0
	}
	return toMemoryPercent((stats[len(stats)-1].Memory.WorkingSet), spec, machine)
}

func getColdMemoryPercent(spec *info.ContainerSpec, stats []*info.ContainerStats, machine *info.MachineInfo) int {
	if len(stats) == 0 {
		return 0
	}
	latestStats := stats[len(stats)-1].Memory
	return toMemoryPercent((latestStats.Usage)-(latestStats.WorkingSet), spec, machine)
}
