<!-- Add banner here -->

# Kusk Gateway

<!-- Add buttons here -->

Kusk Gateway is a self-service API gateway powered by [OpenAPI](https://www.openapis.org/) and [Envoy](https://www.envoyproxy.io/)!

Kusk Gateway is for you if:
- You or your team develop REST APIs running in Kubernetes
- Embrace a contract-first approach to developing your APIS using OpenAPI or Swagger
- You don't want to spend lots of time configuring hard-to-work-with ingress controllers that require a dedicated Ops Engineer

Kusk Gateway configures itself through the endpoints defined in your OpenAPI or Swagger document removing the need to DevOps intervention. 

You can apply your API definition like any other Kubernetes resource using our custom-made Kusk API CustomResourceDefinition.

##  Why did we create Kusk Gateway?
See our [announcement blog post](...) for full details!

But in short, we wanted to achieve the following:
- Hit the ground running when deploying a new REST api to Kubernetes

- Simplified configuration for development workflows when deployed locally

- Provides observability for deployed endpoints without having to learn multiple tools (Prometheus, Grafana, etc) out of the box

- Decoupling: let the Ops configure cluster-wide settings and let Developers deploy the APIs in self-service manner

- Automated QoS: 
    - schema-based request validation
    - rate-limiting
    - timeouts

### How are things normally done today?
By manually creating and updating API-related resources, such as Ingress. The format of which will depend on your Gateway. This is not something develpers normally have control over.

Installing and maintaining several observability tools, such as Prometheus and Grafana and implementing the observability functionality in the code

This means developers have to learn how to use these tools.

### How does Kusk Gateway solve these issues?
- Removes requirement for developers to know multiple tools just to deploy their APIs. If you know how to write an OpenAPI document, you know Kusk Gateway!

- Maintainability/safety: API related settings (rate limiting, security) are held in one place being the (reviewable!) source of truth, removing the manual configuration step, lowering the chance of human error

- Flexibility: services can be created in any language and have same level of observability / security features without having to implement it on your own

- No vendor lock: kusk-gateway can be installed on any cloud or bare metal Kubernetes cluster

# Table of contents
- [Installation](#installation)
- [Usage](#usage)
- [Development](#development)
- [Contribute](#contribute)
- [License](#license)

# Installation
[(Back to top)](#table-of-contents)

TODO(#63) - Add Helm installation instructions.

## Local Installation for Evaluation
If you want to run kusk-gateway locally, you can do this easily using Minikube

Prerequisites - make sure you have the following installed and on your PATH:
- `jq`
- `kubectl`
- `docker`
- `minikube`

Run:
- `make create-cluster` # creates and configures the minikube cluster
- `make install` # install the required CRDs
- `eval $(minikube docker-env --profile "kgw")` # so built docker images are available to Minikube
- `make docker-build deploy` # build and deploy the kusk gateway image
- `kubectl rollout status -w deployment/kusk-controller-manager -n kusk-system`

Once Kusk Gateway is installed and running, you can try and apply your own OpenAPI specs, see Usage below or you can apply one of our examples

### Example
```
kubectl apply -f examples/httpbin && kubectl rollout status -w deployment/httpbin

external_ip=$(kubectl -n kusk-system get svc kusk-envoy --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
curl -v http://$external_ip:8080/get

```

# Usage
[(Back to top)](#table-of-contents)

## API CRD Format
```
apiVersion: gateway.kusk.io/v1
kind: API
metadata:
  name: httpbin-sample
spec:
  # service name and port should be specified inside x-kusk annotation
  spec: |
    $YOUR_API_DOCUMENT_HERE
```

See [httpbin API Resource](examples/httpbin/httpbin_v1_api.yaml) for a full example


TODO(#65) - provide link to fill x-kusk documentation

# Development
[(Back to top)](#table-of-contents)

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
- Run `make run` 
  - This runs the built binary on your machine and creates a tunnel to minikube so envoy and kusk gateway can communicate.

# Contribute
[(Back to top)](#table-of-contents)

- Check out our [Contributor Guide](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md) and
  [Code of Conduct](https://github.com/kubeshop/.github/blob/main/CODE_OF_CONDUCT.md)
- Fork/Clone the repo and make sure you can run it as shown above
- Check out open [issues](https://github.com/kubeshop/kusk-gateway/issues) here on GitHub
- Get in touch with the team by starting a discussion on [GitHub](https://github.com/kubeshop/kusk-gateway/discussions) or on our [Discord Server](https://discord.gg/uNuhy6GDyn).
- or open an issue of your own that you would like to contribute to the project.

# License
[(Back to top)](#table-of-contents)

[MIT](https://mit-license.org/)

