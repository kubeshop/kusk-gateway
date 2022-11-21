# 2. Launch a Kubernetes Cluster

Kusk needs to be installed in a Kubernetes cluster to serve its traffic.

You can start a local Kubernetes cluster or connect to a remote cluster. In this tutorial, you'll find instructions to start a local cluster using [Minikube](https://minikube.sigs.k8s.io/docs/), which will help you get started with Kusk.

For more information on the different options for running a Kubernetes cluster locally or remotely, [check this great resource](https://docs.tilt.dev/choosing_clusters.html) which contains a vast comparison list.

Install Minikube

Use the installation guide from Minikube [here](https://minikube.sigs.k8s.io/docs/start/). 

Start your Minikube cluster

```sh
minikube start
```
```sh title="Expected output:"
😄  minikube v1.28.0 on Darwin 13.0 (arm64)
✨  Automatically selected the docker driver
📌  Using Docker Desktop driver with root privileges
👍  Starting control plane node minikube in cluster minikube
🚜  Pulling base image ...
    > gcr.io/k8s-minikube/kicbase:  0 B [_______________________] ?% ? p/s 1m5s
🔥  Creating docker container (CPUs=2, Memory=7802MB) ...
🐳  Preparing Kubernetes v1.25.3 on Docker 20.10.20 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```
