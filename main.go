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
var ttlEventRiemann = flag.Int("riemann_ttl_event", 20, "specify host in riemann event in seconds (default 20)")
var thresholdWarning = flag.Int("threshold_warning", 80, "specify threshold of warning (default 80)")
var thresholdCritical = flag.Int("threshold_critical", 95, "specify threshold of critical (default 95)")

func pushToRiemann(r *goryman.GorymanClient, host string, service string, metric interface{}, ttl float32, tags []string, state string) {
	err := r.SendEvent(&goryman.Event{
		Host:    host,
		Service: service,
		Metric:  metric,
		Ttl:     ttl,
		Tags:    tags,
		State:   state,
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

	ttl := float32(*ttlEventRiemann)

    stateEmpty := ""

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
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Load %s", container.Aliases[0]), int(container.Stats[0].Cpu.Load), ttl, container.Aliases, stateEmpty)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.Total %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.Total), ttl, container.Aliases, stateEmpty)

                cpuUsagePercent := getCpuTotalPercent(&container.Spec, container.Stats, machineInfo)
                stateCpu := computeStatePercent(cpuUsagePercent)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.TotalPercent %s", container.Aliases[0]), cpuUsagePercent, ttl, container.Aliases, stateCpu)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.User %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.User), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Cpu.Usage.System %s", container.Aliases[0]), int(container.Stats[0].Cpu.Usage.System), ttl, container.Aliases, stateEmpty)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsageMB %s", container.Aliases[0]), getMemoryUsage(container.Stats), ttl, container.Aliases, stateEmpty)

                memoryUsagePercent := getMemoryUsagePercent(&container.Spec, container.Stats, machineInfo)
                stateMemory := computeStatePercent(float64(memoryUsagePercent))
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsagePercent %s", container.Aliases[0]), memoryUsagePercent, ttl, container.Aliases, stateMemory)
				
                pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsageHotPercent %s", container.Aliases[0]), getHotMemoryPercent(&container.Spec, container.Stats, machineInfo), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Memory.UsageColdPercent %s", container.Aliases[0]), getColdMemoryPercent(&container.Spec, container.Stats, machineInfo), ttl, container.Aliases, stateEmpty)

				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxBytes %s", container.Aliases[0]), int(container.Stats[0].Network.RxBytes), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxPackets %s", container.Aliases[0]), int(container.Stats[0].Network.RxPackets), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxErrors %s", container.Aliases[0]), int(container.Stats[0].Network.RxErrors), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.RxDropped %s", container.Aliases[0]), int(container.Stats[0].Network.RxDropped), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxBytes %s", container.Aliases[0]), int(container.Stats[0].Network.TxBytes), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxPackets %s", container.Aliases[0]), int(container.Stats[0].Network.TxPackets), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxErrors %s", container.Aliases[0]), int(container.Stats[0].Network.TxErrors), ttl, container.Aliases, stateEmpty)
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Network.TxDropped %s", container.Aliases[0]), int(container.Stats[0].Network.TxDropped), ttl, container.Aliases, stateEmpty)
			}

			returnedFS, err := c.ContainerInfo("/",nil)
			if err != nil {
				glog.Fatal("unable to ContainerInfo: %s", err)
			}
			containerStats := returnedFS.Stats[0]
			for _, fs := range containerStats.Filesystem  {
				fsUsagePercent := getFsUsagePercent(fs.Usage,fs.Limit)
				stateFS := computeStatePercent(float64(fsUsagePercent))
				tags := []string{fs.Device}
				pushToRiemann(r, *hostEventRiemann, fmt.Sprintf("Filesystem.UsagePercent %s", fs.Device), fsUsagePercent, ttl, tags, stateFS)
			}

			
			
			
		}
	}
}

func getFsUsagePercent(usage uint64, limite uint64) float64 {
	return roundFloat(float64(usage*100)/float64(limite), 2);
}


func computeStatePercent(value float64) string {
    switch {
        case value > float64(*thresholdCritical): return "critical"
        case value > float64(*thresholdWarning): return "warning"
    }
    return "ok"
}

func roundFloat(x float64, prec int) float64 {
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
		cpuUsage = roundFloat(((rawUsage / intervalInNs) / float64(machine.NumCores)) * float64(100), 2);
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
