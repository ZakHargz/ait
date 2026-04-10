# GitHub Authentication

AIT supports authentication with private GitHub repositories using the same approach as Microsoft's APM (Agent Package Manager).

## How It Works

AIT automatically authenticates with GitHub when you have a token configured in your environment. Authentication is resolved per `(host, org)` pair using a priority chain:

1. **Per-org environment variable** — `GITHUB_APM_PAT_{ORG}` (GitHub-like hosts only)
2. **Global environment variables** — `GITHUB_APM_PAT` → `GITHUB_TOKEN` → `GH_TOKEN`
3. **No authentication** — Falls back for public repositories

## Quick Start

### For Public Repositories

No authentication needed! AIT works with public repositories without any configuration.

### For Private Repositories

Set one of these environment variables:

```bash
# Option 1: Use the token set by GitHub CLI (gh auth login)
export GH_TOKEN=$(gh auth token)

# Option 2: Set GITHUB_TOKEN (commonly used in CI/CD)
export GITHUB_TOKEN=ghp_your_personal_access_token

# Option 3: Use APM-style token
export GITHUB_APM_PAT=ghp_your_personal_access_token
```

Add to your shell profile (`.zshrc`, `.bashrc`, etc.) to make it permanent:

```bash
# Add to ~/.zshrc
export GITHUB_TOKEN=ghp_your_personal_access_token
```

## Token Priority

When multiple tokens are set, AIT uses this priority order:

| Priority | Variable | Scope | Notes |
|----------|----------|-------|-------|
| 1 | `GITHUB_APM_PAT_{ORG}` | Per-org, GitHub-like hosts | Org name uppercased, hyphens → underscores |
| 2 | `GITHUB_APM_PAT` | Any host | APM-compatible global token |
| 3 | `GITHUB_TOKEN` | Any host | Common in CI/CD (GitHub Actions) |
| 4 | `GH_TOKEN` | Any host | Set by `gh auth login` |

## Multi-Organization Setup

If you work with multiple organizations that require different tokens:

```bash
# Per-org tokens take priority
export GITHUB_APM_PAT_CONTOSO=ghp_token_for_contoso
export GITHUB_APM_PAT_FABRIKAM=ghp_token_for_fabrikam

# Fallback for other orgs
export GITHUB_TOKEN=ghp_default_token
```

**Naming rules for per-org tokens:**
- Uppercase the org name
- Replace hyphens with underscores
- Example: `contoso-microsoft` → `GITHUB_APM_PAT_CONTOSO_MICROSOFT`

## Creating a Personal Access Token (PAT)

### Classic PAT (Recommended for Multi-Org Access)

1. Go to [github.com/settings/tokens](https://github.com/settings/tokens)
2. Click "Generate new token" → "Generate new token (classic)"
3. Give it a descriptive name (e.g., "AIT Package Manager")
4. Set expiration (recommend: 90 days or no expiration for personal use)
5. Select scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `read:org` (Read org and team membership)
6. Click "Generate token"
7. Copy the token (starts with `ghp_`) and save it securely

### Fine-Grained PAT (Organization-Scoped)

Fine-grained PATs are scoped to a **single organization**. If you need access to multiple orgs, use either:
- Multiple fine-grained PATs (one per org using per-org env vars)
- A classic PAT (works across all orgs you're a member of)
- GitHub CLI authentication (`gh auth login`)

For fine-grained PATs:

1. Go to [github.com/settings/personal-access-tokens/new](https://github.com/settings/personal-access-tokens/new)
2. Set **Resource owner** to the organization (NOT your user account)
3. Configure:
   - **Repository access**: "All repositories" or select specific ones
   - **Permissions**:
     - `Contents`: Read
     - `Metadata`: Read
4. Click "Generate token"
5. Copy the token (starts with `github_pat_`)

**Important:** A user-scoped fine-grained PAT cannot access organization repositories, even if you're a member.

## GitHub Enterprise

### GitHub Enterprise Server (GHES)

Set `GITHUB_HOST` to your enterprise instance:

```bash
export GITHUB_HOST=github.company.com
export GITHUB_APM_PAT=ghp_enterprise_token
```

Now bare package names resolve against your enterprise instance:

```bash
ait install myorg/internal-package  # → github.company.com/myorg/internal-package
```

You can still access public GitHub packages using the full hostname:

```bash
ait install github.com/public/package  # → github.com/public/package
```

### GitHub Enterprise Cloud (`*.ghe.com`)

These instances are always authentication-required:

```bash
export GITHUB_APM_PAT=ghp_enterprise_token
ait install myenterprise.ghe.com/platform/standards
```

### Enterprise Managed Users (EMU)

EMU tokens use standard prefixes (`ghp_` or `github_pat_`) and are enterprise-scoped. They **cannot** access public repos on github.com.

For mixed enterprise and public packages:

```bash
# Enterprise org (EMU)
export GITHUB_APM_PAT_CONTOSO_MICROSOFT=github_pat_enterprise_token

# Public repos work without auth (or set a different token if needed)
export GITHUB_TOKEN=ghp_public_token
```

## Using GitHub CLI

The easiest zero-config approach is using GitHub CLI:

```bash
# Login with GitHub CLI
gh auth login

# Verify authentication
gh auth status

# Now AIT will use the gh token automatically via GH_TOKEN
ait install myorg/private-package
```

The GitHub CLI creates an OAuth token that inherits your full organization membership, avoiding the scoping issues of fine-grained PATs.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Install AIT packages
  run: ait install
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

The `GITHUB_TOKEN` secret is automatically provided by GitHub Actions with access to the current repository.

For private packages in other organizations, create a PAT and store it as a repository secret:

```yaml
- name: Install AIT packages
  run: ait install
  env:
    GITHUB_APM_PAT: ${{ secrets.GITHUB_APM_PAT }}
```

### Other CI Systems

Store your GitHub token as a secret in your CI system and expose it as an environment variable:

```bash
export GITHUB_TOKEN=$GITHUB_PAT_SECRET
ait install
```

## Troubleshooting

### "failed to clone repository: authentication required"

- Verify your token is set: `echo $GITHUB_TOKEN`
- Check token has `repo` scope
- For organization repos, ensure SSO is authorized (Settings → Tokens → Configure SSO)

### "failed to clone repository: repository not found"

This error can mean:
1. The repository doesn't exist
2. You don't have access to the repository
3. Your token isn't authorized (check SSO for org repos)
4. Fine-grained PAT isn't scoped to the right organization

### SSO-Protected Organizations

If your organization uses SAML SSO, you must authorize your token:

1. Go to [github.com/settings/tokens](https://github.com/settings/tokens)
2. Find your token
3. Click "Configure SSO"
4. Click "Authorize" next to your organization

### Rate Limits

Without authentication, GitHub limits API requests. If you hit rate limits on public repos:

```bash
export GITHUB_TOKEN=ghp_any_valid_token
```

Even a token without special scopes helps avoid rate limits.

## Security Best Practices

1. **Never commit tokens** — Add tokens to `.gitignore` and use environment variables
2. **Use fine-grained PATs** when possible — Limit scope to specific repositories
3. **Rotate tokens regularly** — Set expiration dates and renew periodically
4. **Use different tokens** for different purposes — Don't share personal tokens across systems
5. **Revoke unused tokens** — Clean up old tokens at [github.com/settings/tokens](https://github.com/settings/tokens)

## Environment Variables Reference

| Variable | Purpose | Example |
|----------|---------|---------|
| `GITHUB_TOKEN` | Global GitHub token | `ghp_abc123...` |
| `GH_TOKEN` | Token from GitHub CLI | Set automatically by `gh auth login` |
| `GITHUB_APM_PAT` | APM-compatible global token | `ghp_abc123...` |
| `GITHUB_APM_PAT_{ORG}` | Per-organization token | `GITHUB_APM_PAT_MYORG=ghp_...` |
| `GITHUB_HOST` | GitHub Enterprise Server host | `github.company.com` |

## Examples

### Personal Use with Multiple Orgs

```bash
# ~/.zshrc
export GITHUB_TOKEN=$(gh auth token)  # Easiest: use GitHub CLI
```

Or with per-org tokens:

```bash
# ~/.zshrc
export GITHUB_APM_PAT_WORKORG=ghp_work_token
export GITHUB_APM_PAT_PERSONALORG=ghp_personal_token
export GITHUB_TOKEN=ghp_fallback_token
```

### Enterprise + Public Repos

```bash
# ~/.zshrc
export GITHUB_HOST=github.company.com
export GITHUB_APM_PAT=ghp_enterprise_token
```

Access enterprise packages:
```bash
ait install myorg/internal-package  # Uses enterprise token
```

Access public packages:
```bash
ait install github.com/public/package  # Works without auth or uses token if set
```

### CI/CD with Private Packages

```yaml
# .github/workflows/ci.yml
name: CI
on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install AIT
        run: curl -sSL https://install.ait.sh | sh
      
      - name: Install packages
        run: ait install
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Run tests
        run: make test
```

## Comparison with APM

AIT's authentication is **fully compatible** with APM's approach:

- ✅ Same environment variable names
- ✅ Same priority order
- ✅ Same per-org token naming convention
- ✅ Same support for GitHub Enterprise

If you're already using APM, your existing authentication setup works with AIT without any changes.
