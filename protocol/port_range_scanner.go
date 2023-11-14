package protocol

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/utils"
)

type OtherPortsScanner struct {
}

func (sps *OtherPortsScanner) ScanProtocol(hiddenService string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	
	var openPorts []string
	startPort,_ := strconv.Atoi(osc.PortRange[0])	
	endPort,_ := strconv.Atoi(osc.PortRange[1])

	for port := startPort; port <= endPort; port++ {
		osc.LogInfo(fmt.Sprintf("Checking %s port(%d)\n", hiddenService, port))
		conn, err := utils.GetNetworkConnection(hiddenService, port, osc.TorProxyAddress, osc.Timeout)
		if conn != nil {
			conn.Close()
		}
		if err != nil {
			osc.LogInfo(fmt.Sprintf("Failed to connect to service on port %d\n", port))
		} else {
			osc.LogInfo(fmt.Sprintf("Found potential service on port %d\n", port))
			openPorts = append(openPorts, strconv.Itoa(port))
		}
	}

	report.OtherOpenPorts = fmt.Sprintf("Open Ports: %s", strings.Join(openPorts, ", "))
}