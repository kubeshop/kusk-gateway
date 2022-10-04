# Getting Started

In this section, you will:
1. Install Kusk CLI in your development environment and install Kusk Gateway in your cluster 
3. Deploy an API to Kusk Gateway with mocking enabled
4. Deploy a sample application and connect it to Kusk Gateway

### **1. Install Kusk CLI** 

To install Kusk CLI, you will need the following tools available in your terminal:

- [helm](https://helm.sh/docs/intro/install/) command-line tool
- [kubectl](https://kubernetes.io/docs/tasks/tools/) command-line tool

**MacOS**
```sh
brew install kubeshop/kusk/kusk
```

**Linux**
```sh
curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/cmd/kusk/scripts/install.sh | bash
```

**Windows (go binary needed)**
```sh
go install -x github.com/kubeshop/kusk-gateway/cmd/kusk@latest
```

### **2. Install Kusk Gateway in your cluster**

Use the Kusk CLIs [install command](./reference/cli/install-cmd.md) to install Kusk Gateway components in your cluster. 

```sh
kusk cluster install
```

Now that you've installed Kusk Gateway, let's have a look at how you can use OpenAPI to configure the operational and functional parts of your API.

### **3. Create a sample OpenAPI definition**

Kusk Gateway relies on [OpenAPI](https://www.openapis.org/) (f.k.a Swagger) to define your APIs and configure the gateway, all in one place, using the `x-kusk` extension.

```yaml title="openapi.yaml"
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk: # <-- Section that configures Kusk Gateway
  mocking: # <-- Enables returning mock (fake) results
    enabled: true
paths:
  /hello:
    get:
      responses:
        '200':
          description: A simple hello world!
          content:
            text/plain:
              schema:
                type: string
              example: Hello from a mocked response!
```

This approach of deploying an API and mocking it fits great in an **Design-First approach**, allowing, for example, frontend teams to work at the same time as the backend teams as the frontend team can start developing by using the mock results provided by Kusk Gateway. 

### **4. Deploy the API**

```sh
kusk deploy -i openapi.yaml
```

**Given we have enabled gateway-level mocks**, we don't need to have any applications deployed to test the API. Kusk Gateway will provide with mock responses.

Get the IP of Kusk's LoadBalancer with: 

```sh
$ kusk ip

10.12.34.56
```

```sh
$ curl 10.12.34.56/hello

Hello from a mocked response!
```
### **6. Deploy an application**

Once you have created and API and mocked its results using Kusk Gateway, the next step is to deploy an applications and connect it to Kusk Gateway.

Deploy the following `hello-world` Deployment:

```sh
kubectl create deployment hello-world --image=kubeshop/kusk-hello-world:v1.0.0

kubectl expose deployment hello-world --name hello-world-svc --port=8080
```
### **7. Update the OpenAPI definition to connect the application to Kusk Gateway**

First, you will need to stop the mocking of the API. Delete the `mocking` section from the `openapi.yaml` file: 

```diff
...
- mocking: 
-  enabled: true
...
```

Add the `upstream` policy to the top of the `x-kusk` section of the `openapi.yaml` file, with the details of the service we just created:

```yaml
x-kusk:
 upstream:
  service:
    name: hello-world-svc
    namespace: default
    port: 8080
```

The resulting file should look like this: 
```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  upstream:
    service:
      name: hello-world-svc
      namespace: default
      port: 8080
paths:
  /hello:
    get:
      responses:
        '200':
          description: A simple hello world!
          content:
            text/plain:
              schema:
                type: string
              example: Hello from a mocked response!
```



### **8. Apply the new changes**

```
kusk deploy -i openapi.yaml
```

### **9. Test the deploy application**

```
$ curl 100.12.34.56/hello
Hello from an implemented service!
```

This response is served from the deployed application. Now you have successfully deployed an application to Kusk Gateway! 

## Next Steps

The approach from this "Getting Started" section of the documentation follows a [design-first](https://kubeshop.io/blog/from-design-first-to-automated-deployment-with-openapi) approach where you deployed the API first, mocked the API later deployed an application and connected them to Kusk Gateway.

Check out the [available OpenAPI extensions](./guides/working-with-extension.md) to see all the features that you can enable in your gateway through OpenAPI. And, if you want, connect with us on [Discord](https://discord.gg/6zupCZFQbe) to tell us about your experience!