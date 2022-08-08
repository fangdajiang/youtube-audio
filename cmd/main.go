package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"youtube-audio/pkg/handler"
	"youtube-audio/pkg/util"
)

func main() {
	fmt.Printf("Start fetching, converting, sending... from %s\n", time.Now().Format(handler.DateTimeFormat))

	deliveries := handler.MergeIncomingDeliveriesAndHistoryFetches()
	process(deliveries)

}

func init() {
	util.InitResources()
	log.Infof("base: %v", util.MediaBase[0])
	log.Infof("history: %v", util.MediaHistory[0])
}

func process(deliveries []handler.Delivery) {
	var wg sync.WaitGroup
	var updatedDeliveries []handler.Delivery
	for i, delivery := range deliveries {
		log.Infof("%v, url: %s", i, delivery.Parcel.Url)
		if i < len(deliveries)-1 { //have to?
			wg.Add(1)
			go func(de handler.Delivery) {
				handler.ProcessOneVideo(&de)
				updatedDeliveries = append(updatedDeliveries, de)
				wg.Done()
			}(delivery)
		} else {
			handler.ProcessOneVideo(&delivery)
			updatedDeliveries = append(updatedDeliveries, delivery)
			wg.Wait()
		}
	}
	log.Infof("updatedDeliveries: %v", updatedDeliveries)
	handler.FlushFetchHistory(updatedDeliveries)
}
