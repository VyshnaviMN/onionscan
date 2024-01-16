package protocol

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/resultdb"
	"github.com/VyshnaviMN/onionscan/utils"
	"github.com/schollz/progressbar/v3"
)

type OtherPortsScanner struct {
}

const (
	maxConcurrent = 200
)

func (sps *OtherPortsScanner) ScanProtocol(hiddenService string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	openPorts := ""
	status := ""
	checkPort := 80
	portRange := strings.Join(osc.PortRange, "-")
	startPort, _ := strconv.Atoi(osc.PortRange[0])
	endPort, _ := strconv.Atoi(osc.PortRange[1])

	bar := progressbar.Default(int64(endPort - startPort + 1))

	var wg sync.WaitGroup
	var openPortsMutex sync.Mutex
	var openPortsBuilder strings.Builder

	semaphore := make(chan struct{}, maxConcurrent)

	db, err := resultdb.InitDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer db.Close()

	// if strings.Contains(osc.OnionURL, "https"){
	// 	checkPort = 443
	// } 

	conn, err := utils.GetNetworkConnection(hiddenService, checkPort, osc.TorProxyAddress, osc.Timeout)
	
	if conn != nil {
		conn.Close()
	}
	if err != nil && !strings.Contains(err.Error(), "connection refused") {
		fmt.Printf("Error: %v\n", err)
		if strings.Contains(err.Error(), "host unreachable") {
			fmt.Println("Host is unreachable.")
			status = "offline"
		} else if strings.Contains(err.Error(), "server failure") || strings.Contains(err.Error(), "TTL"){
			fmt.Println("Timeout.")
			status = "temporarily_down"
		} else {
			s := strings.Fields(err.Error())
			status = strings.Join(s[len(s)-2:], "_")
		}
	} else {
		if err != nil && strings.Contains(err.Error(), "connection refused"){
			fmt.Printf("Error: %v\n", err)
			status = "connection_refused_but_scanned"
		} else {
			status = "scanned"
		}

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
					openPortsBuilder.WriteString(strconv.Itoa(port))
					openPortsBuilder.WriteString(", ")
				}
				bar.Add(1)
			}(port)
		}
		wg.Wait()
    }

	openPortsMutex.Lock()
	defer openPortsMutex.Unlock()
	openPorts = strings.TrimSuffix(openPortsBuilder.String(), ", ")

	report.OtherOpenPorts = openPorts
	osc.LogInfo(fmt.Sprintf("Open Ports: %s", openPorts))
	if err := resultdb.InsertOrUpdate(db, report.HiddenServiceID, hiddenService, time.Now(), openPorts, portRange, time.Now(), status); err != nil {
        fmt.Printf("Error inserting/updating to database: %v\n", err)
    }
}
