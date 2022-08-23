package reporter

import (
	log "github.com/sirupsen/logrus"
	"time"
	"youtube-audio/pkg/util"
)

type GeneralStats struct {
	StartDatetime   string
	StartTimestamp  int64
	DurationSecs    int64
	SuccessfulFetch int64
	FailedFetch     int64
}

var TotalFetch = GeneralStats{}

const (
	SummaryReportTemplate string = "Brief Summary: \n start at: %v\n duration seconds: %v\n successful fetch: %v\n failed fetch: %v"
)

func InitGeneralStats() {
	now := time.Now()
	TotalFetch.StartDatetime = now.Format(util.DateTimeFormat)
	TotalFetch.StartTimestamp = now.Unix()
}

func EndGeneralStats() {
	now := time.Now()
	TotalFetch.DurationSecs = now.Unix() - TotalFetch.StartTimestamp
	log.Infof("TotalFetch2: %v", TotalFetch)
}
