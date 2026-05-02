## ADDED Requirements

### Requirement: PatchYAMLFile modifies only the target key
`PatchYAMLFile(filename, key, value string) error` SHALL read the raw YAML file, parse it as a `yaml.Node` tree, locate (or create) the leaf node at the given dot-path key, set its value, and write the tree back. All comments, blank lines, field ordering, and non-target field values SHALL be preserved.

#### Scenario: Set a leaf value on an existing path
- **WHEN** the config file contains `dict.llm.timeout: 30s` and `PatchYAMLFile(filename, "dict.llm.timeout", "60s")` is called
- **THEN** only the value for `dict.llm.timeout` changes to `60s`; all other content (comments, other fields) is preserved

#### Scenario: Set a value on a path with missing intermediate nodes
- **WHEN** the config file has no `trans.llm` section and `PatchYAMLFile(filename, "trans.llm.api_key", "sk-xxx")` is called
- **THEN** the `trans.llm` mapping node and `api_key` leaf node SHALL be created, and `api_key` SHALL be set to `sk-xxx`

#### Scenario: Set a value with type coercion
- **WHEN** `PatchYAMLFile` is called with `"dict.llm.timeout"` and value `"60s"` (a Duration)
- **THEN** the YAML node SHALL be written with appropriate tag (`!!str` or unquoted) so that `yaml.Unmarshal` produces the correct typed value

#### Scenario: Patch a boolean value
- **WHEN** `PatchYAMLFile` is called with key `"some.flag"` and value `"true"`
- **THEN** the node SHALL be written with `!!bool` tag

#### Scenario: Patch an integer value
- **WHEN** `PatchYAMLFile` is called with key `"notebook.settings.max_reviews_per_session"` and value `"100"`
- **THEN** the node SHALL be written with `!!int` tag

### Requirement: CLI config set command uses patch-write
The `wordflow config set <key> <value>` command SHALL use `PatchYAMLFile` instead of `setConfigValue` + `Config.Save()`. Only the target key SHALL be modified in the file.

#### Scenario: Set api_key preserves other content
- **WHEN** user runs `wordflow config set dict.llm.api_key sk-xxx` on a config file with comments
- **THEN** only `dict.llm.api_key` is changed; comments, other fields, and formatting are unchanged

### Requirement: Flag-based config mutations use patch-write
Commands that modify config via flags (`-d`, `-e`, `-n`) SHALL use `PatchYAMLFile` instead of `Config.Save()`.

#### Scenario: dict command with -d flag
- **WHEN** user runs `wordflow dict -d ecdict hello`
- **THEN** only `dict.default` is changed to `"ecdict"` in the config file; all other content is preserved

#### Scenario: trans command with -e flag
- **WHEN** user runs `wordflow trans -e llm hello`
- **THEN** only `trans.default` is changed to `"llm"` in the config file

#### Scenario: notebook command with -n flag
- **WHEN** user runs `wordflow notebook -n mybook review` and the notebook differs from current default
- **THEN** only `notebook.default` is changed

#### Scenario: No flag change means no file write
- **WHEN** user runs `wordflow dict hello` without `-d` flag
- **THEN** the config file is not written at all

### Requirement: Original config file formatting is preserved
When `PatchYAMLFile` writes back the file, the YAML content outside of the modified node SHALL be identical to the original file content, including comments, blank lines, key ordering, and indentation style.

#### Scenario: Comments survive a patch
- **WHEN** the config file contains `# Required. Set via WORDFLOW_DICT_LLM_API_KEY` above `dict.llm.api_key`
- **AND** `PatchYAMLFile` modifies `dict.llm.timeout`
- **THEN** the comment above `api_key` is still present in the output

#### Scenario: Commented-out sections survive a patch
- **WHEN** the config file contains `# llm:` (commented out) under `trans:` and `PatchYAMLFile` creates `trans.llm.api_key`
- **THEN** the commented-out `# llm:` line is preserved; a new `llm:` mapping with `api_key` is added