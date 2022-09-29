# How to build kusk cli

## How it works 

Kusk CLI uses embedded manifests for `kusk cluster install`. Embedded files are built using [go-bindata](https://github.com/go-bindata/go-bindata) 

`go-bindata` uses directive to build file that contains all binary data from manifests
```
//go:generate go-bindata -prefix "../../" -o cmd/manifest_data.go -pkg=cmd -ignore=debug/ -ignore=local/ -ignore=prometheus/ -ignore=samples/ ../../config/... manifests/...
```

First set of manifests are taken from `config` directory in ther root directory of the project. These manifests are used by kustomize to build and deploy kusk CRDs, Controller, RBAC and Webhooks. Configuration `config` are generated each time when `make manifests` is executed in the root director of the project. `make manifests` ensures that directory `config` always has the latest changes on the controller.  

Another set of manifests being embedded are located in `cmd/kusk/manifests` directory. These hold standard deployment for Kusk Dashboard and API server. These are split as follows:
1. fleets.yaml - holds manifest for the public envoyfleet we need for hosting the dashboard
2. api_server_api.yaml - holds manifest for API Server API
3. api_server.yaml - holds manifest for the API Server service, deployment, ServiceAccount, ClusterRoleBinding and ClusterRole
3. dashboard_envoyfleet.yaml - holds manifest of the private envoyfleet for Dashboard so it can access API Server
4. dashboard_staticroute.yaml - holds manifest for the Static Route for hosting the Dashboard
5. dashboard.yaml - holds manifests for the dashboard Dervice, Deployment and ServiceAccount


When invoking `kusk cluster install` embedded manifests are unpacked and stored in temporary directory, and are removed after use. 
1. Integrated `kubectl apply -k config/default` is invoked to install all Kubernetes resources required for Kusk to run
2. Installer waits for kusk-gateway-manager pod to become Ready
3. Integrated `kubectl apply -f manifests/fleets.yaml` (see lines 131 to 165 in cmd/kusk/cmd/install.go) which are installing Kusk API Server and Kusk Dashboard resources. 
4. After each step installer waits for Envoyfleet deployment to become ready before moving to installing Kusk API Server on which it also waits. And finally it installs Kusk Dashboard.  
   
`kusk cluster install --latest` - This command will pull the latest manifests from the latest [release](https://github.com/kubeshop/kusk-gateway/releases) in Kusk repository. After downloading the release installer will unzip it and place files in a temporary directory and then execute steps as shown above.


## How to build 

To build kusk CLI Makefile is used for example `VERSION=v1.2.5 make build`

```
build: kustomize
```
Installs kustomize binary 

```
	cd ../../config/default && $(KUSTOMIZE) edit set image kusk-gateway=${MANAGER_IMG}
```
Sets default container image using kustomize. `kustomize edit set image kusk-gateway=${MANAGER_IMG}` will set value in `config/default/kustomization.yaml` to construct manifests with default image as below
```
images:
- name: kusk-gateway
  newName: kubeshop/kusk-gateway
  newTag: v1.2.5
```

```
	go generate 
	go build -v -o ./kusk -ldflags="${LD_FLAGS}" ./main.go
```
Generates embedded manifests and build the binary