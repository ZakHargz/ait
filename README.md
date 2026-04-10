# AIT - AI Toolkit Package Manager

> **npm for AI agents, skills, and prompts** - Manage AI tooling across OpenCode, Cursor, Claude Desktop, and GitHub Copilot with a single command.

[![Version](https://img.shields.io/badge/version-0.8.0-blue.svg)](https://github.com/ZakHargz/ait/releases)
[![Tests](https://img.shields.io/badge/tests-95%20passing-brightgreen.svg)](https://github.com/ZakHargz/ait/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**AIT** is a CLI package manager that brings the npm/pip experience to AI development. Define your AI agents, skills, and prompts in `ait.yml`, and AIT handles installation, versioning, and synchronization across all your AI tools.

> **🔄 Aligning with Standards**: AIT actively aligns with [Microsoft's Agent Package Manager (APM)](https://microsoft.github.io/apm/), [AGENTS.md](https://agents.md), [Agent Skills](https://agentskills.io), and [Model Context Protocol (MCP)](https://modelcontextprotocol.io) for maximum compatibility across the AI ecosystem.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Getting Started Guide](#getting-started-guide)
- [How It Works](#how-it-works)
- [Commands Reference](#commands-reference)
- [Creating an AI Marketplace Monorepo](#creating-an-ai-marketplace-monorepo)
- [Developer Setup](#developer-setup)
- [Use Cases](#use-cases)
- [Supported AI Tools](#supported-ai-tools)
- [Package Format](#package-format)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

---

## Features

### Core Capabilities

- 📦 **Declarative Dependencies** - Define all AI tools in one `ait.yml` file
- 🚀 **Project-Level Installation** - Auto-detected by AI tools (no manual setup)
- 🔄 **Multi-Tool Sync** - Install to OpenCode, Cursor, Claude, GitHub Copilot simultaneously
- 👥 **Team Sharing** - Commit native files to git for zero-config team collaboration
- 🌲 **Dependency Resolution** - Automatic transitive dependency handling
- 📌 **Lock Files** - Reproducible installations with `ait.lock`
- 🏷️ **Semantic Versioning** - Version constraints (`^1.0.0`, `~2.1.0`, or latest)
- 🌍 **Multiple Sources** - GitHub, GitLab, generic Git, or local packages
- 🔐 **Private Repositories** - GitHub token authentication support
- 💾 **Smart Caching** - Faster installations with local cache

### Developer Experience

- 🩺 **Health Checks** - `ait doctor` validates your setup
- 📊 **Update Detection** - `ait outdated` shows available updates
- 🔍 **Package Discovery** - List installed packages across all tools
- 🛠️ **Developer Friendly** - Built in Go, comprehensive tests, CI/CD ready

---

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        AIT Package Manager                       │
│                                                                   │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   ait.yml    │───▶│     CLI      │───▶│   ait.lock   │      │
│  │  (Manifest)  │    │   Commands   │    │  (Lockfile)  │      │
│  └──────────────┘    └──────┬───────┘    └──────────────┘      │
│                              │                                    │
│                    ┌─────────┼─────────┐                         │
│                    │         │         │                         │
│              ┌─────▼───┐ ┌──▼────┐ ┌──▼────┐                   │
│              │ Sources │ │Resolver│ │Adapters│                  │
│              │ Package │ │ Deps   │ │ Tools  │                  │
│              └─────┬───┘ └───────┘ └───┬────┘                   │
│                    │                     │                        │
└────────────────────┼─────────────────────┼────────────────────────┘
                     │                     │
         ┌───────────▼──────────┐   ┌─────▼──────────────────────┐
         │   Remote Sources     │   │   AI Tool Installations    │
         │                      │   │                            │
         │  • GitHub repos      │   │  • .cursorrules (Cursor)   │
         │  • GitLab repos      │   │  • .github/agents/         │
         │  • Git repositories  │   │    (GitHub Copilot)        │
         │  • Local packages    │   │  • .opencode/ (OpenCode)   │
         │                      │   │  • ~/.claude/ (Claude)     │
         └──────────────────────┘   └────────────────────────────┘
```

### Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Internal Components                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  CLI Layer (cmd/ait/, internal/cli/)                            │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ init │ install │ list │ update │ uninstall │ sync │        │ │
│  │ doctor │ outdated │ generate                                │ │
│  └───────────────────────────┬────────────────────────────────┘ │
│                              │                                    │
│  Core Logic                  │                                    │
│  ┌───────────────────────────▼────────────────────────────────┐ │
│  │                                                             │ │
│  │  Config (internal/config/)                                 │ │
│  │  ├─ Manifest (ait.yml parser)                             │ │
│  │  ├─ Lockfile (ait.lock handler)                           │ │
│  │  └─ Package Metadata (package.yml)                        │ │
│  │                                                             │ │
│  │  Resolver (internal/resolver/)                             │ │
│  │  ├─ Dependency graph builder                              │ │
│  │  ├─ Version constraint solver                             │ │
│  │  └─ Transitive dependency handler                         │ │
│  │                                                             │ │
│  │  Sources (internal/sources/)                               │ │
│  │  ├─ GitSource (clone, fetch, checkout)                    │ │
│  │  ├─ LocalSource (filesystem packages)                     │ │
│  │  ├─ Auth (GitHub token handling)                          │ │
│  │  └─ Package spec parser                                   │ │
│  │                                                             │ │
│  │  Adapters (internal/adapters/)                             │ │
│  │  ├─ BaseAdapter (shared install/uninstall logic)          │ │
│  │  ├─ OpenCodeAdapter (~/.config/opencode)                  │ │
│  │  ├─ CursorAdapter (~/Library/.../Cursor)                  │ │
│  │  ├─ ClaudeAdapter (~/.claude)                             │ │
│  │  └─ ProjectRootAdapter (.cursorrules, .github/agents/)    │ │
│  │                                                             │ │
│  │  Packages (internal/packages/)                             │ │
│  │  └─ Package types (Agent, Skill, Prompt, MCP)             │ │
│  │                                                             │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. User runs: ait install org/repo/agents/reviewer@1.0.0
                              │
                              ▼
2. CLI parses command ─────▶ install.go
                              │
                              ▼
3. Load/create ait.yml ────▶ config.Manifest
                              │
                              ▼
4. Parse package spec ─────▶ sources.ParsePackageSpec()
   "org/repo/agents/reviewer@1.0.0"
   ├─ Type: github
   ├─ Repo: org/repo
   ├─ Path: agents/reviewer
   └─ Version: 1.0.0
                              │
                              ▼
5. Resolve dependencies ───▶ resolver.Resolve()
   ├─ Fetch from GitHub
   ├─ Read package.yml
   ├─ Resolve transitive deps
   └─ Build dependency graph
                              │
                              ▼
6. Fetch packages ─────────▶ sources.GitSource.Fetch()
   ├─ Clone/update repository
   ├─ Checkout version tag
   └─ Return package metadata
                              │
                              ▼
7. Install to targets ─────▶ adapters.Install()
   ├─ ProjectRootAdapter:
   │  ├─ .cursorrules
   │  ├─ .github/agents/reviewer.agent.md
   │  └─ .opencode/agents/reviewer/AGENT.md
   └─ Or global adapters (OpenCode, Cursor, Claude)
                              │
                              ▼
8. Update lockfile ────────▶ ait.lock (reproducible installs)
                              │
                              ▼
9. Success! ✓
```

---

## Quick Start

### 5-Minute Setup

```bash
# 1. Install AIT
brew tap zakhargz/ait
brew install ait

# 2. Verify installation
ait --version

# 3. Check your setup
ait doctor

# 4. Initialize a project
mkdir my-ai-project && cd my-ai-project
ait init --defaults

# 5. Install an agent
ait install apex-ai/agents/code-reviewer

# 6. Verify installation
ait list
```

**That's it!** Your AI agent is now available in Cursor, GitHub Copilot, and OpenCode.

---

## Installation

### Using Homebrew (Recommended)

```bash
# Add the AIT tap
brew tap zakhargz/ait

# Install AIT
brew install ait

# Verify installation
ait --version
# Output: ait version 0.8.0
```

### From Source

```bash
# Clone the repository
git clone https://github.com/ZakHargz/ait
cd ait

# Build
make build

# Install globally
sudo cp bin/ait /usr/local/bin/

# Or use locally
./bin/ait --version
```

### System Requirements

- **Go 1.23+** (for building from source only)
- **Git** (required for package fetching)
- **One or more AI tools**: OpenCode, Cursor, Claude Desktop, or GitHub Copilot

---

## Getting Started Guide

### For Individual Developers

#### Step 1: Set Up Your Environment

```bash
# Check if AIT is working
ait doctor

# Output shows:
# ✓ Git installation: git version 2.53.0
# ✓ AI tools detection: Found: [opencode cursor claude]
# ✓ GitHub authentication: GitHub token configured
```

#### Step 2: Create Your First Project

```bash
# Create a new project directory
mkdir my-coding-assistant
cd my-coding-assistant

# Initialize AIT
ait init

# This creates ait.yml:
# name: my-coding-assistant
# version: 1.0.0
# dependencies: []
```

#### Step 3: Install AI Agents

```bash
# Install a code review agent
ait install apex-ai/agents/code-reviewer

# Install a Python expert skill
ait install apex-ai/skills/python

# Install a debugging prompt
ait install apex-ai/prompts/debug-helper
```

#### Step 4: Use Your AI Tools

Open your project in **Cursor** or **GitHub Copilot**, and the agents are automatically loaded!

```bash
# In Cursor, ask:
# "Review this function for bugs"

# The code-reviewer agent will provide expert feedback
```

### For Teams

#### Step 1: Project Lead Sets Up

```bash
# Initialize the project
cd your-team-project
ait init

# Add team agents to ait.yml
cat > ait.yml << EOF
name: team-project
version: 1.0.0
dependencies:
  - apex-ai/agents/code-reviewer
  - apex-ai/agents/test-generator
  - apex-ai/skills/python
  - apex-ai/skills/typescript
  - apex-ai/prompts/pr-description
EOF

# Install everything
ait install

# Commit to git
git add ait.yml ait.lock .cursorrules .github/ .opencode/
git commit -m "Add AI development assistants"
git push
```

#### Step 2: Team Members Clone and Go

```bash
# Clone the repository
git clone https://github.com/your-team/team-project
cd team-project

# AI agents are already available!
# Open in Cursor → agents load automatically
# Open in VS Code with Copilot → agents load automatically

# Optional: Install to personal AI tools
ait sync
```

**No AIT commands needed!** The tool-native files (`.cursorrules`, `.github/agents/`) work automatically.

---

## How It Works

### The AIT Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                    Developer Workflow                        │
└─────────────────────────────────────────────────────────────┘

1. DEFINE (ait.yml)
   ┌─────────────────────────────────┐
   │ name: my-project                │
   │ version: 1.0.0                  │
   │ dependencies:                   │
   │   - org/repo/agents/reviewer    │
   │   - org/repo/skills/python      │
   └─────────────────────────────────┘
                 │
                 ▼
2. INSTALL (ait install)
   ┌─────────────────────────────────┐
   │ • Fetch from remote repos       │
   │ • Resolve dependencies          │
   │ • Install to target paths       │
   │ • Create ait.lock               │
   └─────────────────────────────────┘
                 │
                 ▼
3. TOOL-NATIVE FILES CREATED
   ┌─────────────────────────────────┐
   │ .cursorrules                    │
   │ .github/agents/reviewer.agent.md│
   │ .opencode/agents/reviewer/      │
   └─────────────────────────────────┘
                 │
                 ▼
4. AI TOOLS AUTO-DETECT
   ┌─────────────────────────────────┐
   │ ✓ Cursor loads .cursorrules     │
   │ ✓ Copilot loads .github/agents/ │
   │ ✓ OpenCode loads .opencode/     │
   └─────────────────────────────────┘
                 │
                 ▼
5. COMMIT TO GIT (team sharing)
   ┌─────────────────────────────────┐
   │ git add ait.yml .cursorrules    │
   │ git commit -m "Add AI agents"   │
   │ git push                        │
   └─────────────────────────────────┘
                 │
                 ▼
6. TEAM CLONES → WORKS IMMEDIATELY
   ┌─────────────────────────────────┐
   │ git clone → Open in tool → ✓   │
   │ No ait commands needed!         │
   └─────────────────────────────────┘
```

### Installation Modes

AIT supports two installation modes:

#### 1. Project-Level (Default) - For Teams

Installs packages to tool-native paths in your project root:

```
your-project/
├── ait.yml                          # Dependency manifest
├── ait.lock                         # Version lock file
├── .cursorrules                     # Cursor auto-loads this
├── .github/
│   └── agents/
│       └── reviewer.agent.md        # Copilot auto-loads this
└── .opencode/
    └── agents/
        └── reviewer/
            └── AGENT.md             # OpenCode loads this
```

**Benefits:**
- ✅ Auto-detected by AI tools (no config needed)
- ✅ Commit to git for team sharing
- ✅ Version controlled with your code
- ✅ Works immediately when teammates clone

#### 2. Global Installation - For Personal Use

Installs to your AI tools' config directories:

```bash
ait install --global org/repo/agents/reviewer
```

Installs to:
- `~/.config/opencode/agents/` (OpenCode)
- `~/Library/Application Support/Cursor/User/` (Cursor)
- `~/.claude/` (Claude Desktop)

**Benefits:**
- ✅ Available across all projects
- ✅ Personal workspace customization
- ✅ No project files created

---

## Commands Reference

### `ait init`

Initialize a new project with `ait.yml`.

```bash
# Interactive mode (prompts for details)
ait init

# Use defaults (name from directory)
ait init --defaults

# Specify project details
ait init --name my-project --version 1.0.0
```

**Creates:**
```yaml
name: my-project
version: 1.0.0
dependencies: []
```

---

### `ait install`

Install packages from `ait.yml` or command line.

```bash
# Install all dependencies from ait.yml
ait install

# Install specific packages (auto-adds to ait.yml)
ait install org/repo/agents/code-reviewer
ait install org/repo/skills/python@^2.0.0

# Install without saving to ait.yml
ait install org/repo/agents/reviewer --save=false

# Install to global AI tools
ait install --global org/repo/agents/reviewer

# Install to specific tools only
ait install --target opencode --target cursor org/repo/agents/reviewer
```

**Flags:**
- `--global, -g` - Install to AI tools globally (not project root)
- `--target, -t` - Specify tools: `opencode`, `cursor`, `claude`
- `--save, -s` - Add to ait.yml (default: true, use `--save=false` to disable)

**What it does:**
1. Parses package specs
2. Resolves dependencies (including transitive)
3. Fetches packages from remote sources
4. Installs to target locations
5. Creates/updates `ait.lock`
6. Updates `ait.yml` (if --save=true)

---

### `ait list`

List installed packages.

```bash
# List all packages
ait list

# List for specific tool
ait list --target opencode

# List global packages
ait list --global
```

**Output:**
```
Listing project-local packages from /path/to/project
  • code-reviewer (1.0.0) [agent]
  • python (2.1.0) [skill]
  • debug-helper (1.5.0) [prompt]

✓ Total: 3 package(s) installed
ℹ   • 1 agent(s)
ℹ   • 1 skill(s)
ℹ   • 1 prompt(s)
```

---

### `ait update`

Update packages to their latest compatible versions.

```bash
# Update all packages from ait.yml
ait update

# Update specific packages
ait update code-reviewer python-skill

# Update for specific tools
ait update --target opencode --target cursor
```

**What it does:**
1. Reads `ait.yml` and `ait.lock`
2. Checks for newer versions matching constraints
3. Updates packages respecting semver
4. Updates `ait.lock`

---

### `ait outdated`

Check for outdated packages.

```bash
# Check for outdated packages
ait outdated

# Show all packages (including up-to-date)
ait outdated --all
```

**Output:**
```
ℹ Checking for outdated packages...

PACKAGE         CURRENT    LATEST     TYPE     STATUS
--------------  ---------  ---------  -------  ----------
code-reviewer   1.0.0      2.0.0      agent    outdated
python          2.1.0      2.1.0      skill    up-to-date

⚠ 1 package(s) outdated
ℹ Run 'ait update' to update packages
```

---

### `ait uninstall`

Remove installed packages.

```bash
# Uninstall from project root (default)
ait uninstall code-reviewer

# Uninstall from global tools
ait uninstall code-reviewer --global

# Uninstall from specific tools only
ait uninstall code-reviewer --target cursor

# Uninstall multiple packages
ait uninstall code-reviewer python-skill
```

---

### `ait sync`

Sync project-level packages to global AI tools.

```bash
# Sync all project packages to global tools
ait sync

# Sync to specific tools only
ait sync --target opencode
```

**Use case:** Tools that don't support project-level detection (like Claude Desktop).

---

### `ait doctor`

Check AIT installation health and configuration.

```bash
# Run health checks
ait doctor
```

**Checks:**
- ✓ Git installation
- ✓ AI tools detection (OpenCode, Cursor, Claude)
- ✓ Configuration directories and permissions
- ✓ Project manifest (ait.yml) validity
- ✓ Lockfile (ait.lock) validity
- ✓ GitHub authentication setup

**Output:**
```
ℹ Running AIT health checks...

✓ Git installation: git version 2.53.0
✓ AI tools detection: Found: [opencode cursor claude]
✓ OpenCode: Detected at ~/.config/opencode
✓ Cursor: Detected at ~/Library/.../Cursor/User
✓ Claude Desktop: Detected at ~/.claude
✓ Current directory: /path/to/project
✓ Write permissions: Current directory is writable
⚠ ait.yml: Not found (run 'ait init' to create)
⚠ ait.lock: Not found (will be created on first install)
✓ GitHub authentication: GitHub token configured

⚠ 2 warning(s)
ℹ Your AIT installation has warnings but should work
```

---

### `ait generate`

Generate `ait.yml` from existing installations.

```bash
# Generate from project root installation
ait generate

# Generate from global installations
ait generate --global
```

**Use case:** Migrate existing manual installations to AIT.

---

## Creating an AI Marketplace Monorepo

Create a centralized repository for your organization's AI agents, skills, and prompts.

### Repository Structure

```
ai-marketplace/
├── README.md                         # Marketplace documentation
├── agents/
│   ├── code-reviewer/
│   │   ├── package.yml               # Package metadata
│   │   ├── AGENT.md                  # Agent instructions
│   │   └── README.md                 # Agent documentation
│   ├── test-generator/
│   │   ├── package.yml
│   │   ├── AGENT.md
│   │   └── README.md
│   └── pr-assistant/
│       ├── package.yml
│       └── AGENT.md
├── skills/
│   ├── python/
│   │   ├── package.yml
│   │   ├── SKILL.md
│   │   └── examples/
│   ├── typescript/
│   │   ├── package.yml
│   │   └── SKILL.md
│   └── docker/
│       ├── package.yml
│       └── SKILL.md
├── prompts/
│   ├── debug-helper/
│   │   ├── package.yml
│   │   └── prompt.txt
│   ├── pr-description/
│   │   ├── package.yml
│   │   └── prompt.txt
│   └── code-explanation/
│       ├── package.yml
│       └── prompt.txt
└── mcp/
    └── filesystem-server/
        ├── package.yml
        └── server.js
```

### Step-by-Step Setup

#### 1. Create the Repository

```bash
# Create repository
mkdir ai-marketplace
cd ai-marketplace
git init

# Create directory structure
mkdir -p agents skills prompts mcp
```

#### 2. Create Your First Package

```bash
# Create a code reviewer agent
mkdir -p agents/code-reviewer
cd agents/code-reviewer

# Create package.yml
cat > package.yml << 'EOF'
name: code-reviewer
version: 1.0.0
type: agent
description: Expert code reviewer that checks for bugs, performance, and best practices
author: Your Organization
repository: https://github.com/your-org/ai-marketplace

dependencies: []

metadata:
  tags:
    - code-review
    - quality
    - best-practices
  languages:
    - typescript
    - python
    - go
EOF

# Create AGENT.md
cat > AGENT.md << 'EOF'
# Code Reviewer Agent

You are an expert code reviewer with deep knowledge of software engineering best practices.

## Your Role

Analyze code for:
- Bugs and potential errors
- Performance issues
- Security vulnerabilities
- Code style and readability
- Best practices violations

## Core Competencies

1. **Bug Detection**: Identify logic errors, edge cases, and potential runtime issues
2. **Performance**: Spot inefficient algorithms, memory leaks, and bottlenecks
3. **Security**: Recognize common vulnerabilities (XSS, SQL injection, etc.)
4. **Best Practices**: Enforce SOLID principles, DRY, and language-specific idioms

## Guidelines

- Always explain WHY something is an issue
- Provide specific code examples for fixes
- Prioritize issues by severity (critical, high, medium, low)
- Be constructive and educational in feedback
EOF

# Create README.md
cat > README.md << 'EOF'
# Code Reviewer Agent

Expert code review assistant that provides comprehensive feedback on code quality, bugs, performance, and best practices.

## Features

- 🐛 Bug detection and analysis
- ⚡ Performance optimization suggestions
- 🔒 Security vulnerability identification
- 📚 Best practices enforcement

## Installation

```bash
# Using AIT
ait install your-org/ai-marketplace/agents/code-reviewer

# Or add to ait.yml
dependencies:
  - your-org/ai-marketplace/agents/code-reviewer
```

## Usage

In your AI tool (Cursor, GitHub Copilot, etc.):

```
Review this function for potential issues:
[paste code]
```

The agent will provide detailed feedback on bugs, performance, security, and best practices.

## Version History

- v1.0.0 - Initial release
EOF
```

#### 3. Create More Packages

```bash
# Create a Python skill
mkdir -p ../../skills/python
cd ../../skills/python

cat > package.yml << 'EOF'
name: python
version: 2.1.0
type: skill
description: Expert Python programming knowledge
author: Your Organization

dependencies: []

metadata:
  tags:
    - python
    - programming
  frameworks:
    - django
    - flask
    - fastapi
EOF

cat > SKILL.md << 'EOF'
# Python Expert Skill

You are a Python expert with deep knowledge of the language and its ecosystem.

## Core Competencies

1. **Language Features**: List comprehensions, generators, decorators, context managers
2. **Standard Library**: Collections, itertools, functools, asyncio
3. **Frameworks**: Django, Flask, FastAPI, Celery
4. **Testing**: pytest, unittest, mocking, fixtures
5. **Performance**: Profiling, optimization, Cython, multiprocessing
6. **Best Practices**: PEP 8, type hints, documentation, packaging

## Guidelines

- Write Pythonic code following PEP 8
- Use type hints for better code clarity
- Prefer standard library over third-party when reasonable
- Write comprehensive docstrings
- Consider performance implications
EOF
```

#### 4. Create Package Index (Optional)

```bash
# Create a catalog of all packages
cd ../..
cat > catalog.json << 'EOF'
{
  "version": "1.0",
  "packages": [
    {
      "name": "code-reviewer",
      "type": "agent",
      "version": "1.0.0",
      "path": "agents/code-reviewer",
      "description": "Expert code reviewer",
      "tags": ["code-review", "quality"]
    },
    {
      "name": "python",
      "type": "skill",
      "version": "2.1.0",
      "path": "skills/python",
      "description": "Python programming expert",
      "tags": ["python", "programming"]
    }
  ]
}
EOF
```

#### 5. Create Marketplace README

```bash
cat > README.md << 'EOF'
# AI Marketplace

Centralized repository of AI agents, skills, and prompts for our organization.

## 📦 Available Packages

### Agents

- **code-reviewer** (v1.0.0) - Expert code review assistant
- **test-generator** (v1.0.0) - Automated test generation
- **pr-assistant** (v1.0.0) - Pull request description helper

### Skills

- **python** (v2.1.0) - Python programming expert
- **typescript** (v2.0.0) - TypeScript expert
- **docker** (v1.5.0) - Docker and containerization

### Prompts

- **debug-helper** - Debugging assistance
- **pr-description** - PR description generator
- **code-explanation** - Code explainer

## 🚀 Quick Start

```bash
# Install a package
ait install your-org/ai-marketplace/agents/code-reviewer

# Or add to your ait.yml
dependencies:
  - your-org/ai-marketplace/agents/code-reviewer
  - your-org/ai-marketplace/skills/python
```

## 📋 Usage

1. Browse packages in the repository
2. Add desired packages to your `ait.yml`
3. Run `ait install`
4. Packages are automatically available in your AI tools

## 🤝 Contributing

1. Create a new package in the appropriate directory
2. Include `package.yml` and package file (AGENT.md, SKILL.md, etc.)
3. Add documentation in README.md
4. Submit a pull request
EOF
```

#### 6. Version Control and Tagging

```bash
# Commit initial structure
git add .
git commit -m "Initial marketplace setup with code-reviewer and python skill"

# Create version tags for releases
git tag v1.0.0
git push origin main --tags

# Push to GitHub
git remote add origin https://github.com/your-org/ai-marketplace
git push -u origin main
```

### Using Your Marketplace

#### For Developers

```bash
# In any project
cat > ait.yml << EOF
name: my-project
version: 1.0.0
dependencies:
  - your-org/ai-marketplace/agents/code-reviewer@1.0.0
  - your-org/ai-marketplace/skills/python@^2.0.0
  - your-org/ai-marketplace/prompts/debug-helper
EOF

# Install
ait install

# Your AI tools now have access to all these packages!
```

#### For Teams

```yaml
# ait.yml
name: team-backend
version: 1.0.0
dependencies:
  # Agents for code quality
  - your-org/ai-marketplace/agents/code-reviewer
  - your-org/ai-marketplace/agents/test-generator
  
  # Skills for our tech stack
  - your-org/ai-marketplace/skills/python
  - your-org/ai-marketplace/skills/docker
  
  # Helpful prompts
  - your-org/ai-marketplace/prompts/pr-description
```

### Private Marketplace Authentication

For private repositories:

```bash
# Using GitHub CLI
gh auth login
export GH_TOKEN=$(gh auth token)

# Or personal access token
export GITHUB_TOKEN=ghp_your_token_here

# Install works automatically
ait install
```

### Marketplace Best Practices

1. **Semantic Versioning**: Use semver for package versions
2. **Documentation**: Include comprehensive README for each package
3. **Testing**: Test packages before releasing
4. **Changelog**: Maintain version history in README
5. **Examples**: Provide usage examples
6. **Tags**: Use metadata tags for discoverability
7. **Dependencies**: Declare package dependencies in `package.yml`

---

## Developer Setup

### Prerequisites

- Go 1.23 or higher
- Git
- Make (optional, for convenience)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ZakHargz/ait
cd ait

# Install dependencies
go mod download

# Build
make build
# Or: go build -o bin/ait ./cmd/ait

# Run tests
make test
# Or: go test ./...

# Run linter
make lint
# Or: golangci-lint run

# Build for all platforms
make build-all
```

### Project Structure

```
ait/
├── cmd/
│   └── ait/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/                     # CLI commands
│   │   ├── root.go              # Root command setup
│   │   ├── init.go              # ait init
│   │   ├── install.go           # ait install
│   │   ├── list.go              # ait list
│   │   ├── update.go            # ait update
│   │   ├── uninstall.go         # ait uninstall
│   │   ├── sync.go              # ait sync
│   │   ├── doctor.go            # ait doctor
│   │   ├── outdated.go          # ait outdated
│   │   └── generate.go          # ait generate
│   ├── config/                  # Configuration
│   │   ├── manifest.go          # ait.yml parser
│   │   ├── lockfile.go          # ait.lock handler
│   │   └── metadata.go          # package.yml parser
│   ├── resolver/                # Dependency resolution
│   │   └── resolver.go          # Graph-based resolver
│   ├── sources/                 # Package sources
│   │   ├── source.go            # Source interface
│   │   ├── git.go               # Git repository handler
│   │   ├── auth.go              # GitHub authentication
│   │   └── factory.go           # Source factory
│   ├── adapters/                # Tool adapters
│   │   ├── adapter.go           # Base adapter
│   │   ├── opencode.go          # OpenCode adapter
│   │   ├── cursor.go            # Cursor adapter
│   │   ├── claude.go            # Claude adapter
│   │   └── projectroot.go       # Project-level adapter
│   ├── packages/                # Package types
│   │   └── package.go           # Package struct
│   └── utils/                   # Utilities
│       └── utils.go             # Helper functions
├── .github/
│   └── workflows/
│       ├── ci.yml               # CI pipeline
│       └── release.yml          # Release automation
├── Formula/
│   └── ait.rb                   # Homebrew formula
├── Makefile                     # Build automation
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/cli -v
go test ./internal/resolver -v

# Run with race detection
go test -race ./...
```

### Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**
4. **Run tests**: `make test`
5. **Run linter**: `make lint`
6. **Commit**: `git commit -m "Add amazing feature"`
7. **Push**: `git push origin feature/amazing-feature`
8. **Create a Pull Request**

### Code Style

- Follow Go conventions and idioms
- Run `gofmt` before committing
- Use `golangci-lint` for linting
- Write tests for new features
- Update documentation for user-facing changes

---

## Use Cases

### 1. Personal Developer Setup

**Scenario**: Individual developer wants consistent AI assistants across projects.

```bash
# Install personal favorites globally
ait install --global apex-ai/agents/code-reviewer
ait install --global apex-ai/skills/python
ait install --global apex-ai/skills/typescript

# Available in all projects automatically
```

### 2. Team Onboarding

**Scenario**: New developer joins team and needs AI tools set up.

```bash
# New developer clones repo
git clone https://github.com/team/project
cd project

# AI agents already work!
# .cursorrules and .github/agents/ are in the repo

# Optional: sync to personal tools
ait sync
```

### 3. Multi-Project Consistency

**Scenario**: Organization wants standardized AI assistants across all projects.

```yaml
# Template ait.yml for all projects
name: ${PROJECT_NAME}
version: 1.0.0
dependencies:
  # Standard agents for all projects
  - company/ai-marketplace/agents/code-reviewer
  - company/ai-marketplace/agents/security-scanner
  - company/ai-marketplace/agents/test-generator
  
  # Standard skills
  - company/ai-marketplace/skills/company-standards
```

### 4. Specialized Project Setup

**Scenario**: Machine learning project needs ML-specific AI assistants.

```yaml
name: ml-project
version: 1.0.0
dependencies:
  # ML-specific agents
  - ml-org/agents/model-reviewer
  - ml-org/agents/data-validator
  
  # ML skills
  - ml-org/skills/pytorch
  - ml-org/skills/tensorflow
  - ml-org/skills/data-science
  
  # ML prompts
  - ml-org/prompts/experiment-design
  - ml-org/prompts/hyperparameter-tuning
```

### 5. CI/CD Integration

**Scenario**: Run AI-powered code reviews in CI pipeline.

```yaml
# .github/workflows/ai-review.yml
name: AI Code Review
on: [pull_request]

jobs:
  ai-review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install AIT
        run: |
          brew tap zakhargz/ait
          brew install ait
      
      - name: Install AI agents
        run: ait install
      
      - name: Run AI review
        run: |
          # Use installed agents for automated review
          ait list
```

---

## Supported AI Tools

### OpenCode ✅

- **Status**: Fully supported
- **Project-Level**: `.opencode/agents/<name>/AGENT.md`
- **Global Location**: `~/.config/opencode/agents/`
- **Format**: Native AGENT.md format
- **Auto-Detection**: Proposed standard

### Cursor ✅

- **Status**: Fully supported with project-level detection
- **Project-Level**: `.cursorrules` (auto-detected by Cursor)
- **Global Location**: `~/Library/Application Support/Cursor/User/` (macOS)
- **Format**: Converts AGENT.md to `.cursorrules` format
- **Auto-Detection**: ✅ Yes - Cursor loads `.cursorrules` at project root

### GitHub Copilot ✅

- **Status**: Fully supported with project-level detection
- **Project-Level**: `.github/agents/<name>.agent.md`
- **Format**: Native `.agent.md` format (APM standard)
- **Auto-Detection**: ✅ Yes - works in VS Code, IntelliJ, GitHub.com
- **Standard**: Follows Microsoft APM specification

### Claude Desktop ✅

- **Status**: Fully supported (global only)
- **Global Location**: `~/.claude/` (macOS)
- **Format**: Converts to Claude-specific format
- **Note**: Use `ait sync` to push project packages to Claude

---

## Package Format

### package.yml

Every package must include a `package.yml` metadata file:

```yaml
name: code-reviewer
version: 1.0.0
type: agent                    # agent, skill, prompt, or mcp
description: Expert code reviewer
author: Your Name
repository: https://github.com/org/repo

dependencies:
  - org/repo/skills/python     # Package dependencies
  - org/repo/skills/typescript

metadata:
  tags:
    - code-review
    - quality
  languages:
    - python
    - typescript
    - go
```

### AGENT.md Format

Agent instruction files follow this structure:

```markdown
# Agent Name

Brief description of the agent's purpose.

## Your Role

Describe what this agent does and its primary responsibilities.

## Core Competencies

1. **Competency 1**: Description
2. **Competency 2**: Description
3. **Competency 3**: Description

## Guidelines

- Guideline 1
- Guideline 2
- Guideline 3

## Examples

### Example 1
[Example usage]

### Example 2
[Example usage]
```

### SKILL.md Format

Skill files define domain expertise:

```markdown
# Skill Name

Description of the skill domain.

## Core Competencies

1. **Area 1**: Expertise description
2. **Area 2**: Expertise description

## Best Practices

- Practice 1
- Practice 2

## Common Patterns

### Pattern 1
[Code example]

### Pattern 2
[Code example]
```

### Prompt Format

Simple text files with prompts:

```
You are helping debug a complex issue. Follow these steps:

1. Understand the problem
2. Identify root cause
3. Propose solutions
4. Test fixes

Be thorough and methodical.
```

---

## Troubleshooting

### Common Issues

#### 1. "No AI tools detected"

**Problem**: `ait doctor` shows no AI tools found.

**Solution**:
```bash
# Install at least one AI tool:
# - OpenCode: https://opencode.ai
# - Cursor: https://cursor.sh
# - Claude Desktop: https://claude.ai/download

# Verify installation
ait doctor
```

#### 2. "GitHub authentication failed"

**Problem**: Cannot access private repositories.

**Solution**:
```bash
# Using GitHub CLI
gh auth login
export GH_TOKEN=$(gh auth token)

# Or use personal access token
export GITHUB_TOKEN=ghp_your_token_here

# Verify
ait doctor
```

#### 3. "Package not found"

**Problem**: `ait install org/repo/package` fails.

**Solution**:
```bash
# Check the repository exists
# Verify the path is correct: org/repo/path/to/package
# Ensure you have access (for private repos)
# Check your internet connection

# Try with explicit version
ait install org/repo/package@1.0.0
```

#### 4. "Permission denied"

**Problem**: Cannot write to installation directories.

**Solution**:
```bash
# For project-level (default)
# Ensure you have write permissions in current directory
chmod u+w .

# For global installation
# Ensure AI tool directories are writable
# Or use project-level installation instead
ait install  # without --global
```

#### 5. Agents not loading in AI tool

**Problem**: Installed agents don't appear in Cursor/Copilot.

**Solution**:
```bash
# Verify installation
ait list

# Check files were created
ls -la .cursorrules .github/agents/

# Restart your AI tool
# Some tools require restart to detect new files

# For Cursor specifically
# Close Cursor → Delete .cursorrules cache → Reopen
```

### Debug Mode

```bash
# Enable verbose output
ait --verbose install org/repo/package

# Check system health
ait doctor

# Verify package is in cache
ls ~/.ait/cache/
```

### Getting Help

- **Issues**: https://github.com/ZakHargz/ait/issues
- **Discussions**: https://github.com/ZakHargz/ait/discussions
- **Documentation**: https://github.com/ZakHargz/ait/wiki

---

## Contributing

We welcome contributions! Here's how to get started:

### Reporting Issues

1. Check existing issues first
2. Use issue templates
3. Include reproduction steps
4. Attach relevant logs and `ait doctor` output

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Update documentation
6. Submit PR with clear description

### Development Workflow

```bash
# Setup
git clone https://github.com/ZakHargz/ait
cd ait
go mod download

# Make changes
# ... edit code ...

# Test
make test

# Lint
make lint

# Build
make build

# Test locally
./bin/ait doctor
```

### Code Guidelines

- Follow Go conventions
- Write tests for new features
- Update documentation
- Keep commits focused and atomic
- Write descriptive commit messages

---

## Relationship with Microsoft APM

AIT aligns with [Microsoft's Agent Package Manager (APM)](https://microsoft.github.io/apm/) while maintaining compatibility with existing tools.

### How AIT Differs

| Feature | APM | AIT |
|---------|-----|-----|
| **Package Sources** | npm registry | GitHub, GitLab, Git, Local |
| **Installation** | Node.js required | Go binary (no runtime) |
| **Multi-tool** | Primarily VS Code/Copilot | OpenCode, Cursor, Claude, Copilot |
| **Project-level** | `.github/agents/` | `.github/agents/` + `.cursorrules` + `.opencode/` |
| **Dependency Resolution** | npm-style | Built-in resolver |
| **Lockfiles** | package-lock.json style | ait.lock (YAML) |

### Compatibility

AIT is **fully compatible** with APM packages:
- ✅ Reads `.agent.md` files
- ✅ Installs to `.github/agents/`
- ✅ Supports `package.json` metadata
- ✅ Works with GitHub Copilot, VS Code

### Why Both Projects Matter

- **APM**: Standard for npm-based workflows, VS Code integration
- **AIT**: Multi-tool support, no Node.js dependency, extended features

Use both! AIT can install APM packages, and vice versa.

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Microsoft APM](https://microsoft.github.io/apm/) - Package format inspiration
- [AGENTS.md](https://agents.md) - Agent format specification
- [Agent Skills](https://agentskills.io) - Skills format
- [Model Context Protocol](https://modelcontextprotocol.io) - MCP integration

---

**Made with ❤️ for the AI development community**

🌟 Star us on GitHub: https://github.com/ZakHargz/ait
