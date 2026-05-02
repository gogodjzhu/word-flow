## Why

When any command modifies config (e.g. `wordflow dict -d youdao hello`, `wordflow config set dict.llm.api_key xxx`), `Config.Save()` uses `yaml.Marshal` on the entire in-memory struct to overwrite the file. This destroys all comments, injects resolved default values (paths, zero-valued numeric fields) as hard-coded entries, and can leak environment-overridden secrets (api_key) into the file. Only the target field should change; everything else must be preserved as-is.

## What Changes

- Add `PatchYAMLFile(filename, key, value string) error` to `internal/config/` that reads the raw YAML file, parses it as a `yaml.Node` tree, locates or creates the leaf node at the given dot-path key, sets its value, and writes the tree back — preserving all comments, formatting, and untouched fields
- Replace all command-line `cfg.Save()` calls with `PatchYAMLFile`, so that only the modified key is written:
  - `pkg/cmd/config/config.go`: `setConfigValue` + `Save()` → `PatchYAMLFile`
  - `pkg/cmd/dict/dict.go`: `cfg.Dict.Default = ...; cfg.Save()` → `PatchYAMLFile`
  - `pkg/cmd/trans/trans.go`: `cfg.Trans.Default = ...; cfg.Save()` → `PatchYAMLFile`
  - `pkg/cmd/dict/notebook.go`: `cfg.Notebook.Default = ...; cfg.Save()` → `PatchYAMLFile`
- Keep `Config.Save()` for internal/test use but remove it from CLI command paths
- Ensure `PatchYAMLFile` handles missing intermediate nodes by creating them (e.g. `trans.llm` section missing when setting `trans.llm.api_key`)

## Capabilities

### New Capabilities
- `yaml-patch`: Incremental YAML file patching — modify a single key-path value without rewriting the rest of the file (preserving comments, formatting, and non-target fields)

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- **New file**: `internal/config/patch.go` (~80 lines)
- **Modified files**: `pkg/cmd/config/config.go`, `pkg/cmd/dict/dict.go`, `pkg/cmd/trans/trans.go`, `pkg/cmd/dict/notebook.go` — swap `Save()` calls for `PatchYAMLFile`
- **No breaking changes** to public API, CLI flags, or config format
- **Behavioral change**: `config set` and flag-based config mutations now preserve file comments and non-target fields instead of rewriting the entire file