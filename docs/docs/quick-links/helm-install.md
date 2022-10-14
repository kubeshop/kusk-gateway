# Helm installation 

A full Kusk installation is conformed of: 

- Kusk Gateway: manages the gateway by translating your OpenAPI definition
- EnvoyFleet: the API's LoadBalancer
- Kusk Dashboard: provides a graphical experience to work with Kusk Gateway
- Kusk API: Kusk API server needed to control Kusk through Kusk Dashboard

You can install all these components together (which we recommend), but if you don't need to use the dashboard, then you can skip installing Kusk Dashboard and Kusk API.

## 1. Add the Helm repository 

```sh 
helm repo add kubeshop https://kubeshop.github.io/helm-charts
helm repo update
```

## 2. Install Kusk Gateway

Install Kusk Manager first:

```sh 
helm upgrade \
  --install \
  --wait \
  --create-namespace \
  --namespace kusk-system \
  kusk-gateway \
  kubeshop/kusk-gateway 
```

And now install the default EnvoyFleet (named `kusk-gateway-envoy-fleet`) which will serve public requests to your API: 

```sh 
helm upgrade \
  --install \
  --wait \
  --set service.type=LoadBalancer \
  kusk-gateway-envoy-fleet \
  kubeshop/kusk-gateway-envoyfleet
```

## 3. Install Kusk API and Kusk Dashboard (optional)

Kusk Dashboard allows you to configure Kusk Gateway from a neat GUI. To deploy this, you will have to:

1. Create a private EnvoyFleet: a service of type `ClusterIP`, so Kusk Dashboard isn't accessed publicly
2. Kusk API - configured to be exposed through the private EnvoyFleet
3. Kusk Dashboard - configured to be connected to Kusk API and exposed through the private EnvoyFleet

### 1. Create a Private EnvoyFleet
```sh
helm upgrade \
  --install \
  --wait \
  --set fullnameOverride=kusk-gateway-private-envoy-fleet \
  --set service.type=ClusterIP \
  kusk-gateway-private-envoy-fleet \
  kubeshop/kusk-gateway-envoyfleet
```
### 2. Install Kusk API

Now install Kusk API, which is an API server that Kusk Dashboard uses to configure the gateway: 

```sh
helm upgrade \
  --install \
  --wait \
  --set envoyfleet.name=kusk-gateway-private-envoy-fleet \
  --set envoyfleet.namespace=kusk-system \
  kusk-gateway-api \
  kubeshop/kusk-gateway-api
```

### 3. Install Kusk Dashboard

```sh
helm upgrade \
  --install \
  --wait \
  --set envoyfleet.name=kusk-gateway-private-envoy-fleet \
  --set envoyfleet.namespace=kusk-system \
  kusk-gateway-dashboard \
  kubeshop/kusk-gateway-dashboard
```