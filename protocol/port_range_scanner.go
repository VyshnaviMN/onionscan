package protocol

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"os"
	"encoding/csv"
	"time"
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/utils"
	"github.com/schollz/progressbar/v3"
)

type OtherPortsScanner struct {
}

func appendToCSV(outputFile *os.File, onion, portRange string, result string) error {
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write CSV row
	row := []string{
		time.Now().Format("2006-01-02 15:04:05"), // DateTime
		onion,                                    // OnionAddress
		result,                                   // Open Ports or Error
		portRange,								  // Port-Range scanned
	}
	return writer.Write(row)
}

func (sps *OtherPortsScanner) ScanProtocol(hiddenService string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	var openPorts []string
	portRange := strings.Join(osc.PortRange, "-")
	startPort, _ := strconv.Atoi(osc.PortRange[0])
	endPort, _ := strconv.Atoi(osc.PortRange[1])

	bar := progressbar.Default(int64(endPort - startPort + 1))

	var wg sync.WaitGroup
	var openPortsMutex sync.Mutex

	maxConcurrent := 15
	semaphore := make(chan struct{}, maxConcurrent)

	outputCSV := "../scan_results/latest_scan.csv"
	outputFile, err := os.OpenFile(outputCSV, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening/creating output CSV file %s: %v\n", outputCSV, err)
		return
	}
	defer outputFile.Close()

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
		}(port)
		bar.Add(1)
	}

	wg.Wait()

	if len(openPorts) > 0 {
		result := strings.Join(openPorts, ", ")
		osc.LogInfo(fmt.Sprintf("Detected Open Ports: %s", result))
		report.OtherOpenPorts = result
		if err := appendToCSV(outputFile, hiddenService, portRange, result); err != nil {
			fmt.Printf("Error appending to CSV: %v\n", err)
		}
	} else {
		result := ""
		osc.LogInfo(fmt.Sprintf("No Open Ports Detected"))
		report.OtherOpenPorts = fmt.Sprintf("No Open Ports Detected")
		if err := appendToCSV(outputFile, hiddenService, portRange, result); err != nil {
			fmt.Printf("Error appending to CSV: %v\n", err)
		}
	}
}