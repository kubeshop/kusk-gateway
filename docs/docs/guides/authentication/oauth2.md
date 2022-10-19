# OAuth2 via OIDC

OAuth2 ensures that your application (upstream) doesn't get requests which are not authenticated and authorized. It effectively helps to protect your API.

Kusk implements the OpenID Connect (OIDC) layer which ensures you can choose any OAuth2 provider and Kusk will support it as long as the OAuth2 provider also supports OIDC.

Kusk makes it easy to configure OAuth2, using the `auth` option in the `x-kusk` extension.

:::caution

The OAuth2 feature is currently under active development. See [upstream issues](#upstream-issues).

:::

## Setup Guide

We'll go through step-by-step of configuring OAuth2. In this example we'll be using [Auth0](https://auth0.com/) as OAuth2 provider.
### 1. Configuring Auth0

1. Signup for an account at [Auth0](https://auth0.com/).
2. Create an Auth0 Application
3. Configure the following Auth0 Application fields: 
    1. Allowed Callback URLs (e.g. `http://**KUSK_IP**/oauth2/callback`)
    2. Allowed Logout URLs (e.g. `http://**KUSK_IP**/oauth2/signout`)
(**KUSK_IP** is the IP of the Kusk's LoadBalancer, you can obtain it by running `kusk ip`)
4. Take note of the credentials as we will need this later on:

```json
{
  "client_id": "go8brZmF6eE6r7TObzpGaD5KFjJkm6Qb",
  "client_secret": "bkryzZGGA6Ko0VGnUEl_1YeREMHDpjGP8r1BTN1HYlmXpAWaiWNkD4bqIDuAuCKV"
}
```

### 2. Deploy a protected API to Kusk

The example below ensures the whole API is protected via OAuth2, and that the upstream `auth-oauth2-oauth0-authorization-code-grant-go-httpbin` can be only accessed when authenticated and authorized.

```yaml title="api.yaml"
openapi: 3.0.0
info:
  title: oauth2-example
  description: oauth2-example
  version: '0.1.0'
x-kusk:
  mocking:
    enabled: true
  auth:
    oauth2:
      token_endpoint: https://**YOUR_DOMAIN**.eu.auth0.com/oauth/token
      authorization_endpoint: https://**YOUR_DOMAIN**.eu.auth0.com/authorize
      credentials:
        client_id: **CLIENTID**
        client_secret: **CLIENT_SECRET**
      redirect_uri: /oauth2/callback
      redirect_path_matcher: /oauth2/callback
      signout_path: /oauth2/signout
      forward_bearer_token: true
      auth_scopes:
        - openid
paths:
  "/":
    get:
      description: Returns GET data.
      responses:
        200:
          content: 
            text/plain:
              example: Hello from a mocked response
```

You are required to change:

1. `token_endpoint`.
2. `authorization_endpoint`.
3. `credentials.client_id`.
4. `credentials.client_secret`.
5. `auth_scopes`: Strictly speaking this is not required but we strongly suggest entering `openid` for testing purposes

After that, deploy the API by running: 

```
kusk deploy -i api.yaml
```

### 4. Update EnvoyFleet ConfigMap

We need to include to `client_secret` to Envoy's `ConfigMap`, by running 

```
kubectl edit -n kusk-system configmaps kusk-gateway-envoy-fleet
```

And then update the field `inline_string` replacing it with the `client_secret` obtained from Auth0. 

```yaml
  secrets:
  - name: token
    generic_secret:
      secret:
        inline_string: "<stub_token_secret>" # <- replace with "CLIENT_SECRET"
```

Finally, you should save the configuration and the changes will be applied. 

### 5. Restart Envoy Fleet

As we currently have upstream issues with Envoy waiting to be fixed, the temporal solution is to restart Envoy manually (this only needs to be done once): 

For this:

1. Port-Forward into Envoy's control plane: 

```sh
kubectl port-forward --namespace kusk-system deployment/kusk-gateway-envoy-fleet 19000:19000
```
2. Restart Envoy by making a POST request to `/quitquitquit`

```sh
curl -X POST 'http://localhost:19000/quitquitquit'
```

3. Wait until the changes to be propogated, this could take a while. To know if the Envoy has restarted, check if the Envoy container has restarted

<pre>
kubectl get pods -w -n kusk-system --selector=app.kubernetes.io/instance=kusk-gateway-envoy-fleet
<br />
<br />
NAME                                        READY   STATUS    RESTARTS      AGE
<br />
kusk-gateway-envoy-fleet-6f9ff68bcd-jxlwk   1/1     <b>Running</b>   1 (30s ago)   1h
</pre>

### 6. Test using the browser

You're all set now, test your OAuth2 implementation through the browser by visiting Kusk's LoadBalancer. Get the URL with: 

```sh
kusk ip

100.12.34.56
```

--- 

## Upstream Issues

Certain OAuth2 features are blocked/constrained by upstream issues. Please see:

* [Segmentation Fault after `assert failure: false. Details: attempted to add shared target SdsApi <NAME_OF_SECRET> to initialized init manager Server](https://github.com/envoyproxy/envoy/issues/22678).
* [`SecretDiscoveryServiceServer`: `StreamSecrets` issues](https://github.com/envoyproxy/go-control-plane/issues/581).

So the implementation is constrained by these issues.