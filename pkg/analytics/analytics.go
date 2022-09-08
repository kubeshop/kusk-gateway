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
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/segmentio/analytics-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/pkg/build"
)

var (
	TelemetryToken = "" // value needs to be passed with LDFLAG set to github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken
)

func SendAnonymousInfo(ctx context.Context, client client.Client, event string, message string) error {
	properties := analytics.NewProperties()
	properties.Set("message", message)
	properties.Set("version", build.Version)

	track := analytics.Track{
		AnonymousId: ClusterID(ctx, client),
		Event:       event,
		Properties:  properties,
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
	properties.Set("version", build.Version)
	track := analytics.Track{
		AnonymousId: MachineID(),
		UserId:      MachineID(),
		Event:       "kusk-cli",
		Properties:  properties,
	}
	return sendDataToGA(track)
}

// SendAnonymousCommandError will send CLI event to GA
func SendAnonymousCommandError(command string, err error, miscInfo map[string]interface{}) {
	properties := analytics.NewProperties()
	properties.Set("event", "command")
	properties.Set("command", command)
	properties.Set("error", err)
	properties.Set("version", build.Version)

	for key, value := range miscInfo {
		// Make sure we are not overwriting any previously set property
		if _, ok := properties[key]; !ok {
			properties.Set(key, value)
		}
	}

	track := analytics.Track{
		AnonymousId: MachineID(),
		UserId:      MachineID(),
		Event:       "kusk-cli",
		Properties:  properties,
	}

	if sendError := sendDataToGA(track); sendError != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("analytics: failed to report error - command=%v, err=%v, miscInfo=%v, %w", command, err, miscInfo, sendError))
	}
}

func sendDataToGA(track analytics.Track) error {
	// enabled by default, dont send anything if explicitely disabled
	if enabled, ok := os.LookupEnv("ANALYTICS_ENABLED"); ok && enabled == "false" {
		return nil
	}

	client := analytics.New(TelemetryToken)
	if err := client.Enqueue(track); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("analytics: failed to enqueue track=%v, %w", track, err))
		return err
	}
	if err := client.Close(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("analytics: failed to close client, %w", err))
		return err
	}

	return nil
}

func ClusterID(ctx context.Context, client client.Client) string {
	ns := &corev1.Namespace{}
	if err := client.Get(ctx, types.NamespacedName{Name: "kube-system"}, ns); err != nil {
		return MachineID()
	}
	return string(ns.UID)
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
