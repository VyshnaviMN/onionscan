package protocol

import (
	"github.com/CypherpunkSamurai/onionscan/config"
	"github.com/CypherpunkSamurai/onionscan/report"
)

type Scanner interface {
	ScanProtocol(hiddenService string, onionscanConfig *config.OnionScanConfig, report *report.OnionScanReport)
}
