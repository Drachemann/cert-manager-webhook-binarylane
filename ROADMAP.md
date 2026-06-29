# Roadmap — cert-manager-webhook-binarylane

## Phase 1: Foundation (v0.1)
- [ ] Project scaffolding: go mod init, directory structure
- [ ] BinaryLane API client (pkg/binarylane)
- [ ] DNS01 solver implementation (pkg/solver)
- [ ] Webhook server entrypoint (main.go)
- [ ] Unit tests for solver and client

## Phase 2: Integration (v0.2)
- [ ] Helm chart for deployment
- [ ] cert-manager GroupName / solver configuration
- [ ] Integration tests against BinaryLane sandbox
- [ ] Health checks and metrics

## Phase 3: Production Readiness (v0.3)
- [ ] cert-manager conformance test suite
- [ ] Error handling and retry logic
- [ ] README with installation and configuration docs
- [ ] Rate limiting and backpressure

## Phase 4: Polish (v1.0)
- [ ] Performance optimization
- [ ] Observability (structured logging, metrics)
- [ ] Security hardening
- [ ] Release automation
