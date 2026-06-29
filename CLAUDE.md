# cert-manager-webhook-binarylane

Kubernetes cert-manager ACME DNS01 solver webhook for BinaryLane DNS.

## Tech Stack
- Go 1.22+
- Kubernetes / cert-manager
- Helm (deployment)
- Reference: cert-manager/webhook-example, vultr/cert-manager-webhook-vultr, go-acme/lego

## Conventions
- Standard Go conventions: gofmt, golangci-lint
- Conventional Commits (feat:, fix:, chore:, docs:)
- Table-driven tests with testify
- Error handling: standard Go with wrapped errors
- cert-manager webhook patterns

## Babysitter

This project uses babysitter for AI-assisted development orchestration.

### Recommended Processes
- `tdd-quality-convergence` — Primary development process (test-first with quality gates)
- `gsd/new-project` — Greenfield project setup with architecture planning
- `gsd/execute` — Feature implementation workflow
- `gsd/verify` — Quality verification gates

### Recommended Skills
- `tdd` — Test-driven development for Go
- `verification` — Evidence-based verification before completion
- `debugging` — Systematic debugging for Go/Kubernetes issues
- `code-review` — Deep code review for solo developer self-review

### Methodology
**top-down** — Decompose features from high-level design through implementation with quality gates.

### How to Invoke
```bash
# Start a babysitter run
babysitter run:create --process-id tdd-quality-convergence --entry .a5c/processes/tdd-quality-convergence.js#process --prompt "..." --harness oh-my-pi --json

# Resume a run
babysitter session:resume --session-id <id> --run-id <runId> --runs-dir .a5c/runs --json
```

### Configuration
- Profile: `.a5c/project-profile.json`
- Runs: `.a5c/runs/`
- Local git only — no CI/CD integration configured
