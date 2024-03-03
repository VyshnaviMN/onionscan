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
)

type OtherPortsScanner struct {
}

var defaultPorts = []int{11, 18, 21, 22, 23, 25, 52, 53, 67, 70, 79, 80, 81, 82, 83, 84, 110, 113, 119, 143, 194, 195, 222, 235, 300, 418, 433, 443, 444, 465, 513, 587, 701, 853, 873, 993, 994, 995, 1234, 1337, 1813, 1935, 1965, 1990, 2083, 2221, 2222, 2345, 2628, 3000, 3012, 3333, 4190, 4242, 4422, 5000, 5060, 5061, 5201, 5202, 5222, 5223, 5269, 5280, 5281, 5400, 5443, 5555, 6660, 6661, 6662, 6663, 6664, 6665, 6666, 6667, 6668, 6669, 6679, 6690, 6691, 6692, 6693, 6694, 6695, 6696, 6697, 6698, 6699, 7000, 7070, 7777, 7778, 7779, 7890, 8000, 8001, 8067, 8080, 8081, 8085, 8090, 8332, 8333, 8334, 8443, 8444, 8446, 8448, 8667, 8888, 9000, 9236, 9735, 9736, 9737, 9788, 9889, 9911, 9999, 10009, 10010, 10022, 11121, 11236, 16667, 18080, 18081, 18083, 18084, 18085, 18089, 18090, 18180, 28080, 28081, 28083, 28089, 28332, 28333, 30029, 34568, 38081, 38089, 45871, 50001, 50002, 64738}

func makeRange(start, end int) []int {
    size := end - start + 1
    if size <= 0 {
        return nil
    }

    slice := make([]int, size)
    for i := range slice {
        slice[i] = start + i
    }
    return slice
}

func (sps *OtherPortsScanner) ScanProtocol(onion string, onionId string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	openPorts := ""
	status := ""
	portRange := "Default"
	checkPort := 80
	maxConcurrent := 200
	if len(osc.PortRange) == 2 {
		portRange = strings.Join(osc.PortRange, "-")
		startPort, _ := strconv.Atoi(osc.PortRange[0])
		endPort, _ := strconv.Atoi(osc.PortRange[1])
		defaultPorts = makeRange(startPort, endPort)
	}
	hiddenService := utils.WithoutProtocol(onion)

	var wg sync.WaitGroup
	var openPortsMutex sync.Mutex
	var openPortsBuilder strings.Builder
	if len(defaultPorts) < 200 {
		maxConcurrent = len(defaultPorts)
	}
	semaphore := make(chan struct{}, maxConcurrent)

	db, err := resultdb.InitDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer db.Close()

	if strings.Contains(onion, "https"){
		checkPort = 443
	}

	conn, err := utils.GetNetworkConnection(hiddenService, checkPort, osc.TorProxyAddress, osc.Timeout)
	
	if conn != nil {
		conn.Close()
	}
	if err != nil && !strings.Contains(err.Error(), "connection refused") {
		// fmt.Printf("Error: %v\n", err)
		if strings.Contains(err.Error(), "host unreachable") {
			// fmt.Println("Host is unreachable.")
			status = "offline"
		} else if strings.Contains(err.Error(), "server failure") || strings.Contains(err.Error(), "TTL"){
			// fmt.Println("Timeout.")
			status = "temporarily_down"
		} else {
			s := strings.Fields(err.Error())
			status = strings.Join(s[len(s)-2:], "_")
		}
	} else {
		if err != nil && strings.Contains(err.Error(), "connection refused"){
			// fmt.Printf("Error: %v\n", err)
			status = "connection_refused_but_scanned"
		} else {
			status = "scanned"
		}

		for port := range defaultPorts {
			
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
			}(port)
		}
		wg.Wait()
    }

	openPortsMutex.Lock()
	defer openPortsMutex.Unlock()
	openPorts = strings.TrimSuffix(openPortsBuilder.String(), ", ")

	osc.LogInfo(fmt.Sprintf("%s: %s\n\n", onion, openPorts))
	if err := resultdb.InsertOrUpdate(db, onionId, onion, time.Now(), openPorts, portRange, time.Now(), status); err != nil {
        fmt.Printf("Error inserting/updating to database: %v\n", err)
    }
}
