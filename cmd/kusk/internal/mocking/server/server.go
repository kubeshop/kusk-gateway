package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type MockServer struct {
	client     *client.Client
	image      string
	configFile string
	apiToMock  string
	port       uint32

	LogCh chan AccessLogEntry
	ErrCh chan error
}

type AccessLogEntry struct {
	TimeStamp  string
	Method     string
	Path       string
	StatusCode string

	Error error
}

func New(ctx context.Context, client *client.Client, configFile, apiToMock string, port uint32) (MockServer, error) {
	const openApiMockImage = "muonsoft/openapi-mock:v0.3.1"

	reader, err := client.ImagePull(ctx, openApiMockImage, types.ImagePullOptions{})
	if err != nil {
		return MockServer{}, fmt.Errorf("unable to pull mock server image: %w", err)
	}

	// wait for download to complete, discard output
	defer reader.Close()
	io.Copy(io.Discard, reader)

	return MockServer{
		client:     client,
		image:      openApiMockImage,
		configFile: configFile,
		apiToMock:  apiToMock,
		port:       port,
		LogCh:      make(chan AccessLogEntry),
		ErrCh:      make(chan error),
	}, nil
}

func (m MockServer) Start(ctx context.Context) (string, error) {
	u, err := url.Parse(m.apiToMock)
	if err != nil {
		return "", err
	}

	containerMockingConfigFilePath := "/app/mocking/openapi-mock.yaml"
	binds := []string{
		m.configFile + ":" + containerMockingConfigFilePath,
	}

	containerApiSpecPath := "mocking/fake-api.yaml"
	if u.Host != "" {
		containerApiSpecPath = m.apiToMock // serve from url
	} else {
		// serving from local file so mount it into container
		binds = append(binds, m.apiToMock+":/app/"+containerApiSpecPath)
	}

	resp, err := m.client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        m.image,
			ExposedPorts: nat.PortSet{"8080": struct{}{}},
			Tty:          true,
			AttachStdout: true,
			AttachStderr: true,
			Env: []string{
				"OPENAPI_MOCK_SPECIFICATION_URL=" + containerApiSpecPath,
			},
			Cmd: strslice.StrSlice{
				"serve",
				"--configuration",
				containerMockingConfigFilePath,
			},
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds:      binds,
			PortBindings: map[nat.Port][]nat.PortBinding{
				nat.Port("8080"): {
					{
						HostIP: "127.0.0.1", HostPort: fmt.Sprint(m.port),
					},
				},
			},
		},
		nil,
		nil,
		"",
	)

	if err != nil {
		return "", fmt.Errorf("unable to create mocking server: %w", err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("unable to start mocking server: %w", err)
	}

	return resp.ID, nil
}

func (m MockServer) Restart(ctx context.Context, MockServerId string) error {
	timeout := 5 * time.Second
	return m.client.ContainerRestart(ctx, MockServerId, &timeout)
}

func (m MockServer) Stop(ctx context.Context, MockServerId string) error {
	timeout := 5 * time.Second
	return m.client.ContainerStop(ctx, MockServerId, &timeout)
}

func (m MockServer) ServerWait(ctx context.Context, MockServerId string) (<-chan container.ContainerWaitOKBody, <-chan error) {
	return m.client.ContainerWait(ctx, MockServerId, container.WaitConditionNextExit)
}

func (m MockServer) StreamLogs(ctx context.Context, containerId string) {
	reader, err := m.client.ContainerLogs(ctx, containerId, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		m.ErrCh <- err
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if le, err := newAccessLogEntry(scanner.Text()); err != nil {
			m.ErrCh <- err
		} else {
			m.LogCh <- le
		}
	}
}

func newAccessLogEntry(rawLog string) (AccessLogEntry, error) {
	if strings.Contains(rawLog, "warning") || strings.Contains(rawLog, "error") {
		return AccessLogEntry{}, errors.New(rawLog)
	}

	logLine := strings.Split(rawLog, " ")

	timeStamp := strings.TrimPrefix(logLine[3], "[")
	method := strings.TrimPrefix(logLine[5], "\"")
	path := logLine[6]
	statusCode := logLine[8]

	return AccessLogEntry{
		TimeStamp:  timeStamp,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
	}, nil
}
