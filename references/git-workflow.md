# Git Workflow Reference

## Branch Naming

No strictly enforced convention. Use descriptive names that reflect the change:
- `feat/add-network-acl-resource`
- `fix/read-action-on-404`
- `chore/bump-go-auth0-v3`

## Commit Messages

Conventional Commits format is used across the repo:

| Prefix | When to use |
|--------|-------------|
| `feat:` | New resource, data source, or feature |
| `fix:` | Bug fix |
| `chore:` | Dependency updates, CI changes, non-functional changes |
| `docs:` | Documentation-only changes |
| `refactor:` | Internal refactoring with no behavior change |
| `test:` | Test-only changes |

All commits must be **GPG-signed** (configured in CONTRIBUTING.md). Configure with `git config commit.gpgsign true`.

## Pull Requests

The repo has a local PR template at `.github/PULL_REQUEST_TEMPLATE.md`. PRs must include:

- **Changes** — what was added, changed, or fixed
- **References** — links to related GitHub issues or Auth0 Management API documentation
- **Testing** — how the change was verified (unit tests, recorded acceptance tests, or live e2e)
- **Checklist:**
  - [ ] Tests added or updated for new/changed functionality
  - [ ] Documentation updated — `examples/` updated and `make docs` run if schemas changed
  - [ ] Correct base branch selected

PRs with an incomplete checklist are closed without review.

## Base Branch

- `main` — default; all features and fixes target this branch
- `v1` — legacy maintenance branch (for backporting only)
