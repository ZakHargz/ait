# AIT - AI Toolkit Package Manager

A CLI package manager for AI agents, skills, prompts, and MCP servers. Think npm/pip for AI tooling across OpenCode, Cursor, Claude Desktop, and more.

## Features

- 📦 **Declarative Dependencies** - Define AI tools in `ait.yml`
- 🔄 **Multi-Tool Sync** - Install to OpenCode, Cursor, Claude Desktop simultaneously
- 🌲 **Dependency Resolution** - Automatic transitive dependency handling
- 📌 **Lock Files** - Reproducible installations with `ait.lock`
- 🏷️ **Semantic Versioning** - Version constraints (`^1.0.0`, `~2.1.0`, `>=3.0.0`)
- 🌍 **Multiple Sources** - GitHub, GitLab, generic Git, or local packages
- 💾 **Smart Caching** - Local cache to speed up installations

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/apex-ai/ait
cd ait

# Build and install globally
make build
make install

# Or build without installing
make build
./bin/ait --version
```

### Requirements

- Go 1.26+ (for building from source)
- One or more AI tools: OpenCode, Cursor, or Claude Desktop

## Quick Start

### 1. Initialize a Project

```bash
# Interactive initialization
ait init

# Or use defaults
ait init --defaults
```

This creates an `ait.yml` manifest:

```yaml
name: my-project
version: 1.0.0
dependencies:
  agents: []
  skills: []
  prompts: []
targets:
  - opencode
```

### 2. Add Dependencies

Edit `ait.yml` to add packages:

```yaml
name: my-project
version: 1.0.0
dependencies:
  agents:
    - github:org/repo/agents/code-reviewer@1.0.0
  skills:
    - github:org/repo/skills/python@^2.0.0
  prompts:
    - github:org/repo/prompts/debug@~1.5.0
targets:
  - opencode
  - cursor
```

### 3. Install Packages

```bash
# Install all dependencies from ait.yml
ait install

# Or install specific packages
ait install github:org/repo/agents/code-reviewer@1.0.0

# Install to specific tools only
ait install --target opencode
```

### 4. List Installed Packages

```bash
# List all installed packages
ait list

# List for specific tool
ait list --target opencode
```

## Package Specification Format

Packages are specified using the format: `type:location@version`

### Supported Sources

**GitHub:**
```
github:org/repo/path/to/package@version
```

**GitLab:**
```
gitlab:org/repo/path/to/package@version
```

**Generic Git:**
```
git:https://git.example.com/repo/path/to/package@version
```

**Local Filesystem:**
```
local:./path/to/package@version
local:~/packages/agents/code-reviewer@1.0.0
```

### Version Formats

- **Exact**: `1.0.0` - Exact version only
- **Caret**: `^1.0.0` - Compatible with 1.x.x (minor/patch updates)
- **Tilde**: `~1.5.0` - Compatible with 1.5.x (patch updates only)
- **Range**: `>=1.0.0 <2.0.0` - Version range
- **Branch**: `main`, `develop` - Git branch names
- **Commit**: `abc123...` - Specific commit hash

## Creating Packages

### Package Structure

Each package needs a `package.yml` metadata file:

```
my-package/
├── package.yml          # Required: Package metadata
├── AGENT.md            # For agents
├── SKILL.md            # For skills
└── prompt.txt          # For prompts
```

### package.yml Format

```yaml
name: code-reviewer
version: 1.0.0
type: agent              # agent, skill, prompt, or mcp
description: Expert code reviewer for best practices
author:
  name: Your Name
  email: you@example.com
license: MIT
compatibility:
  - opencode
  - cursor
  - claude
files:
  opencode: AGENT.md
  cursor: AGENT.md
  claude: AGENT.md
dependencies:           # Optional: Package dependencies
  skills:
    - github:org/repo/skills/git-workflow@^1.0.0
tags:
  - code-review
  - quality
```

### Agent Format (AGENT.md)

Agents use frontmatter for metadata:

```markdown
---
name: code-reviewer
version: 1.0.0
description: Expert code reviewer
---

# Code Reviewer Agent

You are an expert code reviewer...

## Your Role
...
```

### Skill Format (SKILL.md)

Similar to agents:

```markdown
---
name: python
version: 2.1.0
description: Python development expertise
---

# Python Development Skill

## Core Competencies
...
```

### Prompt Format

Simple text files:

```
Please help me debug this issue:
1. Analyze the error
2. Identify root cause
3. Suggest fixes
```

## Configuration Files

### ait.yml (Project Manifest)

```yaml
name: my-project
version: 1.0.0
description: My AI-powered project

# Optional: Named package sources
sources:
  - name: company-toolkit
    url: github:myorg/ai-toolkit

dependencies:
  agents:
    - github:org/repo/agents/code-reviewer@^1.0.0
    - company-toolkit/agents/custom-agent@~2.0.0
  skills:
    - github:org/repo/skills/python@^2.0.0
  prompts:
    - local:./prompts/custom@1.0.0
  mcp:
    - github:org/repo/servers/custom-mcp@^1.0.0

# Target AI tools
targets:
  - opencode
  - cursor
  - claude
```

### ait.lock (Lock File)

Auto-generated on install:

```yaml
version: "1.0"
generated: 2026-03-27T16:09:03Z
packages:
  code-reviewer:
    name: code-reviewer
    version: 1.0.0
    type: agent
    source: github:org/repo/agents/code-reviewer@^1.0.0
    resolved: 1.0.5  # Exact resolved version
    installed:
      - opencode
      - cursor
```

## Commands Reference

### ait init

Initialize a new project with `ait.yml`:

```bash
# Interactive mode
ait init

# Use defaults
ait init --defaults

# Specify project details
ait init --name my-project --version 1.0.0
```

### ait install

Install packages:

```bash
# Install from ait.yml
ait install

# Install specific packages
ait install github:org/repo/agents/reviewer@1.0.0

# Install to specific tools
ait install --target opencode --target cursor

# Save installed packages to ait.yml
ait install github:org/repo/agents/reviewer@1.0.0 --save
```

### ait list

List installed packages:

```bash
# List all
ait list

# List for specific tool
ait list --target opencode
```

## Supported AI Tools

### OpenCode ✅

- **Status**: Fully supported
- **Location**: `~/.config/opencode/`
- **Agents**: `~/.config/opencode/agents/<name>/AGENT.md`
- **Skills**: `~/.config/opencode/skills/<name>/SKILL.md`
- **Prompts**: `~/.config/opencode/prompts/<name>.txt`

### Cursor 🔄

- **Status**: Planned
- **Location**: `~/Library/Application Support/Cursor/User/`
- **Format**: Converts agents to `.cursorrules`

### Claude Desktop 🔄

- **Status**: Planned
- **Location**: `~/.claude/`
- **Format**: Same as OpenCode (AGENT.md, SKILL.md)

## Development Status

### ✅ Implemented (MVP)

- CLI framework (Cobra/Viper)
- Configuration parsing (ait.yml, package.yml, ait.lock)
- `ait init` command
- `ait install` command with full features
- `ait list` command
- OpenCode adapter (agents, skills, prompts)
- Git-based sources (GitHub, GitLab, generic)
- Local filesystem sources
- Semantic versioning with constraints
- Dependency resolution with cycle detection
- Lock file generation
- Local caching (~/.ait/cache/)
- Test package repository

### 🔄 Planned

- `ait update` command - Update packages to latest
- `ait uninstall` command - Remove packages
- Cursor adapter - .cursorrules conversion
- Claude Desktop adapter
- `ait search` command - Package discovery
- `ait audit` command - Security scanning
- `ait doctor` command - Health checks
- MCP server support

## Project Structure

```
ait/
├── cmd/ait/              # CLI entry point
├── internal/
│   ├── adapters/         # Platform adapters
│   │   ├── adapter.go   # Interface
│   │   ├── opencode.go  # OpenCode implementation
│   │   └── detector.go  # Tool detection
│   ├── cli/              # Commands
│   │   ├── root.go      # Root command
│   │   ├── init.go      # Init command
│   │   ├── install.go   # Install command
│   │   └── list.go      # List command
│   ├── config/           # Configuration
│   │   ├── manifest.go  # ait.yml
│   │   ├── package.go   # package.yml
│   │   └── lockfile.go  # ait.lock
│   ├── sources/          # Package sources
│   │   ├── source.go    # Interface
│   │   ├── git.go       # Git implementation
│   │   └── factory.go   # Source factory
│   ├── resolver/         # Dependency resolution
│   ├── packages/         # Package model
│   └── utils/            # Utilities
├── test-packages/        # Test packages
├── Makefile             # Build automation
├── README.md            # This file
├── DEVELOPMENT.md       # Development notes
└── go.mod               # Go dependencies
```

## Examples

### Example 1: Personal Agents

```yaml
# ait.yml
name: my-agents
version: 1.0.0
dependencies:
  agents:
    - local:~/my-agents/code-reviewer@1.0.0
    - local:~/my-agents/documentation@1.0.0
targets:
  - opencode
```

### Example 2: Team Toolkit

```yaml
# ait.yml
name: team-project
version: 1.0.0
sources:
  - name: company
    url: github:mycompany/ai-toolkit
dependencies:
  agents:
    - company/agents/security-reviewer@^2.0.0
    - company/agents/performance-analyzer@^1.5.0
  skills:
    - company/skills/python@^3.0.0
    - company/skills/terraform@^1.0.0
targets:
  - opencode
  - cursor
```

### Example 3: Public Packages

```yaml
# ait.yml
name: open-source-project
version: 1.0.0
dependencies:
  agents:
    - github:ai-packages/agents/code-reviewer@^1.0.0
    - github:ai-packages/agents/test-generator@~2.1.0
  skills:
    - github:ai-packages/skills/javascript@^3.0.0
  prompts:
    - github:ai-packages/prompts/debug-helper@^1.0.0
targets:
  - opencode
```

## Contributing

Contributions welcome! This is an early MVP implementation.

### Building

```bash
make build       # Build binary
make install     # Install globally
make test        # Run tests (when available)
make clean       # Clean build artifacts
```

## License

MIT

## Acknowledgments

Inspired by:
- npm (Node.js package manager)
- pip (Python package manager)
- Microsoft APM (AI package manager concept)
