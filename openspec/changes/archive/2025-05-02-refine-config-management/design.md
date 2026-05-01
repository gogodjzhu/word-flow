## Context

Word-flow's configuration lives in `~/.config/wordflow/config.yaml` (or `$WORDFLOW_HOME`). The current system uses a `Config` struct containing `DictConfig` and `NotebookConfig`, each holding a flat `map[string]interface{}` with dotted string keys (e.g., `"llm.api_key"`). A reflection-based `ConfigMapper` translates these maps into typed endpoint structs. This approach works but has significant drawbacks: typos in keys are silently ignored, `time.Duration` serializes as nanosecond integers, and there is no way to manage config from the CLI. The `Factory` re-reads the config file on every call with no caching. LLM API keys are stored plaintext with no env var fallback.

The design uses pure Go stdlib (`os`, `encoding/yaml`, `reflect`) for all config operations — no third-party config libraries. This keeps dependencies minimal and gives full control over env var mapping and type coercion.

## Goals / Non-Goals

**Goals:**
- Provide a `wordflow config` CLI with `view`, `set`, `get`, `path`, and `init` subcommands
- Add a `--config` global flag for specifying an alternative config file path
- Support environment variable overrides with `WORDFLOW_` prefix
- Validate config on startup with clear error messages for missing required values (e.g., LLM API key)
- Refactor config from flat `map[string]interface{}` with dotted keys to properly nested YAML structs
- Generate a well-documented config file with required values filled in and optional values commented out
- Cache config in `Factory` to avoid repeated file reads

**Non-Goals:**
- Encrypted storage or OS keychain integration for API keys (future consideration)
- Config hot-reload or filesystem watch
- Multi-profile or environment-based config switching
- Web UI for configuration
- Replacing Cobra with a different CLI framework
- Config migration from old format (breaking change; users regenerate via `config init`)

## Decisions

### 1. Nested YAML struct format (replacing flat parameters map)

**Decision**: Replace `DictConfig.parameters` and `NotebookConfig.parameters` maps with properly nested Go structs that map directly to nested YAML.

**Rationale**: Direct struct mapping eliminates the reflection-based mapper, provides compile-time type safety, and produces human-readable YAML. The current flat map with dotted keys (`"llm.api_key"`) is error-prone — typos silently default, and `time.Duration` serializes as nanosecond integers.

**Alternative considered**: Keep the flat map but add stricter validation. Rejected because it doesn't solve the readability and type-safety problems.

### 2. Config struct design — runtime state vs. persisted fields

**Decision**: Fields that are runtime-derived (`basePath`, `configFilename`) are marked `yaml:"-"` and never written to or read from the YAML file. They are populated in memory after loading based on `WORDFLOW_HOME`, `--config` flag, or defaults. Dynamic path defaults (`notebook.basepath`, `ecdict.db_filename`) are left empty in YAML; when empty at load time, they are filled with computed defaults based on `configDir()`.

**Rationale**: Absolute paths like `/home/user/.config/wordflow/...` should never appear in the config file — they break when the file is shared across machines. By computing them at load time, the config file stays portable.

**Dynamic defaults logic** (handled inside `LoadConfig`, not a separate function):
- `llm.url` and `llm.model` → **no default**; required fields that must be explicitly configured
- `notebook.settings.basepath` → `filepath.Join(configDir(), "notebooks")` if empty
- `ecdict.db_filename` → `filepath.Join(configDir(), "stardict.db")` if empty
- `common.basePath` → `configDir()` (always, not from YAML)
- `common.configFilename` → determined by `--config` flag or default path (always, not from YAML)

### 3. Config file generation with comments

**Decision**: When `config init` creates a new config file (or when the program auto-creates one on first run), it writes a template YAML with all required values filled in and optional/sensitive values commented out with inline descriptions.

**Rationale**: A commented config file serves as self-documenting configuration. Users can see every available option, understand what each does, and uncomment the ones they need. This is especially important for `llm.api_key` which must be set before using LLM features.

Generated config file (`wordflow config init`):
```yaml
# Wordflow configuration
version: v1

dict:
  # Default dictionary endpoint. Options: youdao, llm, ecdict, etymonline, mwebster
  default: youdao

  youdao: {}

  llm:
    # LLM provider settings (required if dict.default is llm or using trans command)
    # api_key: ""           # Required. Set via WORDFLOW_DICT_LLM_API_KEY or wordflow config set dict.llm.api_key
    # url: ""               # Required. Full API endpoint URL, not base URL. e.g. https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
    # model: ""             # Required. LLM model name, e.g. glm-4
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3

  ecdict:
    # Local dictionary database path. Defaults to <WORDFLOW_HOME>/stardict.db if empty
    # db_filename: ""

  etymonline: {}

  mwebster:
    # Merriam-Webster API key (required if using mwebster)
    # key: ""

notebook:
  default: default

  settings:
    # Notebook storage path. Defaults to <WORDFLOW_HOME>/notebooks if empty
    # basepath: ""
    max_reviews_per_session: 50
    new_cards_per_day: 20
```

Key decisions about the template:
- `youdao`: empty struct (`{}`), no config fields — Youdao is a free public API, no credentials needed
- `llm.url` and `llm.model`: commented out and required — no sensible default since users choose their own LLM provider and model
- `llm.api_key`: commented out and required — must be configured before using LLM features
- `llm.timeout`, `max_tokens`, `temperature`: filled with sensible defaults that work across providers
- `ecdict.db_filename` and `notebook.settings.basepath`: commented out with dynamic default notes — left empty in YAML, computed at load time
- `etymonline`: empty struct (`{}`), consistent with `youdao`
- `mwebster.key`: commented out, required only when using mwebster endpoint
- `notebook.default`: set to `"default"` matching current behavior

### 4. Config version field

**Decision**: Keep the `version` field in config, set to `"v1"`. This is the first stable config format version. No migration or backward compatibility logic is included — `v1` is simply a marker for the current format.

**Rationale**: Having a version field costs nothing and provides a clear signal for future format changes. The `v` prefix distinguishes it from the old `0.1` format and indicates this is a stable format. If a future `v2` format is introduced, the version field enables clean detection; the current implementation simply validates `version == "v1"` on load or treats a missing version as needing initialization.

### 5. Pure stdlib config loading with env var override

**Decision**: Implement config loading and env var override using Go stdlib (`os`, `encoding/yaml`, `reflect`) without any third-party config library.

**Rationale**: Word-flow's config scope is small — a single YAML file with a handful of nested structs. A custom `LoadConfig` function that reads YAML, unmarshals into structs, then walks `os.Environ()` to apply `WORDFLOW_` overrides via reflection is straightforward and avoids adding a dependency. The `set`/`get` CLI commands can use a key-path parser that walks struct fields by YAML tags.

**Alternative considered**: 
- **Koanf/Viper**: Feature-rich but overkill for this scope; adds transitive dependencies.
- **Keep custom mapper**: Doesn't handle env vars, and the reflection-based approach is fragile.

### 6. Config CLI command design

**Decision**: Add `wordflow config` with subcommands: `view`, `set <key> <value>`, `get <key>`, `path`, `init`.

**Rationale**: Follows the pattern of `git config`, `gh config`, and similar CLI tools. Key names use dot notation (e.g., `dict.llm.api_key`) for consistency with env var mapping (`WORDFLOW_DICT_LLM_API_KEY`).

`set` and `get` use a key-path parser that walks the struct tree by YAML tags to address nested values with dot-notation keys.

### 7. Environment variable override

**Decision**: All config values can be overridden via `WORDFLOW_` prefix env vars. Nested keys use underscore delimiter: `WORDFLOW_DICT_LLM_API_KEY` → `dict.llm.api_key`.

**Rationale**: Environment variables are the standard way to provide secrets (API keys) in CI/CD and container environments. This is critical for the LLM API key which should not be stored in plaintext YAML.

### 8. Startup validation

**Decision**: Add a `Validate()` method on the top-level `Config` struct that checks all active endpoint configs. When `dict.default` is `"llm"`, validate that `llm.api_key` is set (from config or env). Print actionable error messages like `"LLM API key is required when using LLM dictionary. Set it via: wordflow config set dict.llm.api_key <key> or WORDFLOW_DICT_LLM_API_KEY=<key>"`.

**Rationale**: Current behavior silently fails at request time with cryptic errors. Startup validation with guidance improves the first-run experience significantly.

## Risks / Trade-offs

- **Breaking YAML format change** → No migration path; existing configs must be regenerated via `wordflow config init`. This is acceptable for an early-stage project with a small user base.
- **No third-party config library** → Custom env var parsing and key-path logic must handle type coercion (string-to-duration, string-to-int, etc.); test coverage mitigates this risk
- **Dot-notation key addressing may confuse users** → Mitigated by `config view` command showing the full YAML structure and help text showing example key paths
- **Dynamic defaults require runtime computation** → Empty path fields in config file are filled at load time; users who inspect the YAML file won't see the computed values until running `wordflow config view`