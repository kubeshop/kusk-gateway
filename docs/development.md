# How to develop Kusk Gateway

Kusk Gateway code is managed with the help of [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) that provides code scaffolding and generation of K8s Custom Resource Definitions files.

Internally Kusk Gateway uses the [go-control-plane](https://github.com/envoyproxy/go-control-plane) package to configure Envoy with its xDS protocol.

## Set up development environment
### with in-cluster debugging
- Set up remote debugging for your IDE pointed at localhost:40000 
  - [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#attach-to-a-process-in-the-docker-container)
  - [VSCode](https://github.com/golang/vscode-go/blob/master/docs/debugging.md#configure) (see below for a working example)
- Run `make create-env`
- When the make script is waiting for kusk-gateway-manager to become healthy, run `kubectl port-forward deployment/kusk-gateway-manager -n kusk-system 40000:40000` in a new terminal window
- Run your debug configuration from your IDE. The pod won't become healthy until you do this as Delve waits for a connection on :40000.
- When the script completes, you can now deploy the httpbin example that creates a backend API service and pushes gateway CRDs to configure Envoy with `kubectl apply -f examples/httpbin`.
- Place breakpoints in the code and debug as normal

To test changes to the code, run the following:
- `make generate manifests install docker-build`
	- If your running the code in minikube, don't forget to `eval $(minikube docker-env [--profile "$PROFILE_NAME"])`
	- e.g. `eval $(minikube docker-env --profile "kgw")` if you ran `make create-env`
- restart kusk-gateway deployment to pick up the new image - `kubectl rollout restart deployment/kusk-gateway-manager -n kusk-system`

#### VSCode launch.json example
```
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
