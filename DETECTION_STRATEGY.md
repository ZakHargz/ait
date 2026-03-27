# AIT Project-Level Detection Strategy

## Goal
Install AI agents/skills to directories that IDEs/CLIs automatically detect, without requiring manual sync.

## Research: What Each Tool Checks

### Cursor ✅ AUTO-DETECTS
- **Primary**: `.cursorrules` file at project root
- **Behavior**: Automatically loads `.cursorrules` when opening a project
- **AIT Support**: Already implemented - we convert AGENT.md → .cursorrules
- **Team Workflow**: ✅ Commit `.cursorrules` to git, team gets it automatically

### GitHub Copilot ✅ AUTO-DETECTS  
- **Primary**: `.github/copilot-instructions.md` (officially supported since 2024)
- **Behavior**: Automatically loads instructions when GitHub Copilot is active
- **AIT Support**: Should implement - convert AGENT.md → copilot-instructions.md
- **Team Workflow**: ✅ Commit to git, team gets it automatically

### VSCode (with extensions)
- **Settings**: `.vscode/settings.json` can contain AI assistant prompts
- **Extensions**: Various AI extensions check workspace settings
- **AIT Support**: Could generate .vscode/settings.json with AI configs

### OpenCode 🔍 NEEDS RESEARCH
- **Global**: `~/.config/opencode/agents/`, `~/.config/opencode/skills/`
- **Project-Level**: UNKNOWN - does OpenCode check `.opencode/` or workspace config?
- **Question for User**: Does OpenCode support project-level agent detection?

### Claude Desktop ❌ NO PROJECT DETECTION
- **Global Only**: `~/.claude/agents/`
- **Project-Level**: Not supported
- **Workaround**: Use `ait sync` to copy from project to global

## Recommended Implementation Strategy

### Phase 1: Support Tools with Native Project Detection ✅

**For Cursor** (Already Working):
```
project/
├── .cursorrules              # Auto-detected by Cursor
└── ait.yml
```

**For GitHub Copilot** (Should Add):
```
project/
├── .github/
│   └── copilot-instructions.md   # Auto-detected by GitHub Copilot
└── ait.yml
```

**Team Workflow**:
```bash
# Developer 1
ait init
ait install github:org/repo/agents/code-reviewer@1.0.0
git add .cursorrules .github/ ait.yml ait.lock
git commit -m "Add AI code reviewer"
git push

# Developer 2  
git clone repo
# Cursor and GitHub Copilot automatically pick up the agents!
# No ait command needed!
```

### Phase 2: Tools Without Project Detection

**For OpenCode, Claude, etc.** (Need Sync):
```
project/
├── .ait/                     # Local cache (gitignored)
│   ├── agents/
│   ├── skills/
│   └── prompts/
├── .cursorrules              # Committed (auto-detected)
├── .github/
│   └── copilot-instructions.md  # Committed (auto-detected)
└── ait.yml                   # Committed
```

**Team Workflow**:
```bash
# Developer 1
ait install github:org/repo/agents/code-reviewer@1.0.0
# Creates: .cursorrules ✅ auto-detected
#          .github/copilot-instructions.md ✅ auto-detected  
#          .ait/agents/code-reviewer (cache)

# For OpenCode/Claude (if you want them):
ait sync --target opencode
# Copies .ait/* to ~/.config/opencode/

# Commit only the auto-detected files
git add .cursorrules .github/ ait.yml ait.lock
git commit -m "Add AI agents"
```

### Phase 3: Future - Propose Standards

Work with OpenCode/Claude teams to support project-level detection:
- `.opencode/agents/` at project root
- `.claude/agents/` at project root
- Make it official so no sync needed

## Implementation Priority

1. ✅ **DONE**: Cursor support via .cursorrules
2. **TODO**: GitHub Copilot support via .github/copilot-instructions.md
3. **TODO**: VSCode settings.json support
4. ✅ **DONE**: Sync command for tools without project detection
5. **FUTURE**: Advocate for project-level support in OpenCode/Claude

## Key Insight

**Don't fight against tool conventions - embrace them!**

- Tools with project detection: Use their native paths (`.cursorrules`, `.github/copilot-instructions.md`)
- Tools without project detection: Use sync command as bridge
- Make the common case (Cursor, GitHub Copilot) zero-friction
