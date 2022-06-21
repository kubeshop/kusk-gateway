/*
MIT License

Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
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
	TelemetryToken = "" // value needs to be passed with LDFLAG set to github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken
)

func SendAnonymousInfo(event string) error {
	track := analytics.Track{
		AnonymousId: MachineID(),
		Event:       "kusk-heartbeat",
		Properties:  analytics.NewProperties().Set("event", event),
		Timestamp:   time.Now()}

	return sendDataToGA(track)
}

// SendAnonymouscmdInfo will send CLI event to GA
func SendAnonymousCMDInfo() error {
	event := "command"
	command := []string{}
	if len(os.Args) > 1 {
		command = os.Args[1:]
		event = command[0]
	}

	properties := analytics.NewProperties()
	properties.Set("event", event)
	properties.Set("command", command)
	track := analytics.Track{
		AnonymousId: MachineID(),
		UserId:      MachineID(),
		Event:       "kusk-cli",
		Properties:  properties,
	}
	return sendDataToGA(track)
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
