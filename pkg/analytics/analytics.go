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
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/denisbrodbeck/machineid"
	kuskBuild "github.com/kubeshop/kusk-gateway/pkg/build"
)

const (
	gaUrl           = "https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s"
	gaValidationUrl = "https://www.google-analytics.com/debug/mp/collect?measurement_id=%s&api_secret=%s"
)

var (
	KuskGAMeasurementID = "" // value needs to be passed with LDFLAG set to github.com/kubeshop/kusk-gateway/pkg/analytics.KuskGAMeasurementID
	KuskGAApiSecret     = "" // value needs to be passed with LDFLAG set to github.com/kubeshop/kusk-gateway/pkg/analytics.KuskGAApiSecret
)

func SendAnonymousInfo(event string) error {
	payload := Payload{
		ClientID: MachineID(),
		Events: []Event{
			{
				Name: event,
				Params: Params{
					EventCount:       1,
					EventCategory:    "api-request",
					AppName:          "kusk-gateway",
					CustomDimensions: event,
					DataSource:       "gateway",
					AppVersion:       kuskBuild.Version,
				},
			},
		},
	}
	return sendDataToGA(payload)
}

// SendAnonymouscmdInfo will send CLI event to GA
func SendAnonymousCMDInfo() error {
	event := "command"
	command := []string{}
	if len(os.Args) > 1 {
		command = os.Args[1:]
		event = command[0]
	}

	payload := Payload{
		ClientID: MachineID(),
		Events: []Event{
			{
				Name: event,
				Params: Params{
					EventCount:       1,
					EventCategory:    "beacon",
					AppName:          "kusk-cli",
					CustomDimensions: strings.Join(command, " "),
					AppVersion:       kuskBuild.Version,
				},
			}},
	}
	return sendDataToGA(payload)
}

func sendDataToGA(data Payload) error {
	// if environment variable is set return and collect nothing
	if _, ok := os.LookupEnv("KUSK_ANALYTICS_DISABLED"); ok {
		return nil
	}
	sendValidationRequest(data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf(gaUrl, KuskGAMeasurementID, KuskGAApiSecret), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	o, _ := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return fmt.Errorf("could not POST, statusCode: %d, body: %s", resp.StatusCode, string(o))
	}
	return nil
}

func sendValidationRequest(payload Payload) (out string, err error) {

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return out, err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf(gaValidationUrl, KuskGAMeasurementID, KuskGAApiSecret), bytes.NewBuffer(jsonData))
	if err != nil {
		return out, err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode > 300 {
		return out, fmt.Errorf("could not POST, statusCode: %d", resp.StatusCode)
	}
	return fmt.Sprintf("status: %d - %s", resp.StatusCode, b), err
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
