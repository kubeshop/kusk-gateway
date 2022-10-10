# Helm installation 

A full Kusk installation is conformed of Kusk Gateway, EnvoyFleet, Kusk API and Kusk Dashboard.

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
  --create-namespace \
  --namespace kusk-system \
  --set service.type=LoadBalancer \
  kusk-gateway-envoy-fleet \
  kubeshop/kusk-gateway-envoyfleet
```

## 3. Install Kusk API and Kusk Dashboard (optional)

Kusk Dashboard allows you to configure Kusk Gateway from a neat GUI. To deploy this, you will have to expose it on a private EnvoyFleet (of type `ClusterIP`), so Kusk Dashboard is not accessed publicly:

```sh
helm upgrade \
  --install \
  --wait \
  --create-namespace \
  --namespace kusk-system \
  --set fullnameOverride=kusk-gateway-private-envoy-fleet \
  --set service.type=ClusterIP \
  kusk-gateway-private-envoy-fleet \
  kubeshop/kusk-gateway-envoyfleet
```

Now install Kusk API, which is an API server that Kusk Dashboard uses to configure the gateway: 

```sh
helm upgrade \
  --install \
  --wait \
  --create-namespace \
  --namespace kusk-system \
  --set fullnameOverride=kusk-gateway-api \
  --set envoyfleet.name=kusk-gateway-private-envoy-fleet \
  --set envoyfleet.namespace=kusk-system \
  kusk-gateway-api \
  kubeshop/kusk-gateway-api
```

And finally install Kusk Dashboard: 

```sh
helm upgrade \
  --install \
  --wait \
  --create-namespace \
  --namespace kusk-system \
  --set fullnameOverride=kusk-gateway-dashboard \
  --set envoyfleet.name=kusk-gateway-private-envoy-fleet \
  --set envoyfleet.namespace=kusk-system \
  kusk-gateway-dashboard \
  kubeshop/kusk-gateway-dashboard
```