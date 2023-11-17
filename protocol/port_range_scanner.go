package protocol

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/utils"
	"github.com/schollz/progressbar/v3"
)

type OtherPortsScanner struct {
}

func (sps *OtherPortsScanner) ScanProtocol(hiddenService string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	var openPorts []string
	startPort, _ := strconv.Atoi(osc.PortRange[0])
	endPort, _ := strconv.Atoi(osc.PortRange[1])

	bar := progressbar.Default(int64(endPort - startPort + 1))

	var wg sync.WaitGroup

	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)

	for port := startPort; port <= endPort; port++ {

		wg.Add(1)
		semaphore <- struct{}{}

		go func(port int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			conn, err := utils.GetNetworkConnection(hiddenService, port, osc.TorProxyAddress, osc.Timeout)
			if conn != nil {
				conn.Close()
			}
			if err == nil {
				openPorts = append(openPorts, strconv.Itoa(port))
			}
		}(port)
		bar.Add(1)
	}

	wg.Wait()

	if len(openPorts) > 0 {
		osc.LogInfo(fmt.Sprintf("Detected Open Ports: %s", strings.Join(openPorts, ", ")))
		report.OtherOpenPorts = fmt.Sprintf("Open Ports: %s", strings.Join(openPorts, ", "))
	} else {
		osc.LogInfo(fmt.Sprintf("No Open Ports Detected"))
		report.OtherOpenPorts = fmt.Sprintf("No Open Ports Detected")
	}
}