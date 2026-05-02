## Context

The wordflow CLI manages a YAML config file (`~/.config/wordflow/config.yaml`). Currently, every write path uses `Config.Save()` which calls `yaml.Marshal` on the entire in-memory struct. This was acceptable when the config was simple, but now causes real problems:

- Comments (usage hints, API key placeholders) are destroyed on every save
- `applyDefaults()` fills zero-value fields with computed paths (`ecdict.db_filename`, `notebook.settings.basepath`), and `Save()` writes those into the file as hard-coded absolute paths
- Environment variable overrides (e.g. `WORDFLOW_DICT_LLM_API_KEY`) leak into the persisted file
- Several commands (`wordflow dict`, `wordflow trans`) call `Save()` in `PreRunE` unconditionally, meaning even a read-only lookup rewrites the entire file

The project already uses `gopkg.in/yaml.v3` (vendored), which provides a `yaml.Node` API that preserves comments and formatting.

## Goals / Non-Goals

**Goals:**
- Only the target key's value is modified in the YAML file; all other content (comments, blank lines, field order, non-target values) is preserved byte-for-byte (aside from the changed value)
- `config set trans.llm.api_key xxx` works even when the `trans.llm` section is entirely commented out or absent in the file
- Flag-driven mutations (`-d`, `-e`, `-n`) also use patch-write
- Remove all unconditional `Save()` calls from CLI command paths

**Non-Goals:**
- Full `yaml.Node` migration for reads — reads continue using struct unmarshal + `applyDefaults`
- Changing the Config struct or any validation logic
- Removing `Config.Save()` entirely (still useful for tests and internal use)

## Decisions

### 1. Patch-write via `yaml.Node` tree manipulation

**Decision**: Add `PatchYAMLFile(filename, key, value string) error` that reads the raw file, parses it into a `yaml.Node` tree, walks/creates nodes along the dot-path, sets the leaf value, and writes back.

**Alternatives considered**:
- *Sed-style regex replacement*: Fragile, can't handle nested keys or missing sections correctly
- *Viper/koanf*: Heavy dependency for this single concern; also doesn't preserve comments
- *Full yaml.Node read/write for everything*: Too large a refactor; reads work fine with struct unmarshal

### 2. Value type coercion in patch

**Decision**: `PatchYAMLFile` accepts all values as strings. The existing `setFieldValue` logic (string → string/int/float/bool/Duration) is reused for type coercion. The YAML node's tag is set appropriately (e.g. `!!bool` for booleans, `!!int` for integers) so that `yaml.Unmarshal` on the patched file produces correct types.

### 3. Missing intermediate nodes

**Decision**: When a dot-path like `trans.llm.api_key` is set but `trans.llm` doesn't exist in the file, `PatchYAMLFile` creates the intermediate mapping nodes. This matches the current `setConfigValue` behavior which allocates nil pointers via `reflect.New`.

### 4. Read path unchanged

**Decision**: Keep `ReadConfig` → `yaml.Unmarshal` → `applyDefaults` → `applyEnvOverrides` unchanged. The struct-based read path works correctly for runtime configuration. Only the write path changes.

## Risks / Trade-offs

- **[YAML formatting drift]** `yaml.Node` encoding may change whitespace/quote style for the modified value. Mitigation: only the target leaf node is touched; unused nodes are not re-encoded. Acceptable since the alternative (full rewrite) is strictly worse.
- **[Edge case: truly empty file]** If the config file is empty or invalid YAML, `PatchYAMLFile` cannot parse it. Mitigation: this is already a broken state; the existing `ReadConfig` would also fail. Patch should return a clear error.
- **[Double write on flag change]** If both `-d` and `-n` flags change values, two separate `PatchYAMLFile` calls write the file twice. Mitigation: acceptable; each call is atomic and cheap. Could batch later if needed.