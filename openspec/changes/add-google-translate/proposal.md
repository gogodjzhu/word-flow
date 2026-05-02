## Why

The `wordflow trans` command currently only supports LLM-based translation, with no abstraction layer for alternative providers. Adding Google Translate (free API) as a translation provider gives users a zero-configuration translation option that doesn't require an LLM API key. Additionally, adding Google as a dictionary provider (using the dictionary data from the same API) enriches the `wordflow dict` command with a free, no-key-required dictionary source.

## What Changes

- Add a `Translator` interface as a top-level abstraction for text translation, mirroring the existing `Dict` interface for dictionary lookup
- Implement `pkg/translator/google` — Google Translate provider using the free `translate.googleapis.com` API, with segment-batched streaming simulation (batch threshold) and ref (side-by-side) mode support
- Implement `pkg/translator/llm` — LLM-based translator with private streaming/non-streaming/ref logic, conforming to the `Translator` interface; removes the public `TranslateWithStream` method from `llm.Client`
- Add `pkg/dict/google` — Google dictionary provider using `dt=bd` parameter to return structured word definitions mapped to `WordItem`
- Add independent `trans` top-level config section (`TransConfig`) with `default`, `llm`, and `google` fields, separate from `dict` config
- Add `GoogleConfig` to `DictConfig` for the new dictionary provider
- Add `--endpoint` / `-e` flag to `wordflow trans` command for overriding `trans.default` per invocation (consistent with `--dictionary` / `-d` in `dict` command)
- Rewrite `pkg/cmd/trans/trans.go` to use the `Translator` interface instead of directly creating `llm.Client`
- Update config template, defaults, and validation to support new `trans` section and `google` dict endpoint

## Capabilities

### New Capabilities
- `translator-interface`: Translator abstraction layer with unified `Translate(text, io.Writer, opts)` interface, factory pattern, streaming/ref/NoStream support
- `google-translator`: Google Translate implementation using free API with segment-batched streaming simulation (batch within char threshold) and ref mode
- `google-dict`: Google dictionary implementation using `dt=bd` API, mapping word definitions to `WordItem`

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- **New packages**: `pkg/translator/`, `pkg/translator/google/`, `pkg/translator/llm/`, `pkg/dict/google/`, `pkg/util/` (segment + batch utilities)
- **Modified packages**: `internal/config/` (add `TransConfig`, `TransLLMConfig`, `TransGoogleConfig`, `GoogleConfig`; update `Config`, defaults, template, validation), `internal/llm/` (remove `TranslateWithStream` and related streaming/ref code; keep `TranslateAndExplain` for dict), `pkg/dict/` (add Google endpoint), `pkg/cmd/trans/` (rewrite to use Translator interface, add `--endpoint` flag)
- **Config breaking change**: New `trans` section in config YAML. Existing configs still work (defaults applied), but users should regenerate config for the new section to appear
- **Default translator behavior change**: The new `trans.default: "google"` default changes translation behavior for existing users who previously got LLM translation. Existing LLM users should set `trans.default: "llm"` or use `--endpoint=llm` to restore previous behavior
- **Dependencies**: No new external dependencies; Google Translate uses the same `internal/util` HTTP helpers as existing providers