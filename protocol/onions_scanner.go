package protocol

import (
	"fmt"
	"sync"
	"time"

	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
)

type OnionsScanner struct {
}

var rescanQueue = NewOnionQueue()

func (sps *OnionsScanner) ScanProtocol(osc *config.OnionScanConfig, report *report.OnionScanReport) {
	var wg sync.WaitGroup
	
	maxOnions := 20
	if len(report.OnionsToScan) < maxOnions {
		maxOnions = len(report.OnionsToScan)
	}
	semaphore := make(chan struct{}, maxOnions)	

	var start time.Time
	var end time.Time

	for onion, onionId := range report.OnionsToScan {
		wg.Add(1)
		semaphore <- struct{}{}
		
		go func(onion string, onionId string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			start = time.Now()
			// Create an instance of the protocol.OtherPortsScanner
			otherPorts := new(OtherPortsScanner)
			otherPorts.ScanProtocol(onion, onionId, osc, report, rescanQueue)
			end = time.Now()
			fmt.Printf("%s:%v\n", onionId, end.Sub(start))
		}(onion, onionId)
	}

	wg.Wait()
	close(semaphore)

	rescanQueue.wg.Wait()
}
