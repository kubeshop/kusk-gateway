# Using Cert Manager with Kusk Gateway
Cert Manager and Kusk Gateway work well together. Cert Manager is a way to easily issue and automatically rotate certificates.

Kusk Gateway can be instructed to use those certificates by defining them in your `EnvoyFleet`.

Kusk Gateway will also watch your certificates for updates and will reload the EnvoyFleet config automatically
without the need for any manual actions.

## **Install Cert Manager**
Cert Manager can be installed using the following command which uses the default configuration.

`kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml`

For other installation methods, refer to Cert Manager's installation [document](https://cert-manager.io/docs/installation/).

## **Issue a Certificate**
To issue a certificate, we need to define an Issuer or ClusterIssuer. This defines which Certificate Authority Cert Manager will be used to issue the certificate.

For demonstration purposes, let's use a simple self-signed certificate issuer:

```yaml
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: test-selfsigned
  namespace: default
spec:
  selfSigned: {}
EOF
```

We can now issue a self-signed certificate using this issuer:

```yaml
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: selfsigned-cert
  namespace: default
spec:
  dnsNames:
    - example.com
  secretName: selfsigned-cert-tls
  issuerRef:
    name: test-selfsigned
EOF
```

Cert manager will react to the creation of this Certificate resource and produce for us a Kubernetes secret
that contains the certificate we can use in Kusk Gateway to secure your endpoints with TLS (Transport Layer Security).

Fetch the list of secrets to confirm that our certificate was created:
```
❯ kubectl get secrets
NAME                  TYPE                                  DATA   AGE
...
selfsigned-cert-tls   kubernetes.io/tls                     3      103s
```
Describe the secret:

```
❯ kubectl describe secret selfsigned-cert-tls
Name:         selfsigned-cert-tls
Namespace:    default
Labels:       <none>
Annotations:  cert-manager.io/alt-names: example.com
              cert-manager.io/certificate-name: selfsigned-cert
              cert-manager.io/common-name:
              cert-manager.io/ip-sans:
              cert-manager.io/issuer-group:
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: test-selfsigned
              cert-manager.io/uri-sans:

Type:  kubernetes.io/tls

Data
====
tls.crt:  1021 bytes
tls.key:  1679 bytes
ca.crt:   1021 bytes
```

## **Using the Certificate in Kusk Gateway**
In your EnvoyFleet definition, add the following TLS settings into the spec field:

```
apiVersion: gateway.kusk.io/v1alpha1
kind: EnvoyFleet
metadata:
  name: default
spec:
    ...
    tls:
     tlsSecrets:
       - secretRef: selfsigned-cert-tls
         namespace: default
```

We defined the hostname in the certificate as example.com, therefore, your API will need to have this host in the hosts array of the x-kusk extension to make use of the secret.

We can confirm the details of the certificate using OpenSSL:
```shell
echo | \
    openssl s_client -servername example.com -connect example.com:443 2>/dev/null | \
    openssl x509 -text
```

For this example, you will need to add example.com to your `/etc/hosts` file pointing at the envoy service public IP running in the cluster.

## **Rotating Secrets**
Kusk Gateway will watch for updates to your secrets in any of its EnvoyFleets and update the config to use them
automatically, without any manual intervention needed

We can force a certificate rotation using cmctl and then check that Kusk Gateway does register the change
and update the config accordingly.

You will need to have [cmctl installed](https://cert-manager.io/docs/usage/cmctl/#installation).

Now we can issue a `renew` command:

```
❯ cmctl renew selfsigned-cert
Manually triggered issuance of Certificate default/selfsigned-cert
```

This will mark the named secret for manual renewal by cert-manager and it should do so relatively quickly.

Use OpenSSL again to check the updated certificate:
```shell
echo | \
    openssl s_client -servername example.com -connect example.com:443 2>/dev/null | \
    openssl x509 -text
```
