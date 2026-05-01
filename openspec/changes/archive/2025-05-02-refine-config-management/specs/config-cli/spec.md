## ADDED Requirements

### Requirement: Config view command
The system SHALL provide a `wordflow config view` command that displays the current configuration in YAML format to stdout.

#### Scenario: View config with default file
- **WHEN** user runs `wordflow config view`
- **THEN** the system SHALL read the configuration from the default path, apply dynamic defaults for empty path fields, apply env var overrides, and print the resulting config as YAML to stdout

#### Scenario: View config with custom config path
- **WHEN** user runs `wordflow config view --config /path/to/custom.yaml`
- **THEN** the system SHALL read the configuration from the specified path and print it as YAML to stdout

#### Scenario: View config with env var overrides
- **WHEN** environment variables like `WORDFLOW_DICT_LLM_API_KEY` are set
- **THEN** the system SHALL display the merged configuration showing values from the config file overridden by environment variables

### Requirement: Config set command
The system SHALL provide a `wordflow config set <key> <value>` command that sets a configuration value and persists it to the config file.

#### Scenario: Set a nested config value
- **WHEN** user runs `wordflow config set dict.llm.api_key sk-abc123`
- **THEN** the system SHALL update `dict.llm.api_key` to `sk-abc123` in the config file and save the file

#### Scenario: Set a top-level config value
- **WHEN** user runs `wordflow config set dict.default llm`
- **THEN** the system SHALL update `dict.default` to `llm` in the config file and save the file

#### Scenario: Set with invalid key
- **WHEN** user runs `wordflow config set invalid.key value`
- **THEN** the system SHALL print an error message listing valid config keys and exit with non-zero status

### Requirement: Config get command
The system SHALL provide a `wordflow config get <key>` command that retrieves a single configuration value.

#### Scenario: Get an existing config value
- **WHEN** user runs `wordflow config get dict.llm.model`
- **THEN** the system SHALL print the current value (e.g., `glm-4`) to stdout

#### Scenario: Get a non-existent key
- **WHEN** user runs `wordflow config get nonexistent.key`
- **THEN** the system SHALL print an error message and exit with non-zero status

#### Scenario: Get value overridden by env var
- **WHEN** user runs `wordflow config get dict.llm.api_key` and `WORDFLOW_DICT_LLM_API_KEY` is set
- **THEN** the system SHALL print the environment variable value, not the file value

#### Scenario: Get value that has a dynamic default
- **WHEN** user runs `wordflow config get notebook.settings.basepath` and the config file has the field empty
- **THEN** the system SHALL print the computed default value (e.g., `/home/user/.config/wordflow/notebooks`)

### Requirement: Config path command
The system SHALL provide a `wordflow config path` command that prints the config file path.

#### Scenario: Show default config path
- **WHEN** user runs `wordflow config path`
- **THEN** the system SHALL print the absolute path to the config file (e.g., `/home/user/.config/wordflow/config.yaml`)

#### Scenario: Show custom config path
- **WHEN** user runs `wordflow config path --config /custom/path.yaml`
- **THEN** the system SHALL print `/custom/path.yaml`

### Requirement: Config init command
The system SHALL provide a `wordflow config init` command that creates a default config file with comments if one does not exist.

#### Scenario: Init config when no config exists
- **WHEN** user runs `wordflow config init` and no config file exists
- **THEN** the system SHALL create a new config file at the default path with all required values filled in and optional/sensitive values commented out with inline descriptions

#### Scenario: Init config when config already exists
- **WHEN** user runs `wordflow config init` and a config file already exists
- **THEN** the system SHALL print a message indicating the config file already exists and its location, and exit without overwriting

### Requirement: Global config flag
The system SHALL support a `--config` global flag on all wordflow commands to specify an alternative configuration file path.

#### Scenario: Use custom config file
- **WHEN** user runs `wordflow --config /path/to/config.yaml dict hello`
- **THEN** the system SHALL load configuration from `/path/to/config.yaml` instead of the default path

#### Scenario: Custom config file does not exist
- **WHEN** user runs `wordflow --config /nonexistent/config.yaml dict hello`
- **THEN** the system SHALL create the config file with defaults (including comments) at the specified path and proceed