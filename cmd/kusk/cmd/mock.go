/*
The MIT License (MIT)

Copyright © 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kubeshop/testkube/pkg/ui"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/config"
	error_reporter "github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/mocking"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/mocking/filewatcher"

	mockingServer "github.com/kubeshop/kusk-gateway/cmd/kusk/internal/mocking/server"
	"github.com/kubeshop/kusk-gateway/pkg/spec"
)

var mockServerPort uint32

// mockCmd represents the mock command
var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Spin up a local mocking server serving your API",
	Long: `Spin up a local mocking server that generates responses from your content schema or returns your defined examples.
Schema example:

content:
 application/json:
  schema:
   type: object
   properties:
    title:
     type: string
     description: Description of what to do
    completed:
     type: boolean
    order:
     type: integer
     format: int32
    url:
     type: string
     format: uri
   required:
    - title
    - completed
    - order
    - url

The mock server will return a response like the following that matches the schema above:
{
 "completed": false,
 "order": 1957493166,
 "title": "Inventore ut.",
 "url": "http://langosh.name/andreanne.parker"
}

Example with example responses:

application/xml:
 example:
  title: "Mocked XML title"
  completed: true
  order: 13
  url: "http://mockedURL.com"

The mock server will return this exact response as its specified in an example:
<doc>
 <completed>true</completed>
 <order>13</order>
 <title>Mocked XML title</title>
 <url>http://mockedURL.com</url>
</doc>
`,
	Example: `
To mock an api on the local file system
$ kusk mock -i path-to-openapi-file.yaml

To mock an api from a url
$ kusk mock -i https://url.to.api.com
`,
	Run: func(cmd *cobra.Command, args []string) {
		reportError := func(err error) {
			if err != nil {
				error_reporter.NewErrorReporter(cmd, err).Report()
			}
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			err := fmt.Errorf("unable to fetch user's home directory: %w", err)
			reportError(err)
			ui.Fail(err)
		}

		if err := config.CreateDirectoryIfNotExists(homeDir); err != nil {
			reportError(err)
			ui.Fail(err)
		}

		kuskConfigDir := path.Join(homeDir, ".kusk")

		apiParser := spec.NewParser(&openapi3.Loader{
			IsExternalRefsAllowed: true,
			ReadFromURIFunc:       openapi3.ReadFromURIs(openapi3.ReadFromHTTP(http.DefaultClient), openapi3.ReadFromFile),
		})
		apiSpec, err := apiParser.Parse(apiSpecPath)
		if err != nil {
			err := fmt.Errorf("error when parsing openapi spec: %w", err)
			reportError(err)
			ui.Fail(err)
		}

		if err := apiSpec.Validate(context.Background()); err != nil {
			err := fmt.Errorf("openapi spec failed validation: %w", err)
			reportError(err)
			ui.Fail(err)
		}

		ui.Info(ui.Green("🎉 successfully parsed OpenAPI spec"))

		u, err := url.Parse(apiSpecPath)
		if err != nil {
			reportError(err)
			ui.Fail(err)
		}

		var watcher *filewatcher.FileWatcher
		var tempApiFileName string
		apiSpecLocation := apiSpecPath
		apiToMock := apiSpecPath
		if apiOnFileSystem := u.Host == ""; apiOnFileSystem {
			var currentWorkingDir string
			apiSpecDir := filepath.Dir(apiSpecPath)

			if apiSpecDir != "" {
				if currentWorkingDir, err = os.Getwd(); err != nil {
					ui.Fail(err)
				}
				if err := os.Chdir(apiSpecDir); err != nil {
					reportError(err)
					ui.Fail(err)
				}
				defer func() {
					if err := os.Chdir(currentWorkingDir); err != nil {
						ui.Fail(err)
					}
				}()
			}

			tempApiFile, err := os.CreateTemp(kuskConfigDir, "mocked-api-*.yaml")
			if err != nil {
				ui.Fail(err)
			}

			tempApiFileName = tempApiFile.Name()

			defer func(fileName string) {
				if err := tempApiFile.Close(); err != nil {
					ui.Fail(err)
				}
				if err := os.Remove(fileName); err != nil {
					ui.Fail(err)
				}
			}(tempApiFileName)

			if err := writeInitialisedApiToTempFile(tempApiFileName, apiSpec); err != nil {
				ui.Fail(err)
			}
			// we need the absolute path of the file in the filesystem
			// to properly mount the file into the mocking container
			absoluteApiSpecPath, err := filepath.Abs(apiSpecPath)
			if err != nil {
				reportError(err)
				ui.Fail(err)
			}

			watcher, err = filewatcher.New(absoluteApiSpecPath)
			if err != nil {
				reportError(err)
				ui.Fail(err)
			}
			defer watcher.Close()

			apiSpecLocation = absoluteApiSpecPath
			apiToMock = tempApiFileName
		}

		ui.Info(ui.White("☀️ initializing mocking server"))
		mockServer, err := setUpMockingServer(kuskConfigDir, apiToMock)
		if err != nil {
			msg := fmt.Errorf("error when setting up mocking server: %w", err)
			reportError(msg)
			ui.Fail(msg)
		}
		ctx := context.Background()
		mockServerId, err := mockServer.Start(ctx)
		if err != nil {
			reportError(err)
			ui.Fail(err)
		}

		statusCh, errCh := mockServer.ServerWait(ctx, mockServerId)

		go mockServer.StreamLogs(ctx, mockServerId)

		ui.Info(ui.Green("🎉 server successfully initialized"))
		ui.Info(ui.DarkGray("URL: ") + ui.White("http://localhost:"+fmt.Sprint(mockServerPort)))

		// set up signal channel listening for ctrl+c
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		// if watcher is nil, then the api comes from a URL and we shouldn't watch it
		// otherwise it's on the file system and we can watch for changes
		if watcher != nil {
			ui.Info(ui.White("⏳ watching for file changes in " + apiSpecPath))
			go watcher.Watch(func() {
				ui.Info("✍️ change detected in " + apiSpecPath)
				err := apiFileUpdateHandler(ctx, mockServer, apiSpecLocation, tempApiFileName, mockServerId)
				if err != nil {
					reportError(err)
					ui.Fail(err)
				}
			}, sigs)
		}

		for {
			select {
			case status, ok := <-statusCh:
				if !ok {
					return
				}
				if status.Error == nil && status.StatusCode > 0 {
					mockServerId, err = mockServer.Start(ctx)
					if err != nil {
						err := fmt.Errorf("unable to restart mocking server")
						reportError(err)
						ui.Fail(err)
					}
					ui.Info("☀️ mock server restarted")

					// reassign status and err channels for new mock server
					// as old ones will now be closed
					statusCh, errCh = mockServer.ServerWait(ctx, mockServerId)
					// restarting the container will kill the log stream
					// so start it up again
					go mockServer.StreamLogs(ctx, mockServerId)
				}
			case err, ok := <-errCh:
				if !ok {
					return
				}
				err = fmt.Errorf("an unexpected error occured: %w", err)
				reportError(err)
				ui.Fail(err)
			case logEntry, ok := <-mockServer.LogCh:
				if !ok {
					return
				}
				ui.Info(decorateLogEntry(logEntry))
			case err, ok := <-mockServer.ErrCh:
				if !ok {
					return
				}
				ui.Warn(err.Error())
			case <-sigs:
				ui.Info("😴 shutting down mocking server")
				if err := mockServer.Stop(ctx, mockServerId); err != nil {
					err := fmt.Errorf("unable to stop mocking server: %w", err)
					reportError(err)
					ui.Fail(err)
				}
				return
			}
		}
	},
}

func setUpMockingServer(kuskConfigDir, apiToMock string) (*mockingServer.MockServer, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("unable to create new docker client from environment: %w", err)
	}

	if mockServerPort == 0 {
		mockServerPort, err = scanForNextAvailablePort(8080)
		if err != nil {
			return nil, fmt.Errorf("unable to find available port for mocking server: %w", err)
		}
	}

	mockingConfigFilePath := path.Join(kuskConfigDir, "openapi-mock.yaml")
	if err := writeMockingConfigIfNotExists(mockingConfigFilePath); err != nil {
		return nil, fmt.Errorf("unable to write mocking config file: %w", err)
	}

	return mockingServer.New(cli, mockingConfigFilePath, apiToMock, mockServerPort), nil
}

func writeInitialisedApiToTempFile(fileName string, api *openapi3.T) error {
	api.InternalizeRefs(context.Background(), nil)
	apiBytes, err := api.MarshalJSON()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if _, err = file.Write(apiBytes); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func apiFileUpdateHandler(
	ctx context.Context,
	mockServer *mockingServer.MockServer,
	apiFileName, tempApiFileName, mockServerId string,
) error {
	apiSpec, err := spec.NewParser(&openapi3.Loader{
		IsExternalRefsAllowed: true,
		ReadFromURIFunc:       openapi3.ReadFromFile,
	}).Parse(apiFileName)
	if err != nil {
		return fmt.Errorf("unable to parse api spec: %w", err)
	}
	if err := writeInitialisedApiToTempFile(tempApiFileName, apiSpec); err != nil {
		return fmt.Errorf("unable to write api spec to temp file: %w", err)
	}
	if err := mockServer.Stop(ctx, mockServerId); err != nil {
		return fmt.Errorf("unable to update mocking server: %w", err)
	}

	return nil
}

func scanForNextAvailablePort(startingPort uint32) (uint32, error) {
	localPortCheck := func(port uint32) error {
		ln, err := net.Listen("tcp", "127.0.0.1:"+fmt.Sprint(port))
		if err != nil {
			return err
		}

		ln.Close()
		return nil
	}

	const maxPortNumber = 65535

	for port := startingPort; port <= maxPortNumber; port++ {
		if localPortCheck(port) == nil {
			return port, nil
		}
	}

	return 0, fmt.Errorf("unable to find available port between %d-%d", startingPort, maxPortNumber)
}

func writeMockingConfigIfNotExists(mockingConfigPath string) error {
	_, err := os.Stat(mockingConfigPath)
	if err == nil {
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("unable to check for mocking config: %w", err)
	}

	f, err := os.Create(mockingConfigPath)
	if err != nil {
		return fmt.Errorf("unable to create mocking config file at %s: %w", mockingConfigPath, err)
	}
	defer f.Close()
	if err := mocking.WriteMockingConfig(f); err != nil {
		return fmt.Errorf("unable to write mocking config to %s: %w", mockingConfigPath, err)
	}

	return nil

}

func decorateLogEntry(entry mockingServer.AccessLogEntry) string {
	methodColors := map[string]func(...interface{}) string{
		http.MethodGet:     ui.Blue,
		http.MethodPost:    ui.Green,
		http.MethodDelete:  ui.LightRed,
		http.MethodHead:    ui.LightBlue,
		http.MethodPut:     ui.Yellow,
		http.MethodPatch:   ui.Red,
		http.MethodConnect: ui.LightCyan,
		http.MethodOptions: ui.LightYellow,
		http.MethodTrace:   ui.White,
	}

	decoratedStatusCode := ui.Green(entry.StatusCode)

	if intStatusCode, err := strconv.Atoi(entry.StatusCode); err == nil && intStatusCode > 399 {
		decoratedStatusCode = ui.Red(entry.StatusCode)
	}

	return fmt.Sprintf(
		"%s %s %s %s",
		ui.DarkGray(entry.TimeStamp),
		methodColors[entry.Method]("[", entry.Method, "]"),
		decoratedStatusCode,
		ui.White(entry.Path),
	)

}

func init() {
	rootCmd.AddCommand(mockCmd)
	mockCmd.Flags().StringVarP(&apiSpecPath, "in", "i", "", "path to openapi spec you wish to mock")
	mockCmd.MarkFlagRequired("in")

	mockCmd.Flags().Uint32VarP(&mockServerPort, "port", "p", 0, "port to expose mock server on. If none specified, will search for next available port starting from 8080")
}
