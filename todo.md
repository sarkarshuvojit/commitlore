## Background

LLM providers are currently statically configured and referenced as `llmProvider` and `llmProviderType` in the TUI's `BaseModel` (`internal/tui/models.go`). The status bar displays the active provider, but there is no dynamic management, listing, or runtime selection of LLM providers.

There will be three types of Providers
- API Provider (API Interfaces for 3rd Party LLM Providers)
- CLI Provider (This is the type for claide CLI, and will eventually add Gemini CLI, AWS CLI etc)
- Local Models (This will be to interface with ollama APIs)

## Feature Goals

0.1. Create a config to manage all providers, also have a flag to disable providers so that user can't pick them from the menu; this will help us show future support for providers

1. **Provider Management Menu in TUI**:
   - Add a top-level "Providers" menu item to the Bubble Tea-based interface.
   - Entering this menu should show a list of all configured LLM providers (e.g., OpenAI, Anthropic, Gemini, mock, etc.).
   - Allow users to:
     - Switch Providers
     - In the list show all providers
     - Detect API Keys in ENV for API based Providers / This is for OpenAI, Anthopic etc
       -  Detect ollama installed in directory and show it as a local provider
     - Show disabled providers ase greyed out and don't let user select them

2. **Provider Selection**:
   - Allow the user to select one active LLM provider per session.
   - The selected provider should be displayed in the main status bar (replacing the current static display).
   - All LLM operations (content generation, topic extraction, etc.) must use the active provider.

3. **Persistence**:
   - Persist the list of configured providers and the last active provider (user’s home directory).
   - On app start, restore providers and automatically select the last used provider if available.

4. **Extensibility**:
   - Design the provider management system to easily support new LLMs in future (add provider type, config params, etc.).
   - UI must validate provider configuration (e.g., required API keys).

5. **Code Integration Points**:
   - Update `AppModel`/`BaseModel` to support a provider list and switching logic.
   - Refactor TUI views to enable provider management and selection.
   - Ensure all LLM usage routes through the selected provider.

6. Do not implement a new provider, rather just add switching menu for current ones available, and for new ones just keep the structure available so that later we can 
  - Implement a new provider struct
  - Add to config or enable existing

## References

- `internal/core/llm/interface.go` (provider interface)
- `internal/tui/models.go` (`BaseModel`, `AppModel`, usage of `llmProvider`)
- `internal/tui/listing_model.go` (status bar shows current provider)
- Current provider is not persisted or runtime-editable.

## Acceptance Criteria

- Users can manage (add/edit/delete) LLM providers from within the TUI.
- Users can select the active provider at runtime.
- The selected provider is shown in the status bar and used for all content generation.
- Provider configuration persists across sessions.
- New LLM provider types can be added with minimal changes.

---
This enhancement will make CommitLore more flexible and user-friendly, allowing users to easily experiment with and switch between different LLM backends.

