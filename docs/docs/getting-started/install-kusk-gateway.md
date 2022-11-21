# 3. Install Kusk Gateway

Kusk Gateway can be installed in the Kubernetes cluster using Kusk CLI, which was installed in the first step.

This will install Kusk's Ingress Controller and additional components to run Kusk in your cluster.

```sh 
kusk cluster install
```
```sh title="Expected output:"
🚀 Installing Kusk in your cluster
  ✔  Installing Kusk Gateway
  ✔  Installing Envoyfleet
  ✔  Installing API Server
  ✔  Installing Dashboard

🎉 Installation complete

💡 Access the dashboard by using the following command
👉 kusk dashboard

💡 Deploy your first API
👉 kusk deploy -i <path or url to your api definition>

💡 Access Help and useful examples to help get you started
👉 kusk --help
```