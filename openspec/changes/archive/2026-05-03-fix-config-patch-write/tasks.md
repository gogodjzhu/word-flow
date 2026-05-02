## 1. Core: PatchYAMLFile function

- [x] 1.1 Create `internal/config/patch.go` with `PatchYAMLFile(filename, key, value string) error` — reads file, parses into `yaml.Node` tree, walks/creates nodes along dot-path, sets leaf value
- [x] 1.2 Implement `findOrCreateNode(root *yaml.Node, parts []string) *yaml.Node` — traverses mapping nodes, creates intermediate mapping nodes when missing, returns the target leaf node
- [x] 1.3 Implement value type coercion (string → bool/int/float/Duration) with proper YAML node tags (`!!bool`, `!!int`, `!!float`, plain string for Duration)
- [x] 1.4 Implement `writeNodeToFile(root *yaml.Node, filename string) error` — encodes the node tree back to YAML and writes to file atomically (write to temp, then rename)

## 2. Tests: PatchYAMLFile

- [x] 2.1 Test: set a leaf value on an existing path preserves comments and other fields
- [x] 2.2 Test: set a value on a path with missing intermediate nodes creates the full path
- [x] 2.3 Test: type coercion — string values written with correct YAML tags for bool, int, float, Duration
- [x] 2.4 Test: commented-out sections survive patching; new nodes are created alongside them
- [x] 2.5 Test: empty or invalid YAML file returns a clear error

## 3. Integration: CLI commands use PatchYAMLFile

- [x] 3.1 Modify `pkg/cmd/config/config.go` — `newCmdConfigSet` uses `PatchYAMLFile(cfg.Common.ConfigFilename, key, value)` instead of `setConfigValue` + `cfg.Save()`
- [x] 3.2 Modify `pkg/cmd/dict/dict.go` — replace unconditional `PreRunE` Save with conditional `PatchYAMLFile("dict.default", ...)` only when `-d` flag changes the value
- [x] 3.3 Modify `pkg/cmd/trans/trans.go` — replace unconditional `PreRunE` Save with conditional `PatchYAMLFile("trans.default", ...)` only when `-e` flag changes the value
- [x] 3.4 Modify `pkg/cmd/dict/notebook.go` — replace `cfg.Save()` with `PatchYAMLFile("notebook.default", ...)`, keeping the existing change-detection guard