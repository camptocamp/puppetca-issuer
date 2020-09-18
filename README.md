# Puppet Certificate Authority Issuer

This is a Cert-Manager issuer for the Puppet CA.


# cert-manager

cert-manager manages certificates in Kubernetes environment (among others) and keeps track of renewal requirements (https://cert-manager.io/). It supports various in-built issuers that issue the certificates to be managed by cert-manager.

# Puppet CA Issuer

This project plugs into cert-manager as an external issuer that talks to the Puppet CA to get certificates issued for your Kubernetes environment.

# Setup

Install cert-manager first (https://cert-manager.io/docs/installation/kubernetes/), version 0.16.1 or later.

Clone this repo and perform following steps to install controller:

```
# make build
# make docker
# make deploy
```

Create secret that holds Puppet CA credentials:

```
# cat secret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: puppetca-credentials
  namespace: puppetca-issuer-system
data:
  url: <base64 encoding of url to the PuppetCA>
  cert: <base64 encoding of certificate to access the PuppetCA>
  key: <base64 encoding of private key to access the PuppetCA>
  cacert: <base64 encoding of CA certificate of the PuppetCA>
```

 _Note_: While generating base64 encoding of above fields, ensure there is no newline character included in the encoded string. For example, following command could be used:
 
 ```
 echo -n "<access key>" | base64
 ```

Apply configuration to create secret: 

```  
# kubectl apply -f secret.yaml
```

Create resource PuppetCAIssuer for our controller:

```
# cat issuer.yaml

apiVersion: certmanager.puppetca/v1alpha2
kind: PuppetCAIssuer
metadata:
  name: puppetca-issuer
  namespace: puppetca-issuer-system
spec:
  provisioner:
    name: puppetca-credentials
    url:
      key: url
    cert:
      key: cert
    key:
      key: key
    cacert:
      key: cacert
```

Apply this configuration:

```
# kubectl apply -f issuer.yaml

# kubectl describe PuppetCAIssuer -n puppetca-issuer-system

Name:         puppetca-issuer
Namespace:    puppetca-issuer-system
Labels:       <none>
Annotations:  API Version:  certmanager.puppetca/v1alpha2
Kind:         PuppetCAIssuer
...
Spec:
  Provisioner:
    Url:
      key: url
    Cert:
      key: cert
    Key:
      key: key
    CaCert:
      key: cacert
Status:
  Conditions:
    Last Transition Time:  2020-08-31T04:34:33Z
    Message:               PuppetCAIssuer verified and ready to sign certificates
    Reason:                Verified
    Status:                True
    Type:                  Ready
Events:
  Type    Reason    Age                    From                     Message
  ----    ------    ----                   ----                     -------
  Normal  Verified  8m22s (x2 over 8m22s)  puppetca-controller      PuppetCAIssuer verified and ready to sign certificates
```

Now create certificate:

```
# cat certificate.yaml

apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: foo-puppet-cert
  namespace: puppetca-issuer-system
spec:
  # The secret name to store the signed certificate
  secretName: puppet-certificate-foo
  # Common Name
  commonName: foo.com
  # DNS SAN
  dnsAltNames:
    - localhost
    - foo.com
  issuerRef:
    group: certmanager.puppetca
    kind: PuppetCAIssuer
    name: puppetca-issuer
```

```
# kubectl apply -f certificate.yaml
# kubectl describe Certificate foo-puppet-cert -n puppetca-issuer-system

Name:         foo-puppet-cert
Namespace:    puppetca-issuer-system
Labels:       <none>
Annotations:  API Version:  cert-manager.io/v1alpha3
Kind:         Certificate
...
Spec:
  Common Name:  foo.com
  Dns Names:
    localhost
    foo.com
  Issuer Ref:
    Group:       certmanager.puppetca
    Kind:        PuppetCAIssuer
    Name:        puppetca-issuer
  Secret Name:   puppet-certificate-foo
Status:
  Conditions:
    Last Transition Time:  2020-08-18T04:34:48Z
    Message:               Certificate is up to date and has not expired
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2020-08-19T04:34:45Z
  Not Before:              2020-08-18T03:34:45Z
  Renewal Time:            2020-08-19T03:34:45Z
  Revision:                1
Events:
  Type    Reason     Age    From          Message
  ----    ------     ----   ----          -------
  Normal  Issuing    6m1s   cert-manager  Issuing certificate as Secret does not exist
  Normal  Generated  6m     cert-manager  Stored new private key in temporary Secret resource "backend-puppetca-7m9sx"
  Normal  Requested  6m     cert-manager  Created new CertificateRequest resource "backend-puppetca-m2gz5"
  Normal  Issuing    5m51s  cert-manager  The certificate has been successfully issued
```

Check certificate and private key are present in secrets:                                             

```
# kubectl describe secrets puppet-certificate-foo -n puppetca-issuer-system   

Name:         foo-puppet-cert
Namespace:    puppetca-issuer-system
Labels:       <none>
Annotations:  cert-manager.io/alt-names: localhost,foo.com
              cert-manager.io/certificate-name: foo-puppet-cert
              cert-manager.io/common-name: foo.com
              cert-manager.io/issuer-kind: PuppetCAIssuer
              cert-manager.io/issuer-name: puppetca-issuer
              cert-manager.io/uri-sans:

Type:  kubernetes.io/tls

Data
====
tls.key:  xxxx bytes
tls.crt:  yyyy bytes
```
