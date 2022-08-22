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

	incomingDeliveries := handler.AssembleDeliveriesFromPlaylists()
	mergedDeliveries := handler.MergeHistoryFetchesInto(incomingDeliveries)
	process(mergedDeliveries)

}

func init() {
	util.InitResources()
	log.Infof("base[0]: %v", util.MediaBase[0])
	log.Infof("history[0]: %v", util.MediaHistory[0])
}

func process(deliveries []handler.Delivery) {
	var wg sync.WaitGroup
	var updatedDeliveries []handler.Delivery
	for i, delivery := range deliveries {
		log.Infof("ready to process %v, url: %s", i, delivery.Parcel.Url)
		if i < len(deliveries)-1 { //have to?
			wg.Add(1)
			go func(de handler.Delivery) {
				log.Infof("processing video by NEW routine: %v", de)
				handler.ProcessOneVideo(&de)
				updatedDeliveries = append(updatedDeliveries, de)
				wg.Done()
			}(delivery)
		} else {
			log.Infof("processing video by ORIGINAL routine: %v", delivery)
			handler.ProcessOneVideo(&delivery)
			updatedDeliveries = append(updatedDeliveries, delivery)
			wg.Wait()
		}
	}
	log.Infof("updatedDeliveries: %v", updatedDeliveries)
	handler.FlushFetchHistory(updatedDeliveries)
}
