# cert-manager-webhook-binarylane

cert-manager ACME DNS01 solver webhook for [BinaryLane](https://www.binarylane.com.au/) DNS.

## Prerequisites

- Kubernetes 1.28+
- [cert-manager](https://cert-manager.io/docs/installation/) 1.14+ installed
- BinaryLane API token

## Installation

```bash
# Create secret with BinaryLane API token
kubectl create secret generic binarylane-api-credentials \
  --from-literal=api-key=YOUR_BINARYLANE_API_TOKEN \
  -n cert-manager

# Install the webhook
helm repo add cert-manager-webhook-binarylane https://drachemann.github.io/cert-manager-webhook-binarylane
helm install cert-manager-webhook-binarylane cert-manager-webhook-binarylane/cert-manager-webhook-binarylane \
  --namespace cert-manager
```

## Usage

Create a ClusterIssuer:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-binarylane
spec:
  acme:
    email: you@example.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-binarylane-account-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.binarylane.com
          solverName: binarylane
          config:
            apiKeySecretRef:
              name: binarylane-api-credentials
              key: api-key
```

Request a certificate:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com
spec:
  secretName: example-com-tls
  dnsNames:
  - example.com
  - '*.example.com'
  issuerRef:
    name: letsencrypt-binarylane
    kind: ClusterIssuer
```

## Development

```bash
go mod download
go build ./...
go test ./...
```

## License

MIT
