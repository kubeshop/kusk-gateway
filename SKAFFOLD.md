# `SKAFFOLD`

See [`https://skaffold.dev`](https://skaffold.dev).

Picks up changes in files (`*.go`) and re-deploys image to cluster _automatically_.

**NB:**

* Use `./skaffold.sh run` to start.
* Use `skaffold delete` to do teardown.
* `minikube` is required.
* [dlv](https://github.com/go-delve/delve) for debugging.
* and finally `kustomize`.

Install `skaffold`
-----------------

```sh
$ ARCH="$([ $(uname -m) = "aarch64" ] && echo "arm64" || echo "amd64")"
curl -L --output /usr/local/bin/skaffold "https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-${ARCH}"
sudo chmod +x /usr/local/bin/skaffold
$ skaffold version
v1.39.2
```

Bring up cluster
----------------

```sh
$ minikube start --profile kgw --addons=metallb
$ ./skaffold.sh run
Generating tags...
 - kubeshop/kusk-gateway -> kubeshop/kusk-gateway:latest
Checking cache...
 - kubeshop/kusk-gateway: Not found. Building
Starting build...
Found [kgw] context, using local docker daemon.
Building [kubeshop/kusk-gateway]...
Target platforms: [linux/amd64]
[+] Building 39.5s (18/18) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                                           0.0s
 => => transferring dockerfile: 1.33kB                                                                                                                                         0.0s
 => [internal] load .dockerignore                                                                                                                                              0.0s
 => => transferring context: 211B                                                                                                                                              0.0s
 => [internal] load metadata for gcr.io/distroless/static:nonroot                                                                                                              1.5s
 => [internal] load metadata for docker.io/library/golang:1.19-alpine                                                                                                          1.8s
 => [auth] library/golang:pull token for registry-1.docker.io                                                                                                                  0.0s
 => [builder 1/8] FROM docker.io/library/golang:1.19-alpine@sha256:f3e683657ddf73726b5717c2ff80cdcd9e9efb7d81f77e4948fada9a10dc7257                                            6.4s
 => => resolve docker.io/library/golang:1.19-alpine@sha256:f3e683657ddf73726b5717c2ff80cdcd9e9efb7d81f77e4948fada9a10dc7257                                                    0.0s
 => => sha256:46752c2ee3bd8388608e41362964c84f7a6dffe99d86faeddc82d917740c5968 1.36kB / 1.36kB                                                                                 0.0s
 => => sha256:f9a40cb7e8ec6730fbc2feaad9b26b429930160c681b6d1e58ad3df1ad72d6f5 5.35kB / 5.35kB                                                                                 0.0s
 => => sha256:213ec9aee27d8be045c6a92b7eac22c9a64b44558193775a1a7f626352392b49 2.81MB / 2.81MB                                                                                 0.4s
 => => sha256:4583459ba0371c715f926a9bbd37a9dae909234f4b898220160425131eb53bd4 284.73kB / 284.73kB                                                                             0.4s
 => => sha256:93c1e223e6f2123b855e0c95898eba50cb6a055881ba9023527c0a361761c1cf 153B / 153B                                                                                     0.4s
 => => sha256:f3e683657ddf73726b5717c2ff80cdcd9e9efb7d81f77e4948fada9a10dc7257 1.65kB / 1.65kB                                                                                 0.0s
 => => extracting sha256:213ec9aee27d8be045c6a92b7eac22c9a64b44558193775a1a7f626352392b49                                                                                      0.0s
 => => extracting sha256:4583459ba0371c715f926a9bbd37a9dae909234f4b898220160425131eb53bd4                                                                                      0.0s
 => => sha256:14bccb3a1cff3b5f09bf1f21ba9ecb17c96329fb82f96605d922cfa4c9a82032 122.25MB / 122.25MB                                                                             3.7s
 => => sha256:861cc3991b26fba6073631f27a1dc32bb27761195d912d05ee2f14b86fdf923f 156B / 156B                                                                                     0.6s
 => => extracting sha256:93c1e223e6f2123b855e0c95898eba50cb6a055881ba9023527c0a361761c1cf                                                                                      0.0s
 => => extracting sha256:14bccb3a1cff3b5f09bf1f21ba9ecb17c96329fb82f96605d922cfa4c9a82032                                                                                      2.4s
 => => extracting sha256:861cc3991b26fba6073631f27a1dc32bb27761195d912d05ee2f14b86fdf923f                                                                                      0.0s
 => [stage-1 1/4] FROM gcr.io/distroless/static:nonroot@sha256:380318dd91fd3bea73ae5fe1eb4d795ef7923f576e6f5f8d4de6ef1ea18ed540                                                1.1s
 => => resolve gcr.io/distroless/static:nonroot@sha256:380318dd91fd3bea73ae5fe1eb4d795ef7923f576e6f5f8d4de6ef1ea18ed540                                                        0.0s
 => => sha256:380318dd91fd3bea73ae5fe1eb4d795ef7923f576e6f5f8d4de6ef1ea18ed540 1.67kB / 1.67kB                                                                                 0.0s
 => => sha256:daf333843de1f9c4f43807bf2554dd27c59db181af10d8d6a065d80044f45ec1 426B / 426B                                                                                     0.0s
 => => sha256:0d4f3baf10895230e85db46bed5a95d6f30f1c18a7d4ae1ac70aa25ee75baae3 478B / 478B                                                                                     0.0s
 => => sha256:79e0d8860fadaab56c716928c84875d99ff5e13787ca3fcced10b70af29bf320 801.34kB / 801.34kB                                                                             0.9s
 => => extracting sha256:79e0d8860fadaab56c716928c84875d99ff5e13787ca3fcced10b70af29bf320                                                                                      0.1s
 => [internal] load build context                                                                                                                                              0.5s
 => => transferring context: 90.55MB                                                                                                                                           0.5s
 => [builder 2/8] WORKDIR /workspace                                                                                                                                           1.4s
 => [builder 3/8] COPY go.mod go.mod                                                                                                                                           0.0s
 => [builder 4/8] COPY go.sum go.sum                                                                                                                                           0.0s
 => [builder 5/8] RUN go mod download                                                                                                                                         14.4s
 => [builder 6/8] COPY . .                                                                                                                                                     0.1s
 => [builder 7/8] RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags "-X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN' -X 'gi  14.4s
 => [builder 8/8] RUN mkdir -m 00755 /opt/manager                                                                                                                              0.2s
 => [stage-1 2/4] COPY --from=builder /workspace/manager .                                                                                                                     0.0s
 => [stage-1 3/4] COPY --from=builder --chown=65532:65532 /opt/manager /opt/manager                                                                                            0.0s
 => exporting to image                                                                                                                                                         0.3s
 => => exporting layers                                                                                                                                                        0.3s
 => => writing image sha256:feae312d1f6c4e4deae0e37c73bb5f23682cc797c9bc799bb96ef810e7a4fdbf                                                                                   0.0s
 => => naming to docker.io/kubeshop/kusk-gateway:latest                                                                                                                        0.0s
Build [kubeshop/kusk-gateway] succeeded
Starting test...
Tags used in deployment:
 - kubeshop/kusk-gateway -> kubeshop/kusk-gateway:feae312d1f6c4e4deae0e37c73bb5f23682cc797c9bc799bb96ef810e7a4fdbf
Starting deploy...
 - customresourcedefinition.apiextensions.k8s.io/apis.gateway.kusk.io created
 - customresourcedefinition.apiextensions.k8s.io/envoyfleet.gateway.kusk.io created
 - customresourcedefinition.apiextensions.k8s.io/staticroutes.gateway.kusk.io created
 - namespace/kusk-system created
 - customresourcedefinition.apiextensions.k8s.io/apis.gateway.kusk.io configured
 - customresourcedefinition.apiextensions.k8s.io/envoyfleet.gateway.kusk.io configured
 - customresourcedefinition.apiextensions.k8s.io/staticroutes.gateway.kusk.io configured
 - serviceaccount/kusk-gateway-manager created
 - role.rbac.authorization.k8s.io/kusk-gateway-leader-election-role created
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-envoyfleet-manager-role created
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-manager-role created
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-metrics-reader created
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-proxy-role created
 - rolebinding.rbac.authorization.k8s.io/kusk-gateway-leader-election-rolebinding created
 - clusterrolebinding.rbac.authorization.k8s.io/kusk-gateway-envoyfleet-manager-rolebinding created
 - clusterrolebinding.rbac.authorization.k8s.io/kusk-gateway-manager-rolebinding created
 - clusterrolebinding.rbac.authorization.k8s.io/kusk-gateway-proxy-rolebinding created
 - configmap/kusk-gateway-manager created
 - service/kusk-gateway-auth-service created
 - service/kusk-gateway-manager-metrics-service created
 - service/kusk-gateway-validator-service created
 - service/kusk-gateway-webhooks-service created
 - service/kusk-gateway-xds-service created
 - deployment.apps/kusk-gateway-manager created
 - mutatingwebhookconfiguration.admissionregistration.k8s.io/kusk-gateway-mutating-webhook-configuration created
 - validatingwebhookconfiguration.admissionregistration.k8s.io/kusk-gateway-validating-webhook-configuration created
 - configmap/config configured
Waiting for deployments to stabilize...
 - kusk-system:deployment/kusk-gateway-manager: waiting for rollout to finish: 0 of 1 updated replicas are available...
 - kusk-system:deployment/kusk-gateway-manager is ready.
Deployments stabilized in 11.134 seconds
Starting post-deploy hooks...
sleeping for 2 seconds before applying `config/samples/gateway_v1_envoyfleet.yaml`
Completed post-deploy hooks
 - envoyfleet.gateway.kusk.io/default created
Waiting for deployments to stabilize...
Deployments stabilized in 68.121516ms
You can also run [skaffold run --tail] to get the logs
$ kubectl get svc -A
kubectl get svc -A
NAMESPACE     NAME                                   TYPE           CLUSTER-IP       EXTERNAL-IP    PORT(S)                      AGE
default       default                                LoadBalancer   10.108.203.10    192.168.58.2   80:30103/TCP,443:32665/TCP   7s
default       kubernetes                             ClusterIP      10.96.0.1        <none>         443/TCP                      100s
kube-system   kube-dns                               ClusterIP      10.96.0.10       <none>         53/UDP,53/TCP,9153/TCP       99s
kusk-system   kusk-gateway-auth-service              ClusterIP      10.103.40.66     <none>         19000/TCP                    19s
kusk-system   kusk-gateway-manager-metrics-service   ClusterIP      10.100.170.224   <none>         8443/TCP                     19s
kusk-system   kusk-gateway-validator-service         ClusterIP      10.111.54.163    <none>         17000/TCP                    19s
kusk-system   kusk-gateway-webhooks-service          ClusterIP      10.96.171.109    <none>         443/TCP                      19s
kusk-system   kusk-gateway-xds-service               ClusterIP      10.97.183.42     <none>         18000/TCP                    19s
$ curl -v 192.168.58.2
*   Trying 192.168.58.2:80...
* Connected to 192.168.58.2 (192.168.58.2) port 80 (#0)
> GET / HTTP/1.1
> Host: 192.168.58.2
> User-Agent: curl/7.82.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< date: Tue, 11 Oct 2022 12:51:06 GMT
< server: envoy
< content-length: 0
<
* Connection #0 to host 192.168.58.2 left intact
```

Cleanup cluster
---------------

This deletes everything kusk-related in the cluster.

```sh
$ skaffold delete
Cleaning up...
 - customresourcedefinition.apiextensions.k8s.io "apis.gateway.kusk.io" deleted
 - customresourcedefinition.apiextensions.k8s.io "envoyfleet.gateway.kusk.io" deleted
 - customresourcedefinition.apiextensions.k8s.io "staticroutes.gateway.kusk.io" deleted
 - namespace "kusk-system" deleted
 - customresourcedefinition.apiextensions.k8s.io "envoyfleet.gateway.kusk.io" deleted
 - serviceaccount "kusk-gateway-manager" deleted
 - role.rbac.authorization.k8s.io "kusk-gateway-leader-election-role" deleted
 - clusterrole.rbac.authorization.k8s.io "kusk-gateway-envoyfleet-manager-role" deleted
 - clusterrole.rbac.authorization.k8s.io "kusk-gateway-manager-role" deleted
 - clusterrole.rbac.authorization.k8s.io "kusk-gateway-metrics-reader" deleted
 - clusterrole.rbac.authorization.k8s.io "kusk-gateway-proxy-role" deleted
 - rolebinding.rbac.authorization.k8s.io "kusk-gateway-leader-election-rolebinding" deleted
 - clusterrolebinding.rbac.authorization.k8s.io "kusk-gateway-envoyfleet-manager-rolebinding" deleted
 - clusterrolebinding.rbac.authorization.k8s.io "kusk-gateway-manager-rolebinding" deleted
 - clusterrolebinding.rbac.authorization.k8s.io "kusk-gateway-proxy-rolebinding" deleted
 - configmap "kusk-gateway-manager" deleted
 - service "kusk-gateway-auth-service" deleted
 - service "kusk-gateway-manager-metrics-service" deleted
 - service "kusk-gateway-validator-service" deleted
 - service "kusk-gateway-webhooks-service" deleted
 - service "kusk-gateway-xds-service" deleted
 - deployment.apps "kusk-gateway-manager" deleted
 - mutatingwebhookconfiguration.admissionregistration.k8s.io "kusk-gateway-mutating-webhook-configuration" deleted
 - validatingwebhookconfiguration.admissionregistration.k8s.io "kusk-gateway-validating-webhook-configuration" deleted
 - configmap "config" deleted
$ kubectl get svc -A
NAMESPACE     NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
default       kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP                  10m
kube-system   kube-dns     ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   10m
$ kubectl get pods -A
NAMESPACE        NAME                          READY   STATUS        RESTARTS   AGE
default          default-69d7898855-d76vr      1/1     Terminating   0          9m5s
kube-system      coredns-6d4b75cb6d-79rl2      1/1     Running       0          10m
kube-system      etcd-kgw                      1/1     Running       0          10m
kube-system      kube-apiserver-kgw            1/1     Running       0          10m
kube-system      kube-controller-manager-kgw   1/1     Running       0          10m
kube-system      kube-proxy-lks5k              1/1     Running       0          10m
kube-system      kube-scheduler-kgw            1/1     Running       0          10m
kube-system      storage-provisioner           1/1     Running       0          10m
metallb-system   controller-6f655c76ff-gszgf   1/1     Running       0          10m
metallb-system   speaker-d4ktr                 1/1     Running       0          10m
$ kubectl get crds -A
No resources found
$ kubectl get crds
No resources found
```

Modify a file
-------------

Execute `./skaffold.sh dev` then modify a `go` file and observe that the new image is deployed in the cluster.

```sh
$ ./skaffold.sh dev
Listing files to watch...
 - kubeshop/kusk-gateway
Generating tags...
 - kubeshop/kusk-gateway -> kubeshop/kusk-gateway:latest
Checking cache...
 - kubeshop/kusk-gateway: Not found. Building
Starting build...
Found [kgw] context, using local docker daemon.
Building [kubeshop/kusk-gateway]...
Target platforms: [linux/amd64]
[+] Building 16.2s (18/18) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                                           0.0s
 => => transferring dockerfile: 38B                                                                                                                                            0.0s
 => [internal] load .dockerignore                                                                                                                                              0.0s
 => => transferring context: 93B                                                                                                                                               0.0s
 => [internal] load metadata for gcr.io/distroless/static:nonroot                                                                                                              0.4s
 => [internal] load metadata for docker.io/library/golang:1.19-alpine                                                                                                          1.1s
 => [auth] library/golang:pull token for registry-1.docker.io                                                                                                                  0.0s
 => [builder 1/8] FROM docker.io/library/golang:1.19-alpine@sha256:f3e683657ddf73726b5717c2ff80cdcd9e9efb7d81f77e4948fada9a10dc7257                                            0.0s
 => [stage-1 1/4] FROM gcr.io/distroless/static:nonroot@sha256:380318dd91fd3bea73ae5fe1eb4d795ef7923f576e6f5f8d4de6ef1ea18ed540                                                0.0s
 => [internal] load build context                                                                                                                                              0.0s
 => => transferring context: 37.00kB                                                                                                                                           0.0s
 => CACHED [builder 2/8] WORKDIR /workspace                                                                                                                                    0.0s
 => CACHED [builder 3/8] COPY go.mod go.mod                                                                                                                                    0.0s
 => CACHED [builder 4/8] COPY go.sum go.sum                                                                                                                                    0.0s
 => CACHED [builder 5/8] RUN go mod download                                                                                                                                   0.0s
 => [builder 6/8] COPY . .                                                                                                                                                     0.1s
 => [builder 7/8] RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags "-X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN' -X 'gi  14.4s
 => [builder 8/8] RUN mkdir -m 00755 /opt/manager                                                                                                                              0.3s
 => CACHED [stage-1 2/4] COPY --from=builder /workspace/manager .                                                                                                              0.0s
 => CACHED [stage-1 3/4] COPY --from=builder --chown=65532:65532 /opt/manager /opt/manager                                                                                     0.0s
 => exporting to image                                                                                                                                                         0.0s
 => => exporting layers                                                                                                                                                        0.0s
 => => writing image sha256:feae312d1f6c4e4deae0e37c73bb5f23682cc797c9bc799bb96ef810e7a4fdbf                                                                                   0.0s
 => => naming to docker.io/kubeshop/kusk-gateway:latest                                                                                                                        0.0s
Build [kubeshop/kusk-gateway] succeeded
Tags used in deployment:
 - kubeshop/kusk-gateway -> kubeshop/kusk-gateway:feae312d1f6c4e4deae0e37c73bb5f23682cc797c9bc799bb96ef810e7a4fdbf
Starting deploy...
 - customresourcedefinition.apiextensions.k8s.io/apis.gateway.kusk.io configured
 - customresourcedefinition.apiextensions.k8s.io/envoyfleet.gateway.kusk.io configured
 - customresourcedefinition.apiextensions.k8s.io/staticroutes.gateway.kusk.io configured
 - namespace/kusk-system unchanged
 - customresourcedefinition.apiextensions.k8s.io/apis.gateway.kusk.io configured
 - customresourcedefinition.apiextensions.k8s.io/envoyfleet.gateway.kusk.io configured
 - customresourcedefinition.apiextensions.k8s.io/staticroutes.gateway.kusk.io configured
 - serviceaccount/kusk-gateway-manager unchanged
 - role.rbac.authorization.k8s.io/kusk-gateway-leader-election-role unchanged
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-envoyfleet-manager-role unchanged
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-manager-role configured
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-metrics-reader unchanged
 - clusterrole.rbac.authorization.k8s.io/kusk-gateway-proxy-role unchanged
 - rolebinding.rbac.authorization.k8s.io/kusk-gateway-leader-election-rolebinding unchanged
 - clusterrolebinding.rbac.authorization.k8s.io/kusk-gateway-envoyfleet-manager-rolebinding unchanged
 - clusterrolebinding.rbac.authorization.k8s.io/kusk-gateway-manager-rolebinding unchanged
 - clusterrolebinding.rbac.authorization.k8s.io/kusk-gateway-proxy-rolebinding unchanged
 - configmap/kusk-gateway-manager unchanged
 - service/kusk-gateway-auth-service configured
 - service/kusk-gateway-manager-metrics-service configured
 - service/kusk-gateway-validator-service configured
 - service/kusk-gateway-webhooks-service configured
 - service/kusk-gateway-xds-service configured
 - deployment.apps/kusk-gateway-manager configured
 - mutatingwebhookconfiguration.admissionregistration.k8s.io/kusk-gateway-mutating-webhook-configuration configured
 - validatingwebhookconfiguration.admissionregistration.k8s.io/kusk-gateway-validating-webhook-configuration configured
 - configmap/config unchanged
Waiting for deployments to stabilize...
 - kusk-system:deployment/kusk-gateway-manager: waiting for rollout to finish: 1 old replicas are pending termination...
 - kusk-system:deployment/kusk-gateway-manager is ready.
Deployments stabilized in 11.13 seconds
Starting post-deploy hooks...
sleeping for 2 seconds before applying `config/samples/gateway_v1_envoyfleet.yaml`
Completed post-deploy hooks
 - envoyfleet.gateway.kusk.io/default configured
Waiting for deployments to stabilize...
Deployments stabilized in 52.110579ms
Press Ctrl+C to exit
[manager] METRICS_BIND_ADDR=127.0.0.1:8080
[kube-rbac-proxy] I1011 12:57:17.430810       1 main.go:190] Valid token audiences:
[manager] HEALTH_PROBE_BIND_ADDR=:8081
[manager] ENVOY_CONTROL_PLANE_BIND_ADDR=:18000
[manager] ENABLE_LEADER_ELECTION=false
[manager] LOG_LEVEL=INFO
[manager] WEBHOOK_CERTS_DIR=/tmp/k8s-webhook-server/serving-certs
[manager] ANALYTICS_ENABLED=true
[kube-rbac-proxy] I1011 12:57:17.430853       1 main.go:262] Generating self signed cert as no cert is provided
[manager]
[manager] {"level":"info","ts":1665493037.510534,"logger":"controller-runtime.metrics","caller":"logr@v1.2.3/logr.go:261","msg":"Metrics server is starting to listen","addr":"127.0.0.1:8080"}
[manager] {"level":"info","ts":1665493037.5108223,"logger":"setup","caller":"manager/main.go:244","msg":"Starting Envoy xDS API Server"}
[kube-rbac-proxy] I1011 12:57:17.686846       1 main.go:311] Starting TCP socket on 0.0.0.0:8443
[kube-rbac-proxy] I1011 12:57:17.686986       1 main.go:318] Listening securely on 0.0.0.0:8443
[manager] {"level":"info","ts":1665493037.5108833,"caller":"authz/authz.go:53","msg":"authz listening on","address":":19000"}
[manager] {"level":"info","ts":1665493037.5109699,"logger":"EnvoyConfigManager","caller":"manager/envoy_config_manager.go:79","msg":"control plane server listening","address":":18000"}
[manager] {"level":"info","ts":1665493038.3406372,"logger":"setup","caller":"manager/main.go:289","msg":"Starting K8s secrets watch for the TLS certificates renewal events"}
[manager] {"level":"info","ts":1665493039.900979,"logger":"setup","caller":"manager/main.go:309","msg":"Created admission webhook server certificates and updated K8s Manager's Admission configs with the generated CA certificate"}
[manager] {"level":"info","ts":1665493039.9010339,"logger":"setup","caller":"manager/main.go:311","msg":"Registering EnvoyFleet mutating and validating webhooks to the webhook server"}
[manager] {"level":"info","ts":1665493039.90111,"logger":"controller-runtime.webhook","caller":"webhook/server.go:148","msg":"Registering webhook","path":"/mutate-gateway-kusk-io-v1alpha1-envoyfleet"}
[manager] {"level":"info","ts":1665493039.9012144,"logger":"controller-runtime.webhook","caller":"webhook/server.go:148","msg":"Registering webhook","path":"/validate-gateway-kusk-io-v1alpha1-envoyfleet"}
[manager] {"level":"info","ts":1665493039.901379,"logger":"setup","caller":"manager/main.go:327","msg":"Registering API mutating and validating webhooks to the webhook server"}
[manager] {"level":"info","ts":1665493039.9014447,"logger":"controller-runtime.webhook","caller":"webhook/server.go:148","msg":"Registering webhook","path":"/mutate-gateway-kusk-io-v1alpha1-api"}
[manager] {"level":"info","ts":1665493039.9015114,"logger":"controller-runtime.webhook","caller":"webhook/server.go:148","msg":"Registering webhook","path":"/validate-gateway-kusk-io-v1alpha1-api"}
[manager] {"level":"info","ts":1665493039.90162,"logger":"setup","caller":"manager/main.go:343","msg":"Registering StaticRoute mutating and validating webhooks to the webhook server"}
[manager] {"level":"info","ts":1665493039.9016645,"logger":"controller-runtime.webhook","caller":"webhook/server.go:148","msg":"Registering webhook","path":"/mutate-gateway-kusk-io-v1alpha1-staticroute"}
[manager] {"level":"info","ts":1665493039.9017282,"logger":"controller-runtime.webhook","caller":"webhook/server.go:148","msg":"Registering webhook","path":"/validate-gateway-kusk-io-v1alpha1-staticroute"}
[manager] {"level":"info","ts":1665493039.9017673,"logger":"setup","caller":"manager/main.go:357","msg":"Starting manager"}
[manager] {"level":"info","ts":1665493039.901852,"logger":"controller-runtime.webhook.webhooks","caller":"webhook/server.go:216","msg":"Starting webhook server"}
[manager] {"level":"info","ts":1665493039.90191,"caller":"manager/internal.go:362","msg":"Starting server","kind":"health probe","addr":"[::]:8081"}
[manager] {"level":"info","ts":1665493039.9019277,"caller":"manager/internal.go:362","msg":"Starting server","path":"/metrics","kind":"metrics","addr":"127.0.0.1:8080"}
[manager] {"level":"info","ts":1665493039.902084,"logger":"controller-runtime.certwatcher","caller":"logr@v1.2.3/logr.go:261","msg":"Updated current TLS certificate"}
[manager] {"level":"info","ts":1665493039.902163,"logger":"controller-runtime.webhook","caller":"logr@v1.2.3/logr.go:261","msg":"Serving webhook server","host":"","port":9443}
[manager] {"level":"info","ts":1665493039.9021924,"logger":"controller-runtime.certwatcher","caller":"logr@v1.2.3/logr.go:261","msg":"Starting certificate watcher"}
[manager] {"level":"info","ts":1665493040.0029674,"caller":"controller/controller.go:185","msg":"Starting EventSource","controller":"envoyfleet","controllerGroup":"gateway.kusk.io","controllerKind":"EnvoyFleet","source":"kind source: *v1alpha1.EnvoyFleet"}
[manager] {"level":"info","ts":1665493040.003018,"caller":"controller/controller.go:185","msg":"Starting EventSource","controller":"staticroute","controllerGroup":"gateway.kusk.io","controllerKind":"StaticRoute","source":"kind source: *v1alpha1.StaticRoute"}
[manager] {"level":"info","ts":1665493040.0030375,"caller":"controller/controller.go:193","msg":"Starting Controller","controller":"envoyfleet","controllerGroup":"gateway.kusk.io","controllerKind":"EnvoyFleet"}
[manager] {"level":"info","ts":1665493040.0030437,"caller":"controller/controller.go:193","msg":"Starting Controller","controller":"staticroute","controllerGroup":"gateway.kusk.io","controllerKind":"StaticRoute"}
[manager] {"level":"info","ts":1665493040.0031295,"caller":"controller/controller.go:185","msg":"Starting EventSource","controller":"api","controllerGroup":"gateway.kusk.io","controllerKind":"API","source":"kind source: *v1alpha1.API"}
[manager] {"level":"info","ts":1665493040.0031805,"caller":"controller/controller.go:193","msg":"Starting Controller","controller":"api","controllerGroup":"gateway.kusk.io","controllerKind":"API"}
[manager] {"level":"info","ts":1665493040.1039643,"caller":"controller/controller.go:227","msg":"Starting workers","controller":"api","controllerGroup":"gateway.kusk.io","controllerKind":"API","worker count":1}
[manager] {"level":"info","ts":1665493040.1039562,"caller":"controller/controller.go:227","msg":"Starting workers","controller":"staticroute","controllerGroup":"gateway.kusk.io","controllerKind":"StaticRoute","worker count":1}
[manager] {"level":"info","ts":1665493040.1040132,"caller":"controller/controller.go:227","msg":"Starting workers","controller":"envoyfleet","controllerGroup":"gateway.kusk.io","controllerKind":"EnvoyFleet","worker count":1}
[manager] {"level":"info","ts":1665493040.1041217,"logger":"envoy-fleet-controller","caller":"controllers/envoyfleet_controller.go:68","msg":"EnvoyFleet changed","controller":"envoyfleet","controllerGroup":"gateway.kusk.io","controllerKind":"EnvoyFleet","envoyFleet":{"name":"default","namespace":"default"},"namespace":"default","name":"default","reconcileID":"10ddb422-9c9d-4596-8ba8-2996c1627e69","changed":"default/default"}
[manager] {"level":"info","ts":1665493040.4170942,"logger":"envoy-fleet-controller","caller":"controllers/envoyfleet_controller.go:113","msg":"Calling Config Manager due to change in Envoy Fleet resource","controller":"envoyfleet","controllerGroup":"gateway.kusk.io","controllerKind":"EnvoyFleet","envoyFleet":{"name":"default","namespace":"default"},"namespace":"default","name":"default","reconcileID":"10ddb422-9c9d-4596-8ba8-2996c1627e69","changed":"default/default"}
[manager] {"level":"info","ts":1665493040.4171255,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Started updating configuration","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.417136,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Getting APIs for the fleet","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4184847,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Successfully processed APIs","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4185064,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Getting Static Routes","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4185386,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Successfully processed Static Routes","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4185548,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Processing EnvoyFleet configuration","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4209602,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Generating configuration snapshot","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4215462,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Configuration snapshot was generated for the fleet","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4216008,"logger":"CacheManager","caller":"manager/cache_manager.go:114","msg":"assigning active snapshot and updating all nodes","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4216113,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Configuration snapshot deployed for the fleet","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4216185,"logger":"controller.config-manager","caller":"logr@v1.2.3/logr.go:261","msg":"Finished updating configuration","fleet":"default.default"}
[manager] {"level":"info","ts":1665493040.4216316,"logger":"envoy-fleet-controller","caller":"controllers/envoyfleet_controller.go:124","msg":"Reconciled EnvoyFleet 'default' resources","controller":"envoyfleet","controllerGroup":"gateway.kusk.io","controllerKind":"EnvoyFleet","envoyFleet":{"name":"default","namespace":"default"},"namespace":"default","name":"default","reconcileID":"10ddb422-9c9d-4596-8ba8-2996c1627e69"}
[manager] {"level":"error","ts":1665493047.4272718,"logger":"SnapshotCache","caller":"manager/envoy_snapshot_cache_logger.go:58","msg":"node does not exist","stacktrace":"github.com/kubeshop/kusk-gateway/internal/envoy/manager.EnvoySnapshotCacheLogger.Warnf\n\t/workspace/internal/envoy/manager/envoy_snapshot_cache_logger.go:58\ngithub.com/envoyproxy/go-control-plane/pkg/cache/v3.(*snapshotCache).GetStatusInfo\n\t/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.10.3/pkg/cache/v3/simple.go:588\ngithub.com/kubeshop/kusk-gateway/internal/envoy/manager.(*cacheManager).IsNodeExist\n\t/workspace/internal/envoy/manager/cache_manager.go:54\ngithub.com/kubeshop/kusk-gateway/internal/envoy/manager.(*Callbacks).OnStreamRequest\n\t/workspace/internal/envoy/manager/envoy_callbacks.go:65\ngithub.com/envoyproxy/go-control-plane/pkg/server/sotw/v3.(*server).process\n\t/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.10.3/pkg/server/sotw/v3/server.go:181\ngithub.com/envoyproxy/go-control-plane/pkg/server/sotw/v3.(*server).StreamHandler\n\t/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.10.3/pkg/server/sotw/v3/server.go:256\ngithub.com/envoyproxy/go-control-plane/pkg/server/v3.(*server).StreamHandler\n\t/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.10.3/pkg/server/v3/server.go:183\ngithub.com/envoyproxy/go-control-plane/pkg/server/v3.(*server).StreamRoutes\n\t/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.10.3/pkg/server/v3/server.go:199\ngithub.com/envoyproxy/go-control-plane/envoy/service/route/v3._RouteDiscoveryService_StreamRoutes_Handler\n\t/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.10.3/envoy/service/route/v3/rds.pb.go:341\ngoogle.golang.org/grpc.(*Server).processStreamingRPC\n\t/go/pkg/mod/google.golang.org/grpc@v1.47.0/server.go:1542\ngoogle.golang.org/grpc.(*Server).handleStream\n\t/go/pkg/mod/google.golang.org/grpc@v1.47.0/server.go:1624\ngoogle.golang.org/grpc.(*Server).serveStreams.func1.2\n\t/go/pkg/mod/google.golang.org/grpc@v1.47.0/server.go:922"}
[manager] {"level":"info","ts":1665493047.427347,"logger":"CacheManager","caller":"manager/cache_manager.go:82","msg":"setting new node snapshot","nodeID":"default-69d7898855-d76vr","fleet":"default.default"}
Watching for changes...
Generating tags...
 - kubeshop/kusk-gateway -> kubeshop/kusk-gateway:latest
Checking cache...
 - kubeshop/kusk-gateway: Not found. Building
Starting build...
Found [kgw] context, using local docker daemon.
Building [kubeshop/kusk-gateway]...
Target platforms: [linux/amd64]
[+] Building 15.9s (17/17) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                                           0.0s
 => => transferring dockerfile: 38B                                                                                                                                            0.0s
 => [internal] load .dockerignore                                                                                                                                              0.0s
 => => transferring context: 93B                                                                                                                                               0.0s
 => [internal] load metadata for gcr.io/distroless/static:nonroot                                                                                                              0.5s
 => [internal] load metadata for docker.io/library/golang:1.19-alpine                                                                                                          0.6s
 => CACHED [stage-1 1/4] FROM gcr.io/distroless/static:nonroot@sha256:380318dd91fd3bea73ae5fe1eb4d795ef7923f576e6f5f8d4de6ef1ea18ed540                                         0.0s
 => [builder 1/8] FROM docker.io/library/golang:1.19-alpine@sha256:f3e683657ddf73726b5717c2ff80cdcd9e9efb7d81f77e4948fada9a10dc7257                                            0.0s
 => [internal] load build context                                                                                                                                              0.0s
 => => transferring context: 49.34kB                                                                                                                                           0.0s
 => CACHED [builder 2/8] WORKDIR /workspace                                                                                                                                    0.0s
 => CACHED [builder 3/8] COPY go.mod go.mod                                                                                                                                    0.0s
 => CACHED [builder 4/8] COPY go.sum go.sum                                                                                                                                    0.0s
 => CACHED [builder 5/8] RUN go mod download                                                                                                                                   0.0s
 => [builder 6/8] COPY . .                                                                                                                                                     0.1s
 => [builder 7/8] RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags "-X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$TELEMETRY_TOKEN' -X 'gi  14.4s
 => [builder 8/8] RUN mkdir -m 00755 /opt/manager                                                                                                                              0.3s
 => [stage-1 2/4] COPY --from=builder /workspace/manager .                                                                                                                     0.0s
 => [stage-1 3/4] COPY --from=builder --chown=65532:65532 /opt/manager /opt/manager                                                                                            0.0s
 => exporting to image                                                                                                                                                         0.3s
 => => exporting layers                                                                                                                                                        0.3s
 => => writing image sha256:87c1c6ca8dd29734a4b96c81c408e6369a34949822f7e35c98db15e7ee8cb60a                                                                                   0.0s
 => => naming to docker.io/kubeshop/kusk-gateway:latest                                                                                                                        0.0s
Build [kubeshop/kusk-gateway] succeeded
Tags used in deployment:
 - kubeshop/kusk-gateway -> kubeshop/kusk-gateway:87c1c6ca8dd29734a4b96c81c408e6369a34949822f7e35c98db15e7ee8cb60a
Starting deploy...
 - deployment.apps/kusk-gateway-manager configured
Waiting for deployments to stabilize...
 - kusk-system:deployment/kusk-gateway-manager: container manager terminated with exit code 1
    - kusk-system:pod/kusk-gateway-manager-8547cd469c-7bml5: container manager terminated with exit code 1
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] METRICS_BIND_ADDR=127.0.0.1:8080
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] HEALTH_PROBE_BIND_ADDR=:8081
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] ENVOY_CONTROL_PLANE_BIND_ADDR=:18000
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] ENABLE_LEADER_ELECTION=false
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] LOG_LEVEL=INFO
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] WEBHOOK_CERTS_DIR=/tmp/k8s-webhook-server/serving-certs
      > [kusk-gateway-manager-8547cd469c-7bml5 manager] ANALYTICS_ENABLED=true
 - kusk-system:deployment/kusk-gateway-manager failed. Error: container manager terminated with exit code 1.
WARN[0062] Skipping deploy due to error:1/1 deployment(s) failed  subtask=-1 task=DevLoop
Watching for changes...
```

Debugging
---------

To start in debug mode run:

```sh
$ ./skaffold dev
$ dlv connect 127.0.0.1:56268
(dlv) stack
0  0x000000000044235d in runtime.gopark
   at /usr/local/go/src/runtime/proc.go:364
1  0x0000000000451e89 in runtime.selectgo
   at /usr/local/go/src/runtime/select.go:328
2  0x0000000001cc0f1f in sigs.k8s.io/controller-runtime/pkg/manager.(*controllerManager).Start
   at /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.3/pkg/manager/internal.go:500
3  0x00000000023d051c in main.main
   at /workspace/cmd/manager/main.go:358
4  0x0000000000441f38 in runtime.main
   at /usr/local/go/src/runtime/proc.go:250
5  0x0000000000471e81 in runtime.goexit
   at /usr/local/go/src/runtime/asm_amd64.s:1594
```

As can seen from above, that's controller's stackframe.

TODO
----

* Currently I've hardcoded a value `runAsNonRoot: false` in `config/manager/manager.yaml` to enable debugging. This isn't ideal or correct. There are better ways of doing this.
* Multiple Platform Images: Investigate <https://github.com/GoogleContainerTools/skaffold/tree/main/examples/custom-buildx>.
