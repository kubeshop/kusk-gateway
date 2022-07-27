package k8s

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config

	// Pod is the selected pod for this port forwarding
	Pod v1.Pod

	// ExternalPort the local port that will be selected to expose the InternalPort
	ExternalPort int
	// InternalPort is the target port for forwarding
	InternalPort int

	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}

func PortForward(req PortForwardRequest) error {
	targetURL, err := url.Parse(req.RestConfig.Host)
	if err != nil {
		return err
	}

	targetURL.Path = path.Join(
		"api", "v1",
		"namespaces", req.Pod.Namespace,
		"pods", req.Pod.Name,
		"portforward",
	)

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		http.MethodPost,
		targetURL,
	)
	fw, err := newPortForwarder(dialer, req)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func newPortForwarder(dialer httpstream.Dialer, req PortForwardRequest) (*portforward.PortForwarder, error) {
	// stream is used to tell the port forwarder where to place its output or
	// where to expect input if needed. For the port forwarding we just need
	// the output eventually
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	return portforward.New(dialer, portMapping(req.ExternalPort, req.InternalPort), req.StopCh, req.ReadyCh, stream.Out, stream.ErrOut)
}

func portMapping(externalPort, internalPort int) []string {
	return []string{fmt.Sprintf("%d:%d", externalPort, internalPort)}
}
