## ADDED Requirements

### Requirement: Environment variable override
The system SHALL support overriding any configuration value via environment variables using the `WORDFLOW_` prefix and underscore-delimited nested key paths.

#### Scenario: Override LLM API key via env var
- **WHEN** the environment variable `WORDFLOW_DICT_LLM_API_KEY` is set to `sk-env-key`
- **THEN** the system SHALL use `sk-env-key` as the LLM API key, regardless of what is in the config file

#### Scenario: Override dict default via env var
- **WHEN** the environment variable `WORDFLOW_DICT_DEFAULT` is set to `llm`
- **THEN** the system SHALL use `llm` as the default dictionary, overriding the config file value

#### Scenario: Env var with nested key path
- **WHEN** the environment variable `WORDFLOW_DICT_LLM_TIMEOUT` is set to `60s`
- **THEN** the system SHALL interpret this as `dict.llm.timeout` = `60s` with proper type coercion to `time.Duration`

### Requirement: Environment variable precedence
Environment variables SHALL take precedence over values in the config file. When both are set, the environment variable value SHALL be used. Dynamic defaults (computed at load time for empty path fields) have the lowest precedence.

#### Scenario: Both file and env var set
- **WHEN** config file sets `dict.llm.api_key` to `file-key` and `WORDFLOW_DICT_LLM_API_KEY` is set to `env-key`
- **THEN** the system SHALL use `env-key` as the API key value

#### Scenario: Env var overrides dynamic default
- **WHEN** `notebook.settings.basepath` is empty in the config file and `WORDFLOW_NOTEBOOK_SETTINGS_BASEPATH` is set to `/custom/path`
- **THEN** the system SHALL use `/custom/path` instead of the computed default

### Requirement: Env var key mapping
The system SHALL map environment variables to config keys by: removing the `WORDFLOW_` prefix, converting to lowercase, and replacing underscores with dots for nested path traversal. Multi-word leaf field names use underscores in the env var (e.g., `api_key`, `max_tokens`).

#### Scenario: Standard key mapping
- **WHEN** `WORDFLOW_DICT_LLM_API_KEY` environment variable is set
- **THEN** the system SHALL map it to config key `dict.llm.api_key`

#### Scenario: Mapped value appears in config view
- **WHEN** `WORDFLOW_DICT_LLM_API_KEY` is set and user runs `wordflow config view`
- **THEN** the output SHALL show `dict.llm.api_key` with the environment variable's value