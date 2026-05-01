## Why

The current configuration system uses a flat `map[string]interface{}` with dotted string keys (e.g., `"llm.api_key"`), which lacks type safety, silently ignores typos, and provides no IDE autocompletion for users editing YAML. There is no CLI command to view or modify configuration, forcing users to manually locate and edit `~/.config/wordflow/config.yaml`. LLM API keys are stored in plaintext with no environment variable fallback, and misconfigurations are only caught at usage time rather than on startup. The reflection-based mapper, while functional, creates inconsistent access patterns between `DictConfig` and `NotebookConfig` and produces unreadable `time.Duration` serialization (nanoseconds instead of `"30s"`).

## What Changes

- Add a `wordflow config` command with subcommands: `view` (show current config), `set <key> <value>` (set a config value), `get <key>` (get a config value), `path` (show config file location), and `init` (initialize config with defaults)
- Add a `--config` global flag to specify an alternative config file path
- Add environment variable overrides for config values (e.g., `WORDFLOW_LLM_API_KEY`, `WORDFLOW_LLM_URL`)
- Add startup validation that checks all configured endpoints (missing API keys produce clear guidance rather than cryptic runtime errors)
- Refactor config structure from flat `map[string]interface{}` with dotted keys to properly nested YAML structs, eliminating the reflection-based mapper in favor of direct struct unmarshaling
- **BREAKING**: Replace config YAML schema — flat `parameters` map with dotted keys becomes nested struct format; existing config files must be regenerated via `wordflow config init`; the new format uses `version: v1`
- Config file template with comments: required values filled in, optional/sensitive values commented out with inline descriptions
- Dynamic default paths: `notebook.basepath` and `ecdict.db_filename` auto-compute from `configDir()` when left empty in YAML; `common.basePath` and `common.configFilename` are runtime-only (`yaml:"-"`)

## Capabilities

### New Capabilities

- `config-cli`: CLI commands for viewing, setting, getting, and initializing configuration values; includes `--config` global flag
- `config-env-override`: Environment variable override support for all config values with `WORDFLOW_` prefix
- `config-validation`: Startup-time validation of configuration with clear error messages for missing or invalid values (especially LLM API keys)

### Modified Capabilities


## Impact

- **Breaking YAML schema change**: Existing config files in the old flat `parameters` format will no longer be compatible; users regenerate via `wordflow config init`
- **Internal config package** (`internal/config/`): Major refactor — `DictConfig`, `NotebookConfig` structs and the reflection-based `ConfigMapper` will be replaced with nested structs and direct YAML unmarshaling
- **CLI commands** (`pkg/cmd/`): New `config` command package; all existing commands gain `--config` global flag
- **Factory** (`pkg/cmdutil/factory.go`): Cache config after first read; support `--config` flag
- **LLM client** (`internal/llm/`): Consume config from new nested struct format; support env var override for API keys
- **Dependencies**: No new third-party dependencies; env var override and config merging will be implemented in pure Go using stdlib