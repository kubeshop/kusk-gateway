# How to develop Kusk Gateway

Kusk Gateway is built on top of Kubebuilder which uses custom-written managers that react to creation of new APIs and updates the envoy configuration using the xDS protocol.

## Set up development environment
### with in-cluster debugging
- Set up remote debugging for your IDE pointed at localhost:40000 
  - [Goland](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html#attach-to-a-process-in-the-docker-container)
  - [VSCode](https://github.com/golang/vscode-go/blob/master/docs/debugging.md#configure)
- Run `make create-env`
- When the make script is waiting for kusk-controller-manager to become healthy, run `kubectl port-forward deployment/kusk-controller-manager -n kusk-system 40000:40000` in a new terminal window
- Run your debug configuration from your IDE. The pod won't become healthy until you do this as Delve waits for a connection on :40000.
- When the script completes, you can now deploy httpbin with `kubectl apply -f examples/httpbin`
- Place breakpoints in the code and debug as normal

### Run kusk gateway locally
- Run `make create-env`
- Run `kubectl apply -f ./config/samples/gateway_v1_envoyfleet.yaml -n kusk-system`
- Run `make run` 
  - This runs the built binary on your machine and creates a tunnel to minikube so envoy and kusk gateway can communicate.
