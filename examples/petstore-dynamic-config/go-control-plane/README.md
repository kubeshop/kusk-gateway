# These files are copied and changed from [Go Control Plane example](https://github.com/envoyproxy/go-control-plane) to make easier prototyping

# Example xDS Server

This is an example of a trivial xDS V3 control plane server.  It serves an Envoy configuration that's roughly equivalent to the one used by the Envoy ["Quick Start"](https://www.envoyproxy.io/docs/envoy/latest/start/start#quick-start-to-run-simple-example) docs: a simple http proxy.

Use Dockerfile above to build it.

## Files

* [main/main.go](main/main.go) is the example program entrypoint.  It instantiates the cache and xDS server and runs the xDS server process.
* [resource.go](resource.go) generates a `Snapshot` structure which describes the configuration that the xDS server serves to Envoy.
* [server.go](server.go) runs the xDS control plane server.
* [logger.go](logger.go) implements the `pkg/log/Logger` interface which provides logging services to the cache.

