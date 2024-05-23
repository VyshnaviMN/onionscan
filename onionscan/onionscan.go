package onionscan

import (
	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/protocol"
	"github.com/VyshnaviMN/onionscan/report"
)

// OnionScan runs the main procol level scans
type OnionScan struct {
	Config *config.OnionScanConfig
}

// GetAllActions returns a list of all possible protocol level  scans.
func (os *OnionScan) GetAllActions() []string {
	return []string{
		"web",
		"tls",
		"ssh",
		"irc",
		"ricochet",
		"ftp",
		"smtp",
		"mongodb",
		"vnc",
		"xmpp",
		"bitcoin",
		"bitcoin_test",
		"litecoin",
		"dogecoin",
		"other",
	}
}

func (os *OnionScan) GetDefaultPortRange() []string {
	return []string{
		"-1",
	}
}

// Do performs all configured protocol level scans in this run.
func (os *OnionScan) Do(osreport *report.OnionScanReport) error {
	onionScan := new(protocol.OnionsScanner)
	onionScan.ScanProtocol(os.Config, osreport)

	if len(osreport.PerformedScans) != 0 {
		osreport.NextAction = osreport.PerformedScans[len(osreport.PerformedScans)-1]
	} else {
		osreport.NextAction = "none"
	}
	return nil
}
