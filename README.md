# AIT - AI Toolkit Package Manager

A package manager for AI agents, skills, prompts, and MCP servers.

## Features

- 📦 Declare AI dependencies in `ait.yml`
- 🔄 Automatic installation and syncing to AI tools (OpenCode, Cursor, Claude, etc.)
- 🌲 Transitive dependency resolution
- 📌 Lock files for reproducible installations
- 🔐 Security auditing
- 🔍 Package search and discovery

## Installation

```bash
# Clone the repository
git clone https://github.com/apex-ai/ait
cd ait

# Build
make build

# Install
make install
```

## Quick Start

```bash
# Initialize a new project
ait init

# Add dependencies to ait.yml
# ...

# Install packages
ait install

# List installed packages
ait list
```

## Development Status

🚧 **Early Development** - This is an MVP implementation with basic features.

### Implemented
- ✅ CLI framework (cobra/viper)
- ✅ Config parsing (ait.yml, package.yml)
- ✅ `ait init` command
- ✅ OpenCode adapter
- ✅ Basic package model

### In Progress
- 🔄 `ait install` command
- 🔄 Git-based package sources
- 🔄 Dependency resolution

### Planned
- ⏳ Cursor adapter
- ⏳ Claude Desktop adapter
- ⏳ Semantic versioning
- ⏳ Lock files
- ⏳ Security auditing
- ⏳ Package search

## Project Structure

```
ait/
├── cmd/ait/              # CLI entry point
├── internal/
│   ├── adapters/         # Platform adapters (OpenCode, Cursor, etc.)
│   ├── cli/              # CLI commands
│   ├── config/           # Configuration models
│   ├── packages/         # Package management
│   ├── resolver/         # Dependency resolution
│   ├── sources/          # Package sources (Git, etc.)
│   ├── security/         # Security features
│   └── utils/            # Utilities
├── docs/                 # Documentation
└── scripts/              # Build/install scripts
```

## License

MIT
