package reporter

import (
	"time"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/log"
)

type GeneralStats struct {
	StartDatetime   string
	StartTimestamp  int64
	DurationSecs    int64
	SuccessfulFetch int64
	FailedFetch     int64
}

var BriefSummary = GeneralStats{}

const (
	SummaryReportTemplate string = "Brief Summary: \n start at: %v\n duration seconds: %v\n successful fetch: %v\n failed fetch: %v"
)

func InitGeneralStats() {
	now := time.Now()
	BriefSummary.StartDatetime = now.Format(util.DateTimeFormat)
	BriefSummary.StartTimestamp = now.Unix()
}

func EndGeneralStats() {
	now := time.Now()
	BriefSummary.DurationSecs = now.Unix() - BriefSummary.StartTimestamp
	log.Infof("BriefSummary: %v", BriefSummary)
}
