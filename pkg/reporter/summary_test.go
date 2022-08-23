package reporter

import (
	"testing"
	"time"
)

func Test_sendSummary(t *testing.T) {
	InitGeneralStats()
	time.Sleep(time.Second * 3)
	TotalFetch.FailedFetch++
	EndGeneralStats()
}
