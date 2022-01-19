# TLS certificates testing

This directory contains files for the generation of self-signed certificates and testing Envoys TLS.

The order of testing is following:

1. Generate CA and Intermediate CA, server certificate from CA (usual cert) and server certificate from the Intermediate.

    ```shell
    make destroycerts
    make all
    cd intermediate
    make destroycerts
    make all
    ```

    The resulting files will be:

    certs/ca.pem - Root CA certificate.

    certs/intermediate/ca.pem - the Intermediate CA, signed by Root CA.

    certs/server.pem, certs/server.key - the certificate and the key for domains www.example.com, second.example.com.

    certs/intermediate/server.pem, certs/intermediate/server.key - the certificate and the key signed by the Intermediate CA for domains www.example2.com, second.example2.com, localhost, localhost.localdomain, 127.0.0.1.

2. Prepare and deploy Kubernetes secrets with such certificates.

    In certs, prepare and deploy the certificate, signed by Root CA (www.example.com):

    ```shell
    make print
    ```

    Select and Copy the certificate part in the bottom beginning with ---BEGIN CERTIFICATE --- and ending with -- END CERTIFICATE --- including these lines.
    Save these lines into s.pem file.

    Create the kubernetes secret in default namespace.

    ```shell
    kubectl create secret tls server-cert --cert=s.pem --key=server.key
    ```

    Second part - prepare and deploy the certificate, signed by the Intermediate CA.

    NOTE: All work in this section should be done in the intermediate directory.

    ```shell
    cd intermediate
    make print
    ```

    Select and Copy the certificate part in the bottom beginning with ---BEGIN CERTIFICATE --- and ending with -- END CERTIFICATE --- including these lines.
    Save these lines into s.pem file.

    Prepend the Intermediate CA certificate to the end of this file, forming the certifite chain:

    ```shell
    cat ca.pem >> s.pem
    ```

    Create the kubernetes secret in default namespace.

    ```shell
    kubectl create secret tls server-int-cert --cert=s.pem --key=server.key
    ```

3. Deplying EnvoyFleet TLS configuration

    Prepare EnvoyFleet file (see config/samples/) that has the following lines in the end:

    ```yaml
    ...
    tls:
        cipherSuites:
           - ECDHE-ECDSA-AES128-SHA
           - ECDHE-RSA-AES128-SHA
           - AES128-GCM-SHA256
        tlsMinimumProtocolVersion: TLSv1_2
        tlsMaximumProtocolVersion: TLSv1_3
        tlsSecrets:
        - secretRef: server-int-cert
            namespace: default
        - secretRef: server-cert
            namespace: default
    ```

    Deploy the file, create the port-forwarding to 19000 port of Envoy and verify that listener configuration was updated with the additional 2 filter_chain_matche with different hostnames in the filter.

4. Testing

    Test that server cert works verifying the certificates with main CA.
    The intermediate certificate should work since Envoy supplies the Intermediate CA cert as a second entry in the certificate.

    Curl should connect successfully.

    ```shell
    export EXTERNAL_IP=<your balancer ip>
    curl --resolve www.example.com:443:$EXTERNAL_IP  https://www.example.com/ --cacert development/certs/ca.pem -v
    curl --resolve www.example2.com:443:$EXTERNAL_IP  https://www.example2.com/ --cacert development/certs/ca.pem -v
    ```
