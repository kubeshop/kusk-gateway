# How to develop Kusk Gateway

Kusk Gateway code is managed with the help of [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) that provides code scaffolding and generation of K8s Custom Resource Definitions files.

Internally Kusk Gateway uses the [go-control-plane](https://github.com/envoyproxy/go-control-plane) package to configure Envoy with its xDS protocol.

We use [Minikube](https://minikube.sigs.k8s.io/docs/start/) as development environment, so the following instructions and the Makefile in the project are tuned to it.

Make sure you have Minikube installed before proceeding further.

For the MacOS users, the additional configuration step is needed to setup and set as the default for Minikube driver [hyperkit](https://minikube.sigs.k8s.io/docs/drivers/hyperkit/).

## Set up development environment

### with in-cluster debugging

- Set up remote debugging for your IDE pointed at localhost:40000
  - [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#attach-to-a-process-in-the-docker-container)
  - [VSCode](https://github.com/golang/vscode-go/blob/master/docs/debugging.md#configure) (see below for a working example)
- Run `make create-env`. Once this command finishes you should have the working environment with kusk-gateway-manager running in kusk-system namespace.
- To attach the IDE to the pod for debugging run `make update-debug` that will build the debug image inside the Minikube cluster and will update the kusk-gateway-manager deployment.
  After the deployment kusk-gateway-manager pod will be alive but not running the application since Delve in the container will wait for you to connect to it on port 4000.
  Run `kubectl port-forward deployment/kusk-gateway-manager -n kusk-system 40000:40000` in a new terminal window to create the port-forwarding to Delve port. It is advised to make this as a kind of Task to run from IDE.
- Run your debug configuration from IDE to connect to port-forwarded localhost port :40000.
- You can now deploy the httpbin example that creates a backend API service and pushes gateway CRDs to configure Envoy with `kubectl apply -f examples/httpbin`.
- Place breakpoints in the code and debug as normal.

To test changes to the code, run the following:

- `make generate manifests install docker-build deploy cycle` for the usual builds (without the debugging)
- `make update` for the usual builds if only the manager code was changed and no CRDs update is needed.
- `make generate manifests install docker-build-debug deploy-debug cycle` for debug build.
- `make update-debug` for the debug builds if only the manager code was changed and no CRDs update is needed.

#### VSCode launch.json example

```yaml
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Connect to server",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "remotePath": "${workspaceFolder}",
            "port": 40000,
            "host": "127.0.0.1"
        }
    ]
}
```

### Run kusk gateway locally

- Run `make create-env`
- Run `kubectl apply -f ./config/samples/gateway_v1_envoyfleet.yaml -n kusk-system`
- Run `make run`
  - This runs the built binary on your machine and creates a tunnel to minikube so envoy and Kusk Gateway running in your IDE can communicate.
