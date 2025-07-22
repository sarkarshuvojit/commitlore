# CommitLore

**Turn your Git commits into compelling content**

[![Demo](https://asciinema.org/a/729244.svg)](https://asciinema.org/a/729244)

CommitLore analyzes your Git history and transforms it into blog posts, social media content, and technical documentation using your choice of AI provider.

## What it does

- **Analyzes Git commits** to extract meaningful patterns and insights
- **Generates content** in multiple formats (blog posts, Twitter threads, LinkedIn posts, technical docs)
- **Works with any LLM** - bring your own OpenAI, Claude, or local model
- **Interactive terminal UI** for easy navigation and content creation

## Supported LLM Providers

| Provider | Status | Setup |
|----------|--------|-------|
| **Claude API** | ✅ Ready | `export CLAUDE_API_KEY="your-key"` |
| **Claude CLI** | ✅ Ready | `npm install -g @anthropic-ai/claude-cli` |
| **OpenAI API** | ✅ Ready | `export OPENAI_API_KEY="your-key"` |
| **Ollama** | 🔄 Planned | Local inference |
| **Gemini** | 🔄 Planned | Google's API |

## Installation

**Prerequisites:** Go 1.21+, Git

```bash
# Install with Go
go install github.com/sarkarshuvojit/commitlore@latest

# Or build from source
git clone https://github.com/sarkarshuvojit/commitlore.git
cd commitlore && go build -o commitlore main.go
```

## Quick Start

1. **Navigate to your Git repo** and run:
   ```bash
   commitlore
   ```

2. **Set up your LLM** (choose one):
   ```bash
   # Claude API
   export CLAUDE_API_KEY="your-key"
   
   # Claude CLI  
   npm install -g @anthropic-ai/claude-cli
   
   # OpenAI
   export OPENAI_API_KEY="your-key"
   ```

3. **Follow the interactive prompts** to select commits, choose content format, and generate your content.

## Architecture

```
internal/
├── core/           # Business logic
│   ├── git.go      # Repository analysis
│   └── llm/        # AI provider integrations
└── tui/            # Terminal interface
    ├── app.go      # Main application
    └── *_model.go  # Screen models
```

**Design principles:** Pluggable LLM system, testable core logic, responsive keyboard-driven UI.

## Development

```bash
# Build and run
go run main.go

# Run tests
go test ./...

# Add dependencies
go get <package> && go mod tidy
```

### Adding LLM Providers

1. Implement `LLMProvider` interface in `internal/core/llm/`
2. Add provider to selection logic
3. Update documentation

## Contributing

1. **Fork** and create a feature branch
2. **Add tests** for new functionality  
3. **Follow Go conventions** and existing patterns
4. **Submit PR** with clear description

**Areas needing help:** New LLM providers, UI improvements, content templates, bug fixes.

## License

MIT License - see [LICENSE](LICENSE) file.

---

**Made for developers who want to share their coding journey.** ⭐ Star if useful!

