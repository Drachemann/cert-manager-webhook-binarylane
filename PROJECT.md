# cert-manager-webhook-binarylane

## Vision
A production-ready Kubernetes cert-manager webhook that automates TLS certificate issuance via ACME DNS01 challenges using the BinaryLane DNS API.

## Problem
BinaryLane users running cert-manager in Kubernetes need an automated way to solve DNS01 challenges for wildcard and non-HTTP-accessible domains. This webhook bridges cert-manager's DNS01 solver interface with the BinaryLane DNS API.

## Success Criteria
- Passes cert-manager's conformance test suite
- Deployable via Helm with minimal configuration
- Handles concurrent challenge requests reliably
- Clean error handling and logging for production debugging

## Architecture
```
cert-manager → webhook HTTP server (Go) → BinaryLane DNS API
```

## Tech Stack
- Go 1.22+
- Kubernetes / cert-manager
- Helm
- Reference: cert-manager/webhook-example, vultr/cert-manager-webhook-vultr, go-acme/lego
