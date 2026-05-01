## ADDED Requirements

### Requirement: Startup configuration validation
The system SHALL validate the configuration on startup before executing any command, and SHALL provide clear, actionable error messages for invalid or missing configuration.

#### Scenario: Missing LLM API key when LLM is default dictionary
- **WHEN** user runs `wordflow dict hello` with `dict.default` set to `llm` and no `dict.llm.api_key` configured (neither in file nor via env var)
- **THEN** the system SHALL print an error message including guidance on how to set the key (e.g., `LLM API key is required when using LLM dictionary. Set it via: wordflow config set dict.llm.api_key <key> or WORDFLOW_DICT_LLM_API_KEY=<key>`) and exit with non-zero status

#### Scenario: Missing LLM API key when LLM is not default
- **WHEN** user runs `wordflow dict hello` with `dict.default` set to `youdao` and `dict.llm.api_key` is empty
- **THEN** the system SHALL NOT produce a validation error; empty LLM config is acceptable when LLM is not the active endpoint

#### Scenario: Invalid temperature value
- **WHEN** user runs `wordflow dict hello` with `dict.default` set to `llm` and `dict.llm.temperature` set to `5.0`
- **THEN** the system SHALL print an error message indicating the valid range (0-2) and exit with non-zero status

#### Scenario: Invalid max_tokens value
- **WHEN** `dict.llm.max_tokens` is set to `0` or a negative number with `dict.default` set to `llm`
- **THEN** the system SHALL print an error message indicating `max_tokens` must be positive and exit with non-zero status

### Requirement: Validate only active endpoints
The system SHALL only validate configuration for endpoints that are actively in use (the `dict.default` endpoint and the `notebook.default` notebook settings).

#### Scenario: Invalid config for inactive endpoint
- **WHEN** `youdao` is the default dictionary and `llm.api_key` is empty
- **THEN** the system SHALL NOT produce a validation error; inactive endpoint configs are not validated

#### Scenario: Validate trans command LLM config
- **WHEN** user runs `wordflow trans hello` and `dict.llm.api_key` is not set
- **THEN** the system SHALL print an error message indicating the LLM API key is required for translation and exit with non-zero status

### Requirement: Config version validation
The system SHALL validate that the config file version is `"v1"` on load. If the version is missing or does not match, the system SHALL print an error message suggesting the user run `wordflow config init` to regenerate the config.

#### Scenario: Valid config version
- **WHEN** user loads a config file with `version: v1`
- **THEN** the system SHALL proceed normally

#### Scenario: Missing or invalid config version
- **WHEN** user loads a config file with `version: 0.1` or no version field
- **THEN** the system SHALL print an error message (e.g., `Unsupported config version. Run 'wordflow config init' to regenerate your config`) and exit