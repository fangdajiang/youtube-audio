package cmd

import (
	"github.com/spf13/cobra"
	"sync"
	"time"
	"youtube-audio/pkg/handler"
	"youtube-audio/pkg/reporter"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/resource"
)

var (
	mode string
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Process YouTube Playlists according to fetch_base.json.",
	Example: `
# Start to process YouTube Playlists.
ya run -m single YOUTUBE_URL
`,
	Run: func(cmd *cobra.Command, args []string) {
		initSetting()

		log.Infof("Start fetching, converting, sending... from %s\n", time.Now().Format(util.DateTimeFormat))

		incomingDeliveries := handler.AssembleDeliveriesFromPlaylists()
		mergedDeliveries := handler.MergeHistoryFetchesInto(incomingDeliveries)
		for _, delivery := range mergedDeliveries {
			log.Debugf("merged delivery: %v", delivery)
		}
		//process(mergedDeliveries)
	},
}

func process(deliveries []handler.Delivery) {
	var wg sync.WaitGroup
	var updatedDeliveries []handler.Delivery
	for i, delivery := range deliveries {
		log.Debugf("ready to process %v, url: %s", i, delivery.Parcel.Url)
		if i < len(deliveries)-1 { //have to?
			wg.Add(1)
			go func(de handler.Delivery) {
				log.Debugf("processing video by NEW routine: %v", de)
				handler.ProcessOneVideo(&de)
				updatedDeliveries = append(updatedDeliveries, de)
				wg.Done()
			}(delivery)
		} else {
			log.Debugf("processing video by ORIGINAL routine: %v", delivery)
			handler.ProcessOneVideo(&delivery)
			updatedDeliveries = append(updatedDeliveries, delivery)
			wg.Wait()
		}
	}
	reporter.EndGeneralStats()
	handler.SendSummary()
	for _, delivery := range updatedDeliveries {
		log.Debugf("processed delivery: %v", delivery)
	}
	handler.FlushFetchHistory(updatedDeliveries)
	util.UploadLog2Oss()
}

func initSetting() {
	log.InitLogging()
	resource.InitResources()
	log.Infof("base[0]: %v", resource.MediaBase[0])
	log.Infof("history[0]: %v", resource.MediaHistory[0])
	reporter.InitGeneralStats()
}

func init() {
	RunCmd.Flags().StringVarP(&mode, "mode", "m", "", "Mode for running: all or single.")
	RunCmd.Flags().BoolP("help", "h", false, "Print this help message.")
	_ = RunCmd.MarkFlagRequired("mode")
	RootCmd.AddCommand(RunCmd)
}