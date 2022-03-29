# KGW CLI
The KGW (Kusk Gateway) CLI is a helper utility used to generate instances of API Custom Resources
from a given API spec.

Using this utility removes the need to manually embed your OpenAPI document inside the Custom Resource and not worry about whether or not you have indented your spec enough.

Another advantage of the KGW CLI is how it can be used in your CI/CD pipelines. As you iterate on your OpenAPI spec and push those changes to Git, kgw can be used to automatically generate new API resources for Kusk Gateway. For an example of this, check out our [petshop cicd example](https://github.com/kubeshop/petshop-kusk-cicd), which uses Github Actions and KGW to automate the configuration of Kusk Gateway.

For more details, check out the [Github Repo](https://github.com/kubeshop/kgw)
