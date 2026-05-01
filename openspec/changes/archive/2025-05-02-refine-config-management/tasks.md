## 1. Define New Config Structs & Helpers

- [x] 1.1 Define new nested config structs in `internal/config/config.go`: replace `DictConfig.parameters` map with typed nested structs (`YoudaoEndpointConfig`, `LLMEndpointConfig`, `EcdictConfig`, `EtymonlineConfig`, `MWebsterConfig`) under a `DictEndpoints` struct; replace `NotebookConfig.parameters` map with typed `NotebookSettings` struct; set version field to `"v1"`
- [x] 1.2 Mark `CommonConfig` fields (`basePath`, `configFilename`) as `yaml:"-"` — they are runtime state, not persisted in YAML
- [x] 1.3 Add `time.Duration` YAML marshal/unmarshal helpers to serialize as human-readable strings (e.g., `"30s"`) instead of nanosecond integers
- [x] 1.4 Implement dynamic default filling inside `LoadConfig`: when `notebook.settings.basepath` or `dict.ecdict.db_filename` are empty after unmarshaling, compute defaults from `configDir()`; populate `CommonConfig` from `configDir()` and config path

## 2. Config Loading with Stdlib

- [x] 2.1 Implement `LoadConfig(path string) (*Config, error)` using pure stdlib: read YAML file, unmarshal into nested structs, apply dynamic defaults for empty path fields, then apply env var overrides by walking `os.Environ()` for `WORDFLOW_` prefix and mapping to struct fields via reflection on YAML tags
- [x] 2.2 Implement env var key mapping: strip `WORDFLOW_` prefix, convert to lowercase, map underscores to dots (e.g., `WORDFLOW_DICT_LLM_API_KEY` → `dict.llm.api_key`), coerce string values to correct types (duration, int, float)
- [x] 2.3 Update `Factory` in `pkg/cmdutil/factory.go` to cache the loaded config after first read instead of re-reading on every call
- [x] 2.4 Add `--config` global flag to root command (`pkg/cmd/root/root.go`) that passes the custom path to `Factory` and then to `LoadConfig`

## 3. Config Validation on Startup

- [x] 3.1 Implement `Config.Validate() error` method on the top-level `Config` struct that validates the active endpoint (based on `dict.default`)
- [x] 3.2 Validate config version: check `version` field is `"v1"`, print error suggesting `wordflow config init` if missing or wrong
- [x] 3.3 Add LLM-specific validation: when `dict.default` is `"llm"` or the command is `trans`, require `api_key` to be non-empty (from config or env var)
- [x] 3.4 Add validation for `llm.temperature` range (0-2), `llm.max_tokens` > 0, `llm.url` and `llm.model` non-empty
- [x] 3.5 Add validation for other active endpoints (e.g., `ecdict.db_filename` when `ecdict` is default)
- [x] 3.6 Provide actionable error messages that include the `wordflow config set` command or env var alternative
- [x] 3.7 Wire validation into `Factory` so it runs before command execution; skip validation for `config` subcommands themselves

## 4. Config CLI Command

- [x] 4.1 Create `pkg/cmd/config/config.go` with Cobra `ConfigCmd` and register under root
- [x] 4.2 Implement `wordflow config view` — loads config, applies dynamic defaults and env var overrides, prints full config as YAML to stdout
- [x] 4.3 Implement `wordflow config get <key>` — retrieves a single value by dot-notation key using struct field walker; resolves dynamic defaults for empty fields; prints error for unknown keys
- [x] 4.4 Implement `wordflow config set <key> <value>` — updates a config value by dot-notation key and saves to file; validates the key is a known config path; validates the value type
- [x] 4.5 Implement `wordflow config path` — prints the resolved config file absolute path
- [x] 4.6 Implement `wordflow config init` — creates a default config file with required values filled in and optional/sensitive values commented out with inline descriptions
- [x] 4.7 Ensure `config` subcommands work with `--config` global flag for alternative config paths

## 5. Config File Template

- [x] 5.1 Create a config template string/constant in `internal/config/` that produces the commented YAML with version `v1`, filler defaults (youdao/etymonline as `{}`, llm.url and llm.model commented out as required fields, llm.api_key commented out, path fields commented out with dynamic default notes), and numeric defaults filled in (timeout, max_tokens, temperature, fsrs settings)
- [x] 5.2 Use the template in both `config init` command and `InitConfig` (auto-create on first run)

## 6. Update Existing Code to Use New Config Structs

- [x] 6.1 Update `pkg/dict/llm/dict_llm.go` to consume `LLMEndpointConfig` directly instead of using `ConfigMapper`
- [x] 6.2 Update `pkg/cmd/trans/trans.go` to consume `LLMEndpointConfig` directly instead of using `ConfigMapper`
- [x] 6.3 Update `pkg/cmd/dict/dict.go` to use new `DictConfig` struct for endpoint selection and config retrieval
- [x] 6.4 Update `pkg/cmd/dict/notebook.go` to use new `NotebookConfig` struct for settings retrieval
- [x] 6.5 Remove `internal/config/mapper.go` and the `ConfigMapper` entirely; remove `GetDefaults()` methods that were only used by the mapper
- [x] 6.6 Update `Config.Save()` to produce clean nested YAML output using the new structs

## 7. Tests

- [x] 7.1 Update `internal/config/config_test.go` for new nested struct format and stdlib loading
- [x] 7.2 Add tests for dynamic defaults: verify empty `notebook.settings.basepath` resolves to `configDir()/notebooks`
- [x] 7.3 Add tests for env var override: verify `WORDFLOW_DICT_LLM_API_KEY` overrides config file value, and overrides dynamic defaults
- [x] 7.4 Add tests for `Config.Validate()`: missing API key, invalid temperature, invalid max_tokens, skipped inactive endpoints, wrong version
- [x] 7.5 Add tests for `config` CLI commands: view, get, set, path, init
- [x] 7.6 Add tests for config template: verify `config init` produces a file with required values and commented-out optional values
- [x] 7.7 Add tests for `--config` global flag pointing to alternative config files
- [x] 7.8 Remove tests for `ConfigMapper` (deleted in 6.5)