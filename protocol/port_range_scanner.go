package protocol

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"os"
	"net"
	"encoding/csv"
	"time"
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/utils"
	"github.com/schollz/progressbar/v3"
)

type OtherPortsScanner struct {
}

func appendToCSV(outputFile *os.File, id, onion, portRange string, result string, status string) error {
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write CSV row
	row := []string{
		time.Now().Format("2006-01-02 15:04:05"), // DateTime
		id,                                       // OnionID
		onion,                                    // OnionAddress
		result,                                   // Open Ports or Error
		portRange,								  // Port-Range scanned
		status,
	}
	return writer.Write(row)
}

const (
	maxConcurrent = 15
	outputCSV     = "../scan_results/latest_scan.csv"
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

	outputFile, err := os.OpenFile(outputCSV, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening/creating output CSV file %s: %v\n", outputCSV, err)
		return
	}
	defer outputFile.Close()

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

	if len(openPorts) > 0 {
		result := strings.Join(openPorts, ", ")
		osc.LogInfo(fmt.Sprintf("Detected Open Ports: %s", result))
		report.OtherOpenPorts = result
		if err := appendToCSV(outputFile, osc.OnionID, hiddenService, portRange, result, status); err != nil {
			fmt.Printf("Error appending to CSV: %v\n", err)
		}
	} else {
		result := ""
		osc.LogInfo(fmt.Sprintf("No Open Ports Detected"))
		report.OtherOpenPorts = fmt.Sprintf("No Open Ports Detected")
		if err := appendToCSV(outputFile, osc.OnionID, hiddenService, portRange, result, status); err != nil {
			fmt.Printf("Error appending to CSV: %v\n", err)
		}
	}
}