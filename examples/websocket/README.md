# Websocket Example

This directory provides the example to test Websockets enabled HTTP connection. TCP passthrough is currently not supported by KGW.
We use [solsson/websocat](https://hub.docker.com/r/solsson/websocat) container image as a WS server and a client.

Websockets are per route and can be enabled globally (for all endpoints), but to test it we need a WS (websocat) server or a specifically written application with Websockets.

To test:

1. Deploy the KGW with the EnvoyFleet.
2. Deploy this directory with ```kubectl apply -f examples/websocket``` .
3. Run test, by connecting to the external IP address of EnvoyFleet and typing anything when it runs - the response should be HELO from ws service. We need to use "--network=host" for Docker to see any Kubernetes clusters deployed on the local host.

```sh
// endpoint from the StaticRoute definition
docker run --rm -ti --network=host solsson/websocat ws://ExternalIPAddress>:80/staticwebsocket

BBB
'[ws] HELO'

// endpoint from the API definition
docker run --rm -ti --network=host solsson/websocat ws://ExternalIPAddress>:80/apiwebsocket

BBB
'[ws] HELO'

// endpoint from the API definition, this one must fail
docker run --rm -ti --network=host solsson/websocat ws://ExternalIPAddress>:80/disabledapiwebsocket
websocat: WebSocketError: WebSocket response error
websocat: error running
 ```

In the stdout of the Envoy container should be access log entry like:

```
"[2022-01-04T12:31:16.991Z]" "GET" "/staticewebsocket" "101" "4298"
```

where 101 is the response HTTP code from Envoy Proxy to the request with HTTP Upgrade header present during the tunnel setup.
