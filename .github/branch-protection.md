# Branch Protection Rules Configuration

This document describes the branch protection rules for the Vortyx project GitHub repository.

## Overview

The Vortyx project uses Git Flow methodology with protected branches to ensure code quality and maintain production stability.

## Protected Branches

| Branch | Purpose | Protection Level | Required Checks |
|--------|---------|------------------|-----------------|
| `main` | Production code | Strict | CI, Dependency Review |
| `dev` | Development code | Moderate | CI |
| `release/*` | Release branches | Moderate | CI |
| `hotfix/*` | Hotfix branches | None | CI (for merge) |

## Main Branch Protection Rules

### Required Status Checks
- ✅ `CI` - Must pass before merging
- ✅ `Dependency Review` - Must pass before merging
- All required checks must pass before merging

### Protection Configuration
1. **Require pull request reviews before merging**
   - Required number of approvals: 1
   - Dismiss stale reviews when new commits are pushed
   - Require review from code owners (optional)

2. **Require status checks to pass before merging**
   - Require branches to be up to date before merging
   - Require all required status checks to pass

3. **Require conversation resolution**
   - All comments on the review thread must be resolved

4. **Include administrators**
   - Apply rules to administrators

5. **Restrict who can push**
   - Allow only specific team members to push directly

## Dev Branch Protection Rules

### Required Status Checks
- ✅ `CI` - Must pass before merging

### Protection Configuration
1. **Require pull request reviews before merging**
   - Required number of approvals: 1
   - Dismiss stale reviews when new commits are pushed

2. **Require status checks to pass before merging**
   - Require branches to be up to date before merging

3. **Include administrators**
   - Apply rules to administrators

## Setup Instructions

### Method 1: GitHub Web Interface

1. Navigate to **Repository Settings** → **Branches**
2. Click **Add branch protection rule**
3. For `main` branch:
   - Branch name pattern: `main`
   - Check "Require pull request reviews before merging"
   - Check "Require status checks to pass before merging"
   - Check "Require branches to be up to date before merging"
   - Check "Include administrators"
4. For `dev` branch:
   - Branch name pattern: `dev`
   - Check "Require pull request reviews before merging"
   - Check "Require status checks to pass before merging"

### Method 2: GitHub CLI

```bash
# Install GitHub CLI if not already installed
# Windows: winget install GitHub.GitHubCLI
# macOS: brew install gh

# Authenticate
gh auth login

# Protect main branch
gh api repos/{owner}/{repo}/protected_branches/main -X PUT \
  -f required_status_checks='{"strict":true,"contexts":["ci"]}' \
  -f required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  -f enforce_admins=true \
  -f allow_force_pushes=false \
  -f allow_deletions=false

# Protect dev branch
gh api repos/{owner}/{repo}/protected_branches/dev -X PUT \
  -f required_status_checks='{"strict":true,"contexts":["ci"]}' \
  -f required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  -f enforce_admins=true \
  -f allow_force_pushs=false \
  -f allow_deletions=false
```

### Method 3: Terraform (Infrastructure as Code)

```hcl
# main.tf
resource "github_branch_protection" "main" {
  repository_id  = github_repository.vortyx.id
  branch         = "main"
  required_status_checks = [
    "ci"
  ]
  required_pull_request_reviews {
    required_approving_review_count = 1
    dismiss_stale_reviews           = true
  }
  enforce_admins = true
}

resource "github_branch_protection" "dev" {
  repository_id  = github_repository.vortyx.id
  branch         = "dev"
  required_status_checks = [
    "ci"
  ]
  required_pull_request_reviews {
    required_approving_review_count = 1
    dismiss_stale_reviews           = true
  }
  enforce_admins = true
}
```

## Auto-Deletion Settings

Enable auto-delete head branches when PRs are merged:
- Go to **Repository Settings** → **General**
- Check "Automatically delete head branches"

## Required Secrets and Variables

### Repository Secrets (Settings → Secrets and variables → Actions)

| Secret | Description | Required |
|--------|-------------|----------|
| `SNYK_TOKEN` | Snyk security scanning token | Optional |
| `PRODUCTION_DATABASE_URL` | Production database connection | Yes |
| `STAGING_DATABASE_URL` | Staging database connection | Yes |
| `PRODUCTION_ZITADEL_URL` | Production Zitadel instance URL | Yes |
| `STAGING_ZITADEL_URL` | Staging Zitadel instance URL | Yes |

### Repository Variables (Settings → Secrets and variables → Actions)

| Variable | Description | Required |
|----------|-------------|----------|
| `PRODUCTION_API_URL` | Production API endpoint | Yes |
| `STAGING_API_URL` | Staging API endpoint | Yes |

## Workflow Permissions

Ensure GitHub Actions has proper permissions:
1. Go to **Repository Settings** → **Actions** → **General**
2. Set "Workflow permissions" to "Read and write"
3. Check "Allow GitHub Actions to create and approve pull requests"

## Verification

After setup, verify protection rules:
```bash
# List protected branches
gh api repos/{owner}/{repo}/protected_branches

# Get main branch protection
gh api repos/{owner}/{repo}/protected_branches/main
```
