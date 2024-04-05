package protocol

import (
	"sync"

	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
)

type OnionsScanner struct {
}

var rescanQueue = NewOnionQueue()

func (sps *OnionsScanner) ScanProtocol(osc *config.OnionScanConfig, report *report.OnionScanReport) {
	var wg sync.WaitGroup
	
	maxOnions := 50
	if len(report.OnionsToScan) < maxOnions {
		maxOnions = len(report.OnionsToScan)
	}
	semaphore := make(chan struct{}, maxOnions)	

	for onion, onionId := range report.OnionsToScan {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(onion string, onionId string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// Create an instance of the protocol.OtherPortsScanner
			otherPorts := new(OtherPortsScanner)
			otherPorts.ScanProtocol(onion, onionId, osc, report, rescanQueue)

		}(onion, onionId)
	}

	wg.Wait()
	close(semaphore)

	rescanQueue.wg.Wait()
}
