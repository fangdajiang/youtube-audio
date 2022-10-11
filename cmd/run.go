package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"sync"
	"time"
	"youtube-audio/pkg/handler"
	"youtube-audio/pkg/reporter"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/resource"
)

const (
	YOUTUBE_PLAYLIST_PREFIX_URL string = "https://www.youtube.com/playlist?list="
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
ya run -m latest
ya run -m playlist YOUTUBE_PLAYLIST_URL
`,
	Run: func(cmd *cobra.Command, args []string) {
		if mode == "" || (mode != "" && mode != "latest" && mode != "single" && mode != "playlist") {
			_, err := fmt.Fprintf(os.Stdout, "An invalid mode was specified.\n")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			os.Exit(1)
		}
		initSetting()
		switch mode {
		case "latest":
			log.Infof("Start fetching, converting, sending... from %s\n", time.Now().Format(util.DateTimeFormat))

			incomingDeliveries := handler.AssembleDeliveriesFromPlaylists(handler.GetYouTubeVideosFromPlaylists())
			mergedDeliveries := handler.MergeHistoryFetchesInto(incomingDeliveries)
			for _, delivery := range mergedDeliveries {
				log.Debugf("merged delivery: %v", delivery)
			}
			process(mergedDeliveries)
		case "playlist":
			if len(args) == 0 {
				fmt.Printf("YouTube Playlist Url Not Specified\n")
				os.Exit(1)
			}
			url := args[0]
			if strings.HasPrefix(url, YOUTUBE_PLAYLIST_PREFIX_URL) {
				log.Infof("processing one PLAYLIST url: %v", url)
				playlistId := strings.TrimPrefix(url, YOUTUBE_PLAYLIST_PREFIX_URL)
				if len(playlistId) < 20 {
					fmt.Printf("Invalid Playlist Id's length of YouTube Playlist Url\n")
					os.Exit(1)
				}
				playlistMetaData := handler.GetYouTubeVideosFromPlaylistId(playlistId)
				incomingDeliveries := handler.AssembleDeliveriesFromPlaylists(playlistMetaData)
				process(incomingDeliveries)
			} else {
				fmt.Printf("Invalid YouTube Playlist Url Format\n")
				os.Exit(1)
			}
		case "single":
			if len(args) == 0 {
				fmt.Printf("YouTube Url Not Specified\n")
				os.Exit(1)
			}
			url := args[0]
			if strings.HasPrefix(url, util.GetYouTubeFetchBase().PrefixUrl) {
				de := handler.AssembleDeliveryFromSingleUrl(url)
				log.Infof("processing one SINGLE url: %v", de)
				handler.ProcessOneVideo(&de)
			} else {
				fmt.Printf("Invalid YouTube Url Format\n")
				os.Exit(1)
			}
		default:
			os.Exit(1)
		}
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
	RunCmd.Flags().StringVarP(&mode, "mode", "m", "", "Mode for running: latest or single or playlist.")
	RunCmd.Flags().BoolP("help", "h", false, "Print this help message.")
	_ = RunCmd.MarkFlagRequired("mode")
	RootCmd.AddCommand(RunCmd)
}
