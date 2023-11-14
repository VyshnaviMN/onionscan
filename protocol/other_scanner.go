package protocol

import (
	"fmt"
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"

	"github.com/VyshnaviMN/onionscan/utils"
)

type OtherProtocolScanner struct {
}

func (sps *OtherProtocolScanner) ScanProtocol(hiddenService string, osc *config.OnionScanConfig, report *report.OnionScanReport) {

	// HTTP
	osc.LogInfo(fmt.Sprintf("Checking %s pop3(110)\n", hiddenService))
	conn, err := utils.GetNetworkConnection(hiddenService, 110, osc.TorProxyAddress, osc.Timeout)
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		osc.LogInfo("Failed to connect to service on port 110\n")
		report.OtherDetected = false
	} else {
		osc.LogInfo("Found potential service on pop3(110)\n")
		report.OtherDetected = true
	}
}