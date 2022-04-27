package analytics

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/segmentio/analytics-go"
)

var (
	TelemetryToken = ""
)

func SendAnonymousInfo(event string) error {
	track := analytics.Track{
		AnonymousId: MachineID(),
		Event:       "kusk-heartbeat",
		Properties:  analytics.NewProperties().Set("event", event),
		Timestamp:   time.Now(),
	}

	return sendDataToGA(track)
}

// SendAnonymouscmdInfo will send CLI event to GA
func SendAnonymousCMDInfo() {
	var event string
	command := []string{}
	if len(os.Args) > 1 {
		command = os.Args[1:]
		event = command[0]
	}

	track := analytics.Track{
		AnonymousId: MachineID(),
		UserId:      MachineID(),
		Event:       "kusk-cli",
		Properties:  analytics.NewProperties().Set("event", event),
	}
	sendDataToGA(track)

}

func sendDataToGA(track analytics.Track) error {
	// if environment variable is set return and collect nothing
	if _, ok := os.LookupEnv("KUSK_ANALYTICS_DISABLED"); ok {
		return nil
	}

	client := analytics.New(TelemetryToken)
	defer client.Close()
	return client.Enqueue(track)

}

type Params struct {
	EventCount       int64  `json:"event_count,omitempty"`
	EventCategory    string `json:"even_category,omitempty"`
	AppVersion       string `json:"app_version,omitempty"`
	AppName          string `json:"app_name,omitempty"`
	CustomDimensions string `json:"custom_dimensions,omitempty"`
	DataSource       string `json:"data_source,omitempty"`
}
type Event struct {
	Name   string `json:"name"`
	Params Params `json:"params,omitempty"`
}
type Payload struct {
	ClientID string  `json:"client_id"`
	Events   []Event `json:"events"`
}

// MachineID returns unique user machine ID
func MachineID() string {
	id, _ := generate()
	return id
}

// Generate returns protected id for the current machine
func generate() (string, error) {
	id, err := machineid.ProtectedID("kusk")
	if err != nil {
		return fromHostname()
	}
	return id, err
}

// fromHostname generates a machine id hash from hostname
func fromHostname() (string, error) {
	name, err := os.Hostname()
	if err != nil {
		return "", err
	}
	sum := md5.Sum([]byte(name))
	return hex.EncodeToString(sum[:]), nil
}
