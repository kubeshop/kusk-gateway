# `kusk cluster install`

The `install` command will install Kusk Gateway and all its components with a single command. 
Kusk uses Helm to do this, so you will need to have [Helm installed](https://helm.sh/docs/intro/install/).

### **Kusk Gateway Components**

* **Kusk Gateway Manager** - Responsible for updating and rolling out the Envoy configuration to your Envoy Fleets as you deploy APIs and Static Routes.
* **Envoy Fleet** - Responsible for exposing and routing to your APIs and frontends.
* **Kusk Gateway API** - REST API, which is exposed by Kusk Gateway and allows you to programmatically query which APIs, Static Routes and Envoy Fleets are deployed.
* **Kusk Gateway Dashboard** - A web UI for Kusk Gateway where you can deploy APIs and see which APIs, StaticRoutes and Envoy Fleets are deployed.

#### **Examples**

The default `kusk cluster install` command will install Kusk Gateway, a public (for your APIs) and private (for the Kusk dashboard and API)
envoy-fleet, api, and dashboard in the kusk-system namespace using Helm and using the current kubeconfig context.

```sh
$ kusk cluster install
  ✔  Looking for Helm...
  ✔  Adding Kubeshop repository...
  ✔  Fetching the latest charts...
  ✔  Installing Kusk Gateway
  ✔  Installing Envoy Fleet...
  ✔  Installing Private Envoy Fleet...
  ✔  Installing Kusk API server...
  ✔  Installing Kusk Dashboard...
  •  kusk dashboard is now available. To access it run: $ kusk dashboard
```

The following command will create a Helm release named with **--name** in the namespace specified by **--namespace**.

```sh
$ kusk cluster install --name=my-release --namespace=my-namespace
...
```

The following command will install Kusk Gateway, but not the dashboard, api, or envoy-fleet.

```sh
$ kusk cluster install --no-dashboard --no-api --no-envoy-fleet
...
```

#### **Arguments**

| Flag                    | Description                                                                                                         | Required? |
|:------------------------|:--------------------------------------------------------------------------------------------------------------------|:---------:|
| `--name`                | The prefix of the name to give to the helm releases for each of the Kusk Gateway components (default: kusk-gateway). |     ❌     |
| `--namespace`/`-n`      | The namespace to install Kusk Gateway into. Will create the namespace if it doesn't exist (default: kusk-system).    |     ❌     |
| `--no-dashboard`        | When set, will not install the Kusk Dashboard.                                                              |     ❌     |
| `--no-api`              | When set, will not install the Kusk Gateway api. implies --no-dashboard.                                            |     ❌     |
| `--no-envoy-fleet`      | When set, will not install any envoy fleets.                                                                        |     ❌     |

#### **Environment Variables**

To disable analytics set following environment variable:

```
export ANALYTICS_ENABLED=false
```

or run 
```
ANALYTICS_ENABLED=false kusk cluster install
```

