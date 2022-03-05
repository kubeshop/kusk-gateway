# Using goproxy to speedup modules download

During the development with Minikube clusters are cattle, so they are being started and deleted without any trace.
Docker builds inside the cluster use downloaded modules cache only on the second build and each container image cache layer could be independent.

It is possible to start a proxy, that caches the downloaded modules and thus can speed up the development (for the Kusk Gateway the difference is 6 seconds with the proxy vs 30 seconds without).
GOPROXY Go variable controls whether the proxy is used during the build.
Thus, in order to speed up the development, a developer can start goproxy instance that bounds to the accessible for the containers IP address and set GORPXOY variable for the build.
For the Docker build we need to pass this variable as a build argument, i.e. code the support for this scheme in Dockerfile.

Steps to setup:

1. It is advised to copy this directory and start goproxy as a constantly running container for all builds on this machine, not only this exactly project.
2. Run `docker-compose up -d` in gorpoxy directory. This will create the container with host networking, meaning that it binds to 8085 port on all interfaces. Container will use a new local container volume for the cache, which will not be deleted with `docker-compose down`.
3. Set `export GOPROXY=<your_stable_ip_address:8085>` in your shell startup configuration, e.g. .bashrc. The address shouldn't be 127.0.0.1. Restart the terminal to pick up the change, check with `echo $GORPOXY`.
4. For the Docker builds, make sure that you have the following in Dockerfile before the step to download modules:

   ```
   ARG GOPROXY
   ENV GOPROXY=$GOPROXY
   RUN go mod download
   ```

5. When running Docker builds, pass this GOPROXY variable as an argument: `docker build --build-arg GOPROXY=${GOPROXY}`
