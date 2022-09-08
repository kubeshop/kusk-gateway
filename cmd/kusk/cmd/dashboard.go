package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/kubeshop/testkube/pkg/process"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/k8s"
)

var (
	dashboardEnvoyFleetName         string
	dashboardEnvoyFleetNamespace    string
	dashboardEnvoyFleetExternalPort int
)

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.Flags().StringVarP(&dashboardEnvoyFleetNamespace, "envoyfleet.namespace", "", "kusk-system", "kusk gateway dashboard envoy fleet namespace")
	dashboardCmd.Flags().StringVarP(&dashboardEnvoyFleetName, "envoyfleet.name", "", "kusk-gateway-private-envoy-fleet", "kusk gateway dashboard envoy fleet service name")
	dashboardCmd.Flags().IntVarP(&dashboardEnvoyFleetExternalPort, "external-port", "", 8080, "external port to access dashboard at")
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Access the kusk dashboard",
	Long: `Access the kusk dashboard. kusk dashboard will start a port-forward session on port 8080 to the envoyfleet
serving the dashboard and will open the dashboard in the browser. By default this is kusk-gateway-private-envoy-fleet.kusk-system.
The flags --envoyfleet.namespace and --envoyfleet.name can be used to change the envoyfleet.
	`,
	Example: `
	$ kusk dashboard

	Opens the kusk gateway dashboard in the browser by exposing the default private envoy fleet on port 8080

	$ kusk dashboard --envoyfleet.namespace=other-namespace --envoyfleet.name=other-envoy-fleet

	Specify other envoyfleet and namespace that is serving the dashboard

	$ kusk dashboard --external-port=9090

	Expose dashboard on port 9090
	`,
	Run: func(cmd *cobra.Command, args []string) {
		reportError := func(err error) {
			if err != nil {
				// Report error
				miscInfo := map[string]interface{}{
					"envoyfleet.namespace": dashboardEnvoyFleetNamespace,
					"envoyfleet.name":      dashboardEnvoyFleetName,
					"external-port":        dashboardEnvoyFleetExternalPort,
					"args":                 args,
					"os.Args":              os.Args,
					"config":               cfgFile,
					"env":                  os.Environ(),
				}
				errors.NewErrorReporter(cmd, err, miscInfo).Report()
			}
		}

		kubeConfig, err := k8s.GetKubeConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get kube config: %v\n", err)
			reportError(err)
			os.Exit(1)
		}

		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			reportError(err)
			os.Exit(1)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			reportError(err)
			os.Exit(1)
		}

		podList, err := clientset.CoreV1().Pods(dashboardEnvoyFleetNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("fleet=%s.%s", dashboardEnvoyFleetName, dashboardEnvoyFleetNamespace),
		})
		if err != nil || len(podList.Items) == 0 {
			fmt.Fprintln(os.Stderr, err)
			reportError(err)
			os.Exit(1)
		}

		var chosenPod *v1.Pod
		for _, pod := range podList.Items {
			// pick the first pod found to be running
			if pod.Status.Phase == v1.PodRunning {
				chosenPod = &pod
			}
		}

		if chosenPod == nil {
			fmt.Fprintln(os.Stderr, "no running pods found for envoyfleet: ", dashboardEnvoyFleetName)
			reportError(err)
			os.Exit(1)
		}

		// stopCh controls the port forwarding lifecycle.
		// When it gets closed the port forward will terminate
		stopCh := make(chan struct{}, 1)
		// readyCh communicates when the port forward is ready to receive traffic
		readyCh := make(chan struct{})

		// managing termination signal from the terminal.
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			<-sigs
			fmt.Println("Exiting...")
			close(stopCh)
			wg.Done()
		}()

		go func() {
			err := k8s.PortForward(k8s.PortForwardRequest{
				RestConfig: config,
				Pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      chosenPod.Name,
						Namespace: chosenPod.Namespace,
					},
				},
				ExternalPort: dashboardEnvoyFleetExternalPort,
				InternalPort: 8080,
				StopCh:       stopCh,
				ReadyCh:      readyCh,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				reportError(err)
				os.Exit(1)
			}
		}()

		<-readyCh

		browserOpenCMD, browserOpenArgs := getBrowserOpenCmdAndArgs("http://localhost:8080")
		process.Execute(browserOpenCMD, browserOpenArgs...)
		wg.Wait()
	},
}

// open opens the specified URL in the default browser of the user.
func getBrowserOpenCmdAndArgs(url string) (string, []string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	return cmd, args
}
