package analytics

import (
	"os"
	"testing"
)

func TestSendAnonymousInfo(t *testing.T) {
	if val, ok := os.LookupEnv("TELEMETRY_TOKEN"); ok {
		TelemetryToken = val
	} else {
		t.Skip()
		return
	}

	if err := SendAnonymousInfo("analytics_test"); err != nil {
		t.Log(err)
		t.Fail()
	}
}
