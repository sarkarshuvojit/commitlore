# CommitLore 🔮

> Transform your Git history into compelling developer stories

[![asciicast](https://asciinema.org/a/729244.svg)](https://asciinema.org/a/729244)

*Watch a quick demo of CommitLore in action!*

CommitLore is an open-source TUI (Terminal User Interface) tool that analyzes your Git commit history and transforms it into engaging content. Turn your development journey into blog posts, social media content, technical narratives, and documentation—all powered by your actual code contributions.

## ✨ Why CommitLore?

Every commit tells a story. CommitLore helps developers:
- **Showcase expertise** through authentic technical content
- **Build personal brand** with content rooted in real work
- **Document learning journeys** and technical decisions
- **Create portfolio content** that demonstrates growth
- **Share knowledge** with the developer community

## 🚀 Key Features

- **🔍 Smart Git Analysis**: Parses commit history to identify patterns and insights
- **🎯 Topic Extraction**: Automatically identifies key technologies and learning moments
- **📝 Multi-Format Content**: Generates blog posts, social media content, and technical narratives
- **🎨 Interactive TUI**: Beautiful terminal interface built with Bubble Tea
- **🔌 BYOL Architecture**: Bring Your Own LLM - works with various AI providers
- **💾 Export Options**: Save content in multiple formats (Markdown, etc.)

## 🛠 Use Cases

### For Individual Developers
- **Portfolio Building**: Transform commits into case studies and project showcases
- **Blog Content**: Generate technical blog posts from development experiences
- **Social Media**: Create authentic developer content for LinkedIn, Twitter, etc.
- **Learning Documentation**: Track and share your technical growth journey

### For Tech Leads & Senior Engineers
- **Architecture Documentation**: Document design decisions and technical choices
- **Team Knowledge Sharing**: Create content from architectural commits and refactoring
- **Technical Leadership**: Showcase problem-solving approaches and best practices

### For Developer Advocates
- **Authentic Content**: Create technical content based on real development work
- **Community Engagement**: Share genuine developer experiences and insights
- **Tutorial Creation**: Turn feature implementations into step-by-step guides

### For Open Source Maintainers
- **Project Evolution**: Document how your project has grown and changed
- **Feature Announcements**: Create content around new features and improvements
- **Community Updates**: Share progress and milestones with your community

## 🤖 BYOL (Bring Your Own LLM)

CommitLore follows a **Bring Your Own LLM** philosophy, giving you the flexibility to choose your preferred AI provider. Connect to various LLM services based on your needs, budget, and preferences.

### Supported LLM Providers

| Provider | Type | Status | Notes |
|----------|------|---------|-------|
| **Claude API** | Cloud API | ✅ Supported | Anthropic's Claude via API |
| **Claude CLI** | Local CLI | ✅ Supported | Official Claude CLI tool |
| **OpenAI API** | Cloud API | 🔄 Planned | GPT-4, GPT-3.5-turbo |
| **Ollama** | Local | 🔄 Planned | Local LLM inference |
| **LM Studio** | Local | 🔄 Planned | Local model management |
| **Gemini API** | Cloud API | 🔄 Planned | Google's Gemini |
| **Azure OpenAI** | Cloud API | 🔄 Planned | Enterprise OpenAI |
| **Hugging Face** | Cloud API | 🔄 Planned | Open model ecosystem |

### LLM Integration Types

- **🌐 Cloud APIs**: Connect to hosted LLM services (OpenAI, Claude, Gemini)
- **🖥️ Local Models**: Use local inference engines (Ollama, LM Studio)
- **⚙️ CLI Tools**: Integrate with command-line AI tools
- **🔗 Proxy Services**: Connect through LLM gateways and proxies

## 🏗 Installation

### Prerequisites

- **Go 1.21+**
- **Git** (for repository analysis)
- **Terminal** with TTY support

### Install from Source

```bash
# Clone the repository
git clone https://github.com/sarkarshuvojit/commitlore.git
cd commitlore

# Build the binary
go build -o commitlore main.go

# Run CommitLore
./commitlore
```

### Install with Go

```bash
go install github.com/sarkarshuvojit/commitlore@latest
```

## 🎮 Usage

### Basic Usage

1. **Navigate to your Git repository**
   ```bash
   cd your-project
   ```

2. **Launch CommitLore**
   ```bash
   commitlore
   ```

3. **Follow the interactive prompts**:
   - Select commit range for analysis
   - Choose your LLM provider
   - Configure content generation preferences
   - Generate and refine content

### LLM Provider Setup

#### Claude API
```bash
export CLAUDE_API_KEY="your-api-key-here"
```

#### Claude CLI
```bash
# Install Claude CLI first
npm install -g @anthropic-ai/claude-cli

# CommitLore will automatically detect the CLI
```

### Command Line Options

```bash
# Analyze specific commit range
commitlore --from=HEAD~10 --to=HEAD

# Use specific LLM provider
commitlore --llm=claude-api

# Generate specific content type
commitlore --format=blog-post

# Export to file
commitlore --output=my-story.md
```

## 🏛 Architecture

CommitLore is built with a clean, modular architecture:

```
├── internal/
│   ├── core/              # Core business logic
│   │   ├── git.go         # Git repository analysis
│   │   ├── llm/           # LLM provider interfaces
│   │   │   ├── interface.go
│   │   │   ├── claude_api.go
│   │   │   ├── claude_cli.go
│   │   │   └── async.go
│   │   └── logger.go      # Logging utilities
│   └── tui/               # Terminal UI components
│       ├── app.go         # Main application
│       ├── models.go      # Data models
│       ├── listing_model.go
│       ├── topic_model.go
│       ├── content_model.go
│       └── styles.go      # UI styling
└── main.go                # Application entry point
```

### Key Design Principles

- **🔌 Pluggable LLM System**: Easy to add new LLM providers
- **📱 Responsive TUI**: Clean, keyboard-driven interface
- **🔄 Async Processing**: Non-blocking LLM operations
- **🧪 Testable Core**: Business logic separated from UI
- **⚡ Performance**: Efficient Git analysis and content generation

## 🛠 Development

### Building from Source

```bash
# Clone and enter directory
git clone https://github.com/sarkarshuvojit/commitlore.git
cd commitlore

# Install dependencies
go mod download

# Run in development mode
go run main.go

# Build binary
go build -o commitlore main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/core/...
```

### Adding New LLM Providers

1. Implement the `LLMProvider` interface in `internal/core/llm/`
2. Add configuration options
3. Update the provider selection logic
4. Add tests for your implementation

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Development Setup

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Add tests** for new functionality
5. **Run tests** to ensure everything works
6. **Submit a pull request**

### Contribution Guidelines

- **Code Quality**: Follow Go best practices and conventions
- **Testing**: Add tests for new features and bug fixes
- **Documentation**: Update documentation for new features
- **Commit Messages**: Use clear, descriptive commit messages
- **Issue First**: For major changes, open an issue first to discuss

### Areas for Contribution

- **🔌 New LLM Providers**: Add support for additional AI services
- **🎨 UI Improvements**: Enhance the terminal interface
- **📝 Content Templates**: Create new content generation templates
- **🐛 Bug Fixes**: Fix issues and improve stability
- **📚 Documentation**: Improve docs and examples

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Support

- **⭐ Star this repo** if you find it useful
- **🐛 Report issues** on GitHub
- **💬 Join discussions** in our community
- **📖 Read the docs** for detailed usage guides

## 🔮 Roadmap

### Near Term (v1.0)
- [ ] OpenAI API integration
- [ ] Local LLM support (Ollama)
- [ ] Enhanced content templates
- [ ] Export format options

### Medium Term (v2.0)
- [ ] Web interface
- [ ] Team collaboration features
- [ ] Analytics and insights
- [ ] Custom prompt engineering

### Long Term (v3.0+)
- [ ] IDE integrations
- [ ] CI/CD pipeline integration
- [ ] Enterprise features
- [ ] Advanced analytics

---

**CommitLore** - Where every commit tells a story worth sharing. 🔮✨

*Built with ❤️ by the developer community, for the developer community.*

