package protocol

import (
	"fmt"
	"sync"
	"time"

	"github.com/VyshnaviMN/onionscan/config"
	"github.com/VyshnaviMN/onionscan/report"
	"github.com/VyshnaviMN/onionscan/resultdb"
	"github.com/VyshnaviMN/onionscan/utils"
)

type OnionQueue struct {
	queue map[string]struct{}
	mux   sync.Mutex
	wg    sync.WaitGroup 
}

// var intervals = []int{5, 10, 20, 30, 40, 60, 120, 240, 480, 720, 960}

func getIntervals() []int{
	var intervals = []int{1, 2}
	var limit = 4000

	for i := 5; i <= limit; i += 5 {
		intervals = append(intervals, 5)
	}

	return intervals
}

var intervals = getIntervals()

func NewOnionQueue() *OnionQueue {
	return &OnionQueue{
		queue: make(map[string]struct{}),
	}
}

func (oq *OnionQueue) AddToQueue(onion string, onionID string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	oq.mux.Lock()

	if oq.queue == nil {
        oq.queue = make(map[string]struct{})
    }

	if _, exists := oq.queue[onion]; !exists {
		oq.queue[onion] = struct{}{}
		oq.wg.Add(1)
		go oq.processQueue(onion, onionID, osc, report)
	}
	oq.mux.Unlock()
}

func (oq *OnionQueue) processQueue(onion string, onionID string, osc *config.OnionScanConfig, report *report.OnionScanReport) {
	
	defer oq.wg.Done()
	
	hiddenService := utils.WithoutProtocol(onion)
	
	intervalWG := sync.WaitGroup{}
    intervalWG.Add(len(intervals))

	for _, interval := range intervals {
		openPorts := ""
		status := getStatus(hiddenService, 80, osc.TorProxyAddress, osc.Timeout)
		if status == "online" {
			openPorts = "80"
		}
		
		db, err := resultdb.InitDB()
		
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		
		if err := resultdb.InsertOrUpdate(db, onionID, onion, time.Now(), openPorts, "80-80", time.Now(), status); err != nil {
			fmt.Printf("Error inserting/updating to database: %v\n", err)
		}
		
		intervalWG.Done()
		db.Close()
		time.Sleep(time.Duration(interval) * time.Minute)
	}
	intervalWG.Wait()
	oq.removeOnion(onion)
}

func (oq *OnionQueue) removeOnion(onion string) {
	oq.mux.Lock()
	delete(oq.queue, onion)
	oq.mux.Unlock()
}