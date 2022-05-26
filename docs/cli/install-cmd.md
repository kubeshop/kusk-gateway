# Installing Kusk Gateway with the Kusk CLI

The `install` command will install Kusk Gateway and all its components with a single command. 
Kusk uses Helm to do this, so you will need to have [Helm installed](https://helm.sh/docs/intro/install/).

### **Kusk Gateway Components**

* **Kusk Gateway Manager** - Responsible for updating and rolling out the Envoy configuration to your Envoy Fleets as you deploy APIs and Static Routes.
* **Envoy Fleet** - Responsible for exposing and routing to your APIs and frontends.
* **Kusk Gateway API** - REST API, which is exposed by Kusk Gateway and allows you to programmatically query which APIs, Static Routes and Envoy Fleets are deployed.
* **Kusk Gateway Dashboard** - A web UI for Kusk Gateway where you can deploy APIs and see which APIs, StaticRoutes and Envoy Fleets are deployed.

#### **Examples**

The default `kusk install` command will install Kusk Gateway, a public (for your APIs) and private (for the Kusk dashboard and API)
envoy-fleet, api, and dashboard in the kusk-system namespace using Helm and using the current kubeconfig context.

```shell
$ kusk install
adding the kubeshop helm repository
done
fetching the latest charts
done
installing Kusk Gateway
done
installing Envoy Fleet
done
installing Kusk API
done
installing Kusk Dashboard
done

To access the dashboard, port forward to the envoy-fleet service that exposes it:
        $ kubectl port-forward -n            kusk-system svc/kusk-gateway-private-envoy-fleet 8080:80

Then go http://localhost:8080/
```

The following command will create a Helm release named with **--name** in the namespace specified by **--namespace**.

```shell
$ kusk install --name=my-release --namespace=my-namespace
...
```

The following command will install Kusk Gateway, but not the dashboard, api, or envoy-fleet.

```shell
$ kusk install --no-dashboard --no-api --no-envoy-fleet
...
```

#### **Arguments**

| Flag                    | Description                                                                                                         | Required? |
|:------------------------|:--------------------------------------------------------------------------------------------------------------------|:---------:|
| `--name`                | The prefix of the name to give to the helm releases for each of the kusk gateway components (default: kusk-gateway). |     ❌     |
| `--namespace`/`-n`      | The namespace to install kusk gateway into. Will create the namespace if it doesn't exist (default: kusk-system).    |     ❌     |
| `--no-dashboard`        | When set, will not install the kusk gateway dashboard.                                                              |     ❌     |
| `--no-api`              | When set, will not install the kusk gateway api. implies --no-dashboard.                                            |     ❌     |
| `--no-envoy-fleet`      | When set, will not install any envoy fleets.                                                                        |     ❌     |
