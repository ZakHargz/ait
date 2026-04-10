# AIT - AI Toolkit Package Manager

A CLI package manager for AI agents, skills, prompts, and MCP servers. Think npm/pip for AI tooling across OpenCode, Cursor, Claude Desktop, and more.

> **🔄 Aligning with Standards**: AIT is actively aligning with [Microsoft's Agent Package Manager (APM)](https://microsoft.github.io/apm/) format and ecosystem standards including [AGENTS.md](https://agents.md), [Agent Skills](https://agentskills.io), and [Model Context Protocol (MCP)](https://modelcontextprotocol.io). This ensures maximum compatibility and interoperability across AI development tools.

## Features

- 📦 **Declarative Dependencies** - Define AI tools in `ait.yml`
- 🚀 **Project-Level Installation** - Install to tool-native paths (`.cursorrules`, `.github/copilot-instructions.md`) for automatic detection
- 🔄 **Multi-Tool Sync** - Install to OpenCode, Cursor, Claude Desktop, GitHub Copilot simultaneously
- 👥 **Team Sharing** - Commit tool-native files to git for instant team collaboration
- 🌲 **Dependency Resolution** - Automatic transitive dependency handling
- 📌 **Lock Files** - Reproducible installations with `ait.lock`
- 🏷️ **Semantic Versioning** - Optional version constraints (`^1.0.0`, `~2.1.0`, or omit for latest)
- 🌍 **Multiple Sources** - GitHub, GitLab, generic Git, or local packages
- 🔐 **Private Repository Support** - Authenticate with GitHub tokens for private packages ([docs](docs/AUTHENTICATION.md))
- 💾 **Smart Caching** - Local cache to speed up installations

## Installation

### Using Homebrew (macOS/Linux) - Recommended

```bash
# Add the AIT tap
brew tap zakhargz/ait

# Install AIT
brew install ait

# Verify installation
ait --version
```

### From Source

```bash
# Clone the repository
git clone https://github.com/ZakHargz/ait
cd ait

# Build and install globally
go build -o /usr/local/bin/ait cmd/ait/main.go

# Or build without installing
go build -o bin/ait cmd/ait/main.go
./bin/ait --version
```

### Requirements

- Go 1.26+ (for building from source only - not needed for Homebrew installation)
- One or more AI tools: OpenCode, Cursor, Claude Desktop, or GitHub Copilot

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
dependencies: []
targets:
  - opencode
```

### 2. Authentication for Private Repositories (Optional)

If you're using private GitHub repositories, set up authentication:

```bash
# Use GitHub CLI (easiest)
gh auth login
export GH_TOKEN=$(gh auth token)

# Or set a personal access token
export GITHUB_TOKEN=ghp_your_token_here
```

For detailed authentication setup, see [Authentication Guide](docs/AUTHENTICATION.md).

### 3. Add Dependencies

Edit `ait.yml` to add packages as a simple flat list:

```yaml
name: my-project
version: 1.0.0
dependencies:
  # GitHub shorthand (no prefix needed - defaults to GitHub)
  - org/repo/agents/code-reviewer           # Latest version (recommended)
  - org/repo/agents/test-generator@1.0.0    # Specific version
  - org/repo/skills/python                  # Latest version
  - org/repo/skills/docker@^2.0.0           # Version constraint (^, ~)
  - org/repo/prompts/debug@~1.5.0
  
  # Virtual packages (single files)
  - org/awesome/agents/reviewer.agent.md@1.0.0
  - org/awesome/skills/python.skill.md
  
  # Other Git hosts (use FQDN)
  - gitlab.com/myorg/packages/agents/helper
  
  # Local packages
  - ./my-packages/custom-agent
```

**Pro Tip**: Omit `@version` to automatically get the latest version!

<details>
<summary><b>Legacy Format (still supported)</b></summary>

The old format with separate `agents`, `skills`, and `prompts` lists is still supported for backward compatibility:

```yaml
dependencies:
  agents:
    - github:org/repo/agents/code-reviewer
  skills:
    - github:org/repo/skills/python
  prompts:
    - github:org/repo/prompts/debug
```

Note: The `github:` prefix is required in legacy format. We recommend migrating to the flat list format above.
</details>

### 4. Install Packages

```bash
# Install specific packages (automatically creates/updates ait.yml)
ait install org/repo/agents/code-reviewer
ait install org/repo/skills/python@^2.0.0

# Install all dependencies from ait.yml
ait install

# Install without saving to ait.yml
ait install org/repo/agents/reviewer --save=false

# Install globally to AI tools instead
ait install --global org/repo/agents/reviewer

# Install to specific tools only
ait install --target opencode org/repo/agents/reviewer
```

**New APM-like Behavior**: When you install a package from the command line, AIT automatically:
1. Creates `ait.yml` if it doesn't exist (using current directory name as project name)
2. Adds the package to the dependencies list
3. Creates tool-native files at your project root

This creates tool-native files at your project root:
- `.cursorrules` - Auto-detected by Cursor
- `.github/agents/*.agent.md` - Auto-detected by GitHub Copilot, VS Code, IntelliJ (APM standard)
- `.opencode/agents/` - For OpenCode (proposed standard)

**Team Workflow**: Commit these files to git! Your team gets AI agents automatically when they clone the repo - no `ait` commands needed.

### 5. Share with Your Team (Optional)

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

### 6. List Installed Packages

```bash
# List all installed packages
ait list

# List for specific tool
ait list --target opencode
```

### 7. Update Packages

```bash
# Update all packages to latest compatible versions
ait update

# Update specific packages
ait update code-reviewer python-skill

# Update for specific tools only
ait update --target opencode
```

### 8. Uninstall Packages

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

Dependencies are specified as a simple flat list with GitHub shorthand by default:

```yaml
dependencies:
  # GitHub (default, no prefix needed)
  - org/repo/agents/code-reviewer@1.0.0
  - org/repo/skills/python               # Latest version
  
  # Other Git hosts (use FQDN)
  - gitlab.com/org/repo/agents/helper@2.0.0
  - bitbucket.org/org/repo/skills/docker
  
  # Virtual packages (single files)
  - org/repo/agents/reviewer.agent.md
  - org/repo/skills/python.skill.md
  - org/repo/prompts/debug.prompt.md
  
  # Local packages
  - ./path/to/package
  - ~/packages/agents/custom
```

**Virtual Packages**: Files ending in `.agent.md`, `.skill.md`, `.prompt.md`, `.instructions.md`, or `.chatmode.md` can be installed directly without requiring a `package.yml` file.

### Legacy Format (Still Supported)

The original format with explicit source type prefixes and nested structure:

```yaml
dependencies:
  agents:
    - github:org/repo/agents/code-reviewer@1.0.0
  skills:
    - gitlab:org/repo/skills/python@^2.0.0
  prompts:
    - local:./my-prompts/debug
```

**Supported source prefixes in legacy format:**
- `github:org/repo/path/to/package@version` - GitHub repositories
- `gitlab:org/repo/path/to/package@version` - GitLab repositories
- `git:https://git.example.com/repo/path@version` - Generic Git URLs
- `local:./path/to/package` - Local filesystem paths

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
dependencies:           # Optional: Package dependencies (flat list)
  - org/repo/skills/git-workflow@^1.0.0
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

dependencies:
  - org/repo/agents/code-reviewer                    # Latest version (recommended)
  - myorg/ai-toolkit/agents/custom-agent@~2.0.0      # Version constraint
  - org/repo/skills/python                           # Latest version
  - org/repo/agents/reviewer.agent.md                # Virtual package
  - ./prompts/custom                                 # Local package
  - org/repo/servers/custom-mcp@^1.0.0               # MCP server

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
# Install specific packages (auto-creates/updates ait.yml)
ait install org/repo/agents/reviewer
ait install org/repo/skills/python@^2.0.0

# Install all dependencies from ait.yml
ait install

# Install without saving to ait.yml
ait install org/repo/agents/reviewer --save=false

# Install globally to AI tools
ait install --global org/repo/agents/reviewer

# Install to specific tools
ait install --target opencode --target cursor org/repo/agents/reviewer
```

**Flags:**
- `--global` or `-g` - Install to AI tools globally instead of project root
- `--target` or `-t` - Specify which tools to install to
- `--save` or `-s` - Add installed packages to ait.yml (default: true, use --save=false to disable)

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
- **Project-Level**: `.github/agents/<name>.agent.md` (auto-detected by GitHub Copilot, VS Code, IntelliJ)
- **Format**: Native `.agent.md` format (APM standard)
- **Auto-Detection**: ✅ Yes - GitHub Copilot automatically loads agents from `.github/agents/`

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
- ✅ **GitHub Copilot** - Full auto-detection via `.github/agents/*.agent.md` (APM standard)
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
  - `.github/agents/*.agent.md` for GitHub Copilot, VS Code, IntelliJ (APM standard)
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
  - ~/my-agents/code-reviewer    # Local path, latest version
  - ~/my-agents/documentation    # Local path, latest version
targets:
  - opencode
```

### Example 2: Team Toolkit

```yaml
# ait.yml
name: team-project
version: 1.0.0
dependencies:
  - mycompany/ai-toolkit/agents/security-reviewer@^2.0.0       # Caret constraint
  - mycompany/ai-toolkit/agents/performance-analyzer           # Latest version
  - mycompany/ai-toolkit/skills/python@^3.0.0
  - mycompany/ai-toolkit/skills/terraform                      # Latest version
targets:
  - opencode
  - cursor
```

### Example 3: Public Packages with Virtual Files

```yaml
# ait.yml
name: open-source-project
version: 1.0.0
dependencies:
  - ai-packages/agents/code-reviewer                    # Latest version (recommended)
  - ai-packages/agents/test-generator@~2.1.0
  - ai-packages/skills/javascript                       # Latest version
  - ai-packages/prompts/debug-helper                    # Latest version
  - ai-packages/agents/reviewer.agent.md@1.0.0          # Virtual package (single file)
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

## Relationship with Microsoft APM

AIT and [Microsoft's Agent Package Manager (APM)](https://microsoft.github.io/apm/) share the same vision: **making AI agent configuration portable, versioned, and easy to share**. While inspired by APM, AIT takes a simpler approach optimized for multi-tool deployment.

### How AIT Differs

- **Multi-tool focus**: AIT deploys to tool-native paths (`.cursorrules`, `.github/copilot-instructions.md`) that work across Cursor, GitHub Copilot, OpenCode, and Claude Desktop
- **Project-first**: Default behavior is project-level installation for team sharing
- **Simpler format**: Flat dependency list instead of nested structure
- **Lightweight**: Single binary, no complex compilation step

### Format Comparison

**AIT Format (Simpler):**
```yaml
name: my-project
version: 1.0.0
dependencies:
  - org/repo/agents/code-reviewer          # Flat list
  - org/repo/skills/python
  - gitlab.com/myorg/packages/agents/helper
  - org/repo/agents/reviewer.agent.md      # Virtual packages
```

**Microsoft APM Format:**
```yaml
name: my-project
version: 1.0.0
dependencies:
  apm:  # Nested structure
    - org/repo/agents/code-reviewer
    - org/repo/skills/python
```

Both formats support:
- ✅ GitHub shorthand (`org/repo/path`)
- ✅ FQDN for other hosts (`gitlab.com/...`)
- ✅ Virtual packages (`.agent.md`, `.skill.md`, etc.)
- ✅ Semantic versioning

### Backward Compatibility

AIT maintains backward compatibility with:
- Legacy `github:org/repo` prefix format
- Nested `agents/skills/prompts` structure
- APM-style `dependencies.apm` format

This allows gradual migration without breaking existing manifests.

### Why Both Projects Matter

- **APM** - Microsoft-backed, comprehensive, enterprise-focused, deep GitHub integration
- **AIT** - Community-driven, lightweight, multi-tool support, simpler format

We believe in ecosystem standards and have aligned with APM's core concepts (GitHub shorthand, virtual packages, semantic versioning) while maintaining AIT's unique value proposition of project-native tool detection and multi-tool support.

## License

MIT

## Acknowledgments

Built on community standards:
- [AGENTS.md](https://agents.md) - Agent definition format
- [Agent Skills](https://agentskills.io) - Skill specification
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io) - Tool integration standard

Inspired by:
- [Microsoft APM](https://microsoft.github.io/apm/) - Agent Package Manager
- npm (Node.js package manager)
- pip (Python package manager)
