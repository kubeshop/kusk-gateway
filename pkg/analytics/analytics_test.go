package analytics

import (
	"os"
	"testing"
)

func TestSendAnonymousInfo(t *testing.T) {

	if val, ok := os.LookupEnv("GA_API_SECRET"); ok {
		KuskGAApiSecret = val
	} else {
		t.Skip()
		return
	}
	if val, ok := os.LookupEnv("GA_MEASUREMENT_ID"); ok {
		KuskGAMeasurementID = val
	} else {
		t.Skip()
		return
	}

	if err := SendAnonymousInfo("analytics_test"); err != nil {
		t.Log(err)
		t.Fail()
	}

}
