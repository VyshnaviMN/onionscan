package protocol

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"net"
	"time"
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/utils"
	"github.com/VyshnaviMN/onionscan/resultdb"
	"github.com/schollz/progressbar/v3"
)

type OtherPortsScanner struct {
}

const (
	maxConcurrent = 15
)

func (sps *OtherPortsScanner) ScanProtocol(hiddenService string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	var openPorts []string
	status := "Scanned"
	portRange := strings.Join(osc.PortRange, "-")
	startPort, _ := strconv.Atoi(osc.PortRange[0])
	endPort, _ := strconv.Atoi(osc.PortRange[1])

	bar := progressbar.Default(int64(endPort - startPort + 1))

	var wg sync.WaitGroup
	var openPortsMutex sync.Mutex

	semaphore := make(chan struct{}, maxConcurrent)

	err := resultdb.InitDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer resultdb.CloseDB()

	conn, err := utils.GetNetworkConnection(hiddenService, 80, osc.TorProxyAddress, osc.Timeout)
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if strings.Contains(err.Error(), "host unreachable") {
			fmt.Println("Host is unreachable.")
			status = "Offline"
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("Timeout.")
			status = "Timeout"
		}
	} else {
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
					openPortsMutex.Lock()
					openPorts = append(openPorts, strconv.Itoa(port))
					openPortsMutex.Unlock()
				}
				bar.Add(1)
			}(port)
		}
		wg.Wait()
    }

	result := strings.Join(openPorts, ", ")
	report.OtherOpenPorts = result
	osc.LogInfo(fmt.Sprintf("Open Ports: %s", result))
	if err := resultdb.InsertScanResult(osc.OnionID, hiddenService, time.Now(), result, portRange, time.Now(), status); err != nil {
        fmt.Printf("Error inserting/updating to database: %v\n", err)
    }
}