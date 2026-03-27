# AIT - AI Toolkit Package Manager

A CLI package manager for AI agents, skills, prompts, and MCP servers. Think npm/pip for AI tooling across OpenCode, Cursor, Claude Desktop, and more.

## Features

- 📦 **Declarative Dependencies** - Define AI tools in `ait.yml`
- 🚀 **Project-Level Installation** - Install to tool-native paths (`.cursorrules`, `.github/copilot-instructions.md`) for automatic detection
- 🔄 **Multi-Tool Sync** - Install to OpenCode, Cursor, Claude Desktop, GitHub Copilot simultaneously
- 👥 **Team Sharing** - Commit tool-native files to git for instant team collaboration
- 🌲 **Dependency Resolution** - Automatic transitive dependency handling
- 📌 **Lock Files** - Reproducible installations with `ait.lock`
- 🏷️ **Semantic Versioning** - Optional version constraints (`^1.0.0`, `~2.1.0`, or omit for latest)
- 🌍 **Multiple Sources** - GitHub, GitLab, generic Git, or local packages
- 💾 **Smart Caching** - Local cache to speed up installations

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ZakHargz/ait
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
    - github:org/repo/agents/code-reviewer          # Latest version (recommended)
    - github:org/repo/agents/test-generator@1.0.0   # Specific version
  skills:
    - github:org/repo/skills/python                 # Latest version (recommended)
    - github:org/repo/skills/docker@^2.0.0          # Version constraint
  prompts:
    - github:org/repo/prompts/debug@~1.5.0
```

**Pro Tip**: Omit `@version` to automatically get the latest version!

### 3. Install Packages

```bash
# Install all dependencies from ait.yml to project root (default)
ait install

# Or install specific packages
ait install github:org/repo/agents/code-reviewer

# Install globally to AI tools instead
ait install --global

# Install to specific tools only
ait install --target opencode
```

This creates tool-native files at your project root:
- `.cursorrules` - Auto-detected by Cursor
- `.github/copilot-instructions.md` - Auto-detected by GitHub Copilot
- `.opencode/agents/` - For OpenCode (proposed standard)

**Team Workflow**: Commit these files to git! Your team gets AI agents automatically when they clone the repo - no `ait` commands needed.

### 4. Share with Your Team (Optional)

Commit the generated tool-native files to git:

```bash
git add .cursorrules .github/copilot-instructions.md .opencode/ ait.yml ait.lock
git commit -m "Add AI agents for code review and testing"
git push
```

When teammates clone the repo:
- **Cursor** auto-loads `.cursorrules` immediately
- **GitHub Copilot** auto-loads `.github/copilot-instructions.md` immediately
- **OpenCode** can load from `.opencode/agents/` (or run `ait sync` if needed)

### 5. List Installed Packages

```bash
# List all installed packages
ait list

# List for specific tool
ait list --target opencode
```

### 6. Update Packages

```bash
# Update all packages to latest compatible versions
ait update

# Update specific packages
ait update code-reviewer python-skill

# Update for specific tools only
ait update --target opencode
```

### 7. Uninstall Packages

```bash
# Uninstall a package from project root
ait uninstall code-reviewer

# Uninstall from global tools
ait uninstall code-reviewer --global

# Uninstall from specific tools only
ait uninstall code-reviewer --target cursor
```

## Installation Modes

AIT supports two installation modes:

### Project-Level (Default) - Recommended for Teams

Install packages to tool-native paths at your project root for automatic detection:

```bash
# Default: install to project root
ait install github:org/repo/agents/code-reviewer

# Creates these files:
# .cursorrules                        (Cursor auto-detects)
# .github/copilot-instructions.md     (GitHub Copilot auto-detects)
# .opencode/agents/<name>/AGENT.md    (OpenCode - proposed standard)
```

**Benefits:**
- ✅ AI tools automatically detect agents without manual configuration
- ✅ Commit to git for instant team sharing
- ✅ Version control your AI tooling alongside code
- ✅ Clone and go - teammates get agents automatically

**Best for:**
- Team projects where everyone should have the same AI agents
- Open source projects with contributor guidelines
- Projects with specific domain expertise encoded in agents

### Global Installation - For Personal Use

Install packages to each AI tool's global directory:

```bash
# Install globally to all detected AI tools
ait install --global github:org/repo/agents/code-reviewer

# Installs to:
# ~/.config/opencode/agents/<name>/AGENT.md
# ~/Library/Application Support/Cursor/User/ait-agents/<name>/.cursorrules
# ~/.claude/agents/<name>/AGENT.md
```

**Benefits:**
- ✅ Available across all your projects
- ✅ Personal agents that don't need team sharing
- ✅ Useful for tools without project-level detection

**Best for:**
- Personal productivity agents
- Agents you want available everywhere
- Tools that don't support project-level detection (like Claude Desktop)

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
    - github:org/repo/agents/code-reviewer        # Latest version (recommended)
    - company-toolkit/agents/custom-agent@~2.0.0  # Version constraint
  skills:
    - github:org/repo/skills/python               # Latest version
  prompts:
    - local:./prompts/custom                      # Local package, latest
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
    version: latest              # Requested version
    type: agent
    source: github:org/repo/agents/code-reviewer
    resolved: 1.0.5             # Exact resolved version
    installed:
      - project-root            # or: opencode, cursor, claude
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
# Install from ait.yml to project root (default)
ait install

# Install specific packages to project root
ait install github:org/repo/agents/reviewer

# Install globally to AI tools
ait install --global

# Install to specific tools
ait install --target opencode --target cursor

# Save installed packages to ait.yml
ait install github:org/repo/agents/reviewer --save
```

**Flags:**
- `--global` or `-g` - Install to AI tools globally instead of project root
- `--target` or `-t` - Specify which tools to install to
- `--save` or `-s` - Add installed packages to ait.yml

### ait list

List installed packages:

```bash
# List all
ait list

# List for specific tool
ait list --target opencode
```

### ait update

Update packages to their latest compatible versions:

```bash
# Update all packages from ait.yml
ait update

# Update specific packages
ait update code-reviewer python-skill

# Update for specific tools
ait update --target opencode --target cursor
```

### ait uninstall

Remove installed packages:

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

**Flags:**
- `--global` or `-g` - Uninstall from AI tools globally instead of project root
- `--target` or `-t` - Specify which tools to uninstall from

### ait generate

Generate `ait.yml` from existing installed packages:

```bash
# Generate from project root installation
ait generate

# Generate from global installations
ait generate --global
```

Useful for:
- Creating `ait.yml` from an existing project
- Documenting currently installed packages
- Migrating from manual installation to AIT

### ait sync

Sync project-level packages to global AI tools:

```bash
# Sync all project packages to global tools
ait sync

# Sync to specific tools only
ait sync --target opencode
```

Useful for:
- Tools that don't support project-level detection (e.g., Claude Desktop)
- Making project agents available globally on your machine
- Refreshing global installations after project changes

## Supported AI Tools

### Cursor ✅

- **Status**: Fully supported with project-level detection
- **Project-Level**: `.cursorrules` (auto-detected by Cursor)
- **Global Location**: `~/Library/Application Support/Cursor/User/ait-*` (macOS)
- **Format**: Converts AGENT.md to `.cursorrules` format
- **Auto-Detection**: ✅ Yes - Cursor automatically loads `.cursorrules` at project root

### GitHub Copilot ✅

- **Status**: Fully supported with project-level detection
- **Project-Level**: `.github/copilot-instructions.md` (auto-detected by GitHub Copilot)
- **Format**: Converts AGENT.md to Copilot instructions format
- **Auto-Detection**: ✅ Yes - GitHub Copilot automatically loads project-level instructions

### OpenCode ✅

- **Status**: Fully supported
- **Project-Level**: `.opencode/agents/<name>/AGENT.md` (proposed standard)
- **Global Location**: `~/.config/opencode/`
- **Format**: Native AGENT.md format
- **Auto-Detection**: ⚠️ Use `ait sync` to copy to global location if needed
- **Agents**: `~/.config/opencode/agents/<name>/AGENT.md`
- **Skills**: `~/.config/opencode/skills/<name>/SKILL.md`
- **Prompts**: `~/.config/opencode/prompts/<name>.txt`

### Claude Desktop ✅

- **Status**: Fully supported (global only)
- **Location**: `~/.claude/`
- **Format**: Same as OpenCode (AGENT.md, SKILL.md)
- **Auto-Detection**: ❌ No - Requires global installation with `--global` flag
- **Agents**: `~/.claude/agents/<name>/AGENT.md`
- **Skills**: `~/.claude/skills/<name>/SKILL.md`
- **Prompts**: `~/.claude/prompts/<name>.txt`

**Summary of Project-Level Support:**
- ✅ **Cursor** - Full auto-detection via `.cursorrules`
- ✅ **GitHub Copilot** - Full auto-detection via `.github/copilot-instructions.md`
- ⚠️ **OpenCode** - Project files created, use `ait sync` for global
- ❌ **Claude Desktop** - Global installation only

## Development Status

### ✅ Implemented (v0.3.0)

- CLI framework (Cobra/Viper)
- Configuration parsing (ait.yml, package.yml, ait.lock)
- `ait init` command
- `ait install` command with project-level and global installation
- `ait list` command
- `ait update` command - Update packages to latest compatible versions
- `ait uninstall` command - Remove packages with --global support
- `ait generate` command - Generate ait.yml from installed packages
- `ait sync` command - Sync project packages to global tools
- **Project-Level Installation** - Tool-native file generation:
  - `.cursorrules` for Cursor (auto-detected)
  - `.github/copilot-instructions.md` for GitHub Copilot (auto-detected)
  - `.opencode/agents/` for OpenCode (proposed standard)
- **Global Installation** - Install to AI tools' global directories
- OpenCode adapter (agents, skills, prompts)
- Cursor adapter (agents, skills, prompts with .cursorrules conversion)
- Claude Desktop adapter (agents, skills, prompts)
- Git-based sources (GitHub, GitLab, generic)
- Local filesystem sources
- Semantic versioning with optional version specifications (defaults to "latest")
- Dependency resolution with cycle detection
- Lock file generation
- Local caching (~/.ait/cache/)
- Test package repository
- Multi-tool detection and installation

### 🔄 Planned

- `ait search` command - Package discovery
- `ait audit` command - Security scanning
- `ait doctor` command - Health checks
- MCP server support
- Package registry/marketplace
- VSCode adapter
- Additional IDE support

## Project Structure

```
ait/
├── cmd/ait/              # CLI entry point
├── internal/
│   ├── adapters/         # Platform adapters
│   │   ├── adapter.go   # Interface
│   │   ├── opencode.go  # OpenCode implementation
│   │   ├── cursor.go    # Cursor implementation
│   │   ├── claude.go    # Claude Desktop implementation
│   │   ├── projectroot.go # Project-root adapter (NEW)
│   │   └── detector.go  # Tool detection
│   ├── cli/              # Commands
│   │   ├── root.go      # Root command
│   │   ├── init.go      # Init command
│   │   ├── install.go   # Install command
│   │   ├── list.go      # List command
│   │   ├── update.go    # Update command
│   │   ├── uninstall.go # Uninstall command
│   │   ├── generate.go  # Generate command (NEW)
│   │   └── sync.go      # Sync command (NEW)
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
├── DETECTION_STRATEGY.md # Tool detection architecture (NEW)
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
    - local:~/my-agents/code-reviewer    # Latest version
    - local:~/my-agents/documentation    # Latest version
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
    - company/agents/security-reviewer@^2.0.0      # Caret constraint
    - company/agents/performance-analyzer          # Latest version
  skills:
    - company/skills/python@^3.0.0
    - company/skills/terraform                     # Latest version
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
    - github:ai-packages/agents/code-reviewer      # Latest version (recommended)
    - github:ai-packages/agents/test-generator@~2.1.0
  skills:
    - github:ai-packages/skills/javascript         # Latest version
  prompts:
    - github:ai-packages/prompts/debug-helper      # Latest version
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
