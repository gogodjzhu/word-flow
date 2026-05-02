## 1. Config Foundation

- [x] 1.1 Create `internal/config/trans_configs.go` with `TransEndpointConfig` interface, `TransLLMConfig` (mirroring `LLMConfig` fields with Validate), and `TransGoogleConfig` (empty struct with nil Validate)
- [x] 1.2 Add `TransConfig` struct to `internal/config/config.go` with `Default`, `LLM`, and `Google` fields; add `Trans *TransConfig` to root `Config` struct
- [x] 1.3 Update `applyDefaults()` in config.go to set trans defaults (`default: "google"`, initialize nil sub-configs)
- [x] 1.4 Update `ConfigTemplate()` to include the `trans` section with `default: google`, `google: {}`, and commented-out llm settings
- [x] 1.5 Add `GetEndpointConfig()` method to `TransConfig` mapping endpoint names to config structs
- [x] 1.6 Update `buildKnownPaths()` and env override system to support `WORDFLOW_TRANS_*` prefix
- [x] 1.7 Add `Validate()` method to `Config` that validates the active trans endpoint config via `Trans.GetEndpointConfig()`; update `ValidateForTrans()` to validate `cfg.Trans` (NOT `cfg.Dict.LLM`)

## 2. Translator Interface & Factory

- [x] 2.1 Create `pkg/translator/translator.go` with `TransOptions` struct (`Ref bool`, `NoStream bool`), `Translator` interface (`Translate(text string, out io.Writer, opts *TransOptions) error`), `TranslatorInfo` struct, `AvailableTranslators()` function, and `NewTranslator()` factory function
- [x] 2.2 Add endpoint constants (`TransLLM`, `TransGoogle`) to `pkg/translator/translator.go`

## 3. LLM Translator Implementation

- [x] 3.1 Create `pkg/translator/llm/trans_llm.go` implementing `Translator` interface with private methods for streaming and non-streaming LLM calls. Move prompt building and streaming/response processing logic from `internal/llm` into this package (or reference shared types)
- [x] 3.2 Implement `Translate()`: dispatch based on `opts.Ref` and `opts.NoStream` to private methods:
  - `!Ref && !NoStream` → call private streaming method (writes incrementally to `out`)
  - `!Ref && NoStream` → call private non-streaming method (writes complete result to `out` in one write)
  - `Ref && !NoStream` → call private streaming-ref method (stream SSE via `io.Pipe`, write JSON segments)
  - `Ref && NoStream` → call private non-streaming-ref method (full JSON response to `out`)
- [x] 3.3 Remove `TranslateWithStream` method from `llm.Client` (keep `TranslateAndExplain` for dict)
- [x] 3.4 Update `pkg/cmd/trans/trans.go` to remove all direct references to `llm.Client.TranslateWithStream` and `llm.NewClient`; use `Translator` interface instead

## 4. Sentence Segmentation Utility

- [x] 4.1 Create `pkg/util/segment.go` with `SegmentText(text string) []string` function
- [x] 4.2 Implement two-step algorithm: (1) split after `.!?` followed by whitespace, (2) re-merge segments where the word before the punctuation is a known abbreviation
- [x] 4.3 Define abbreviation list: Mr, Mrs, Ms, Dr, Prof, Rev, Gen, Rep, Sen, St, Jr, Sr, Inc, Corp, Ltd, Co, vs, etc, e.g, i.e, a.m, p.m, U.S, U.K, approx, dept, est, gov, misc, tech, vol, no (case-insensitive)
- [x] 4.4 Return the entire text as a single segment when no sentence-ending punctuation is found
- [x] 4.5 Add unit tests for `SegmentText`: multiple sentences, single segment, abbreviations ("Mr. Smith went to the U.S. capital.", "e.g. this is a test."), trailing punctuation

## 5. Segment Batching Utility

- [x] 5.1 Create `pkg/util/batch.go` with `BatchSegments(segments []string, threshold int) [][]string` function that groups segments into batches where each batch's total character length does not exceed the threshold
- [x] 5.2 A single segment exceeding the threshold is placed in its own batch
- [x] 5.3 Add unit tests for `BatchSegments`: single batch, multiple batches, segment exceeding threshold, empty input

## 6. Google Translator Implementation

- [x] 6.1 Create `pkg/translator/google/trans_google.go` with `TranslatorGoogle` struct and constructor accepting `TransGoogleConfig`
- [x] 6.2 Use `util.SegmentText()` for sentence splitting, then `util.BatchSegments()` for grouping segments into API-call-sized batches (threshold: 500 chars)
- [x] 6.3 Implement Google Translate API calling logic: URL construction (`client=gtx&sl=en&tl=zh-CN&dt=t`), HTTP request via `internal/util`, and response parsing (extract translation segments from JSON array)
- [x] 6.4 Implement `Translate()` for plain mode: batch segments, send each batch as one API call; when `!NoStream`, write each batch result to `out` immediately after receiving; when `NoStream`, collect all batch results and write complete translation
- [x] 6.5 Implement ref mode (`Ref == true`): batch segments, send each batch as one API call, map batch results back to per-segment `{raw, translation}` pairs, write JSON array `[{"raw": "...", "translation": "..."}]`; in streaming mode write pairs progressively; in NoStream mode write complete JSON
- [x] 6.6 Handle edge cases: empty input, network errors, empty API response

## 7. Google Dict Implementation

- [x] 7.1 Create `pkg/dict/google/dict_google.go` implementing `Dict` interface with `Search(word)` method
- [x] 7.2 Implement Google Dict API call: URL construction with `client=gtx&sl=en&tl=zh-CN&dt=t&dt=bd` parameter
- [x] 7.3 Implement response parsing: extract `dt=bd` dictionary data (part of speech, translations), map to `WordMeaning` structs; fall back to plain translation when bd data is missing
- [x] 7.4 Add `Google` endpoint constant to `pkg/dict/dict.go`, add `DictGoogle` case to `NewDict()` factory, and add Google entry to `AvailableDictionaries()`
- [x] 7.5 Add `GoogleConfig` struct (empty) to `internal/config/dict_configs.go` with nil `Validate()`, add `Google *GoogleConfig` field to `DictConfig`, update `GetEndpointConfig()` switch, update `applyDefaults()`

## 8. Trans Command Rewrite

- [x] 8.1 Rewrite `pkg/cmd/trans/trans.go` to use `Translator` interface: load config, create translator via `NewTranslator()`
- [x] 8.2 Add `--endpoint` / `-e` flag: override `trans.default` config for the current invocation; pass overridden endpoint name to `NewTranslator()`
- [x] 8.3 Handle `--no-stream` mode: construct `TransOptions{NoStream: true}`, call `translator.Translate(text, &buf, opts)` then render buffered result
- [x] 8.4 Handle default streaming mode: construct `TransOptions{NoStream: false}` (or nil), call `translator.Translate(text, f.IOStreams.Out, opts)` to write directly to stdout
- [x] 8.5 Handle `--ref` + streaming mode: use `io.Pipe` with `SimpleStreamReader` and `TranslationRenderer`, construct `TransOptions{Ref: true, NoStream: false}`
- [x] 8.6 Handle `--ref` + `--no-stream` mode: buffer result with `TransOptions{Ref: true, NoStream: true}` then render with `TranslationRenderer`
- [x] 8.7 Update validation to use `ValidateForTrans()` which validates `cfg.Trans` (NOT `cfg.Dict.LLM`)
- [x] 8.8 Remove direct `llm.NewClient` usage and `llm` package import from `trans.go`

## 9. Tests & Verification

- [x] 9.1 Add unit tests for `TransConfig` validation and defaults
- [x] 9.2 Add unit tests for `util.SegmentText` with multiple sentences, single segment, abbreviation edge cases
- [x] 9.3 Add unit tests for `util.BatchSegments` with various segment lengths and thresholds
- [x] 9.4 Add unit tests for Google Translate API response parsing
- [x] 9.5 Add unit tests for Google Dict response parsing to `WordItem`
- [x] 9.6 Add unit tests for `NewTranslator` factory with valid/invalid configs, `--endpoint` override
- [ ] 9.7 Add unit tests for LLM translator `Translate()` mode mapping (streaming/non-streaming/ref)
- [ ] 9.8 Manually test `wordflow trans "hello" --endpoint=google` end-to-end
- [ ] 9.9 Manually test `wordflow trans "hello" --endpoint=llm` end-to-end
- [ ] 9.10 Manually test `wordflow dict hello -d google` end-to-end
- [x] 9.11 Run existing tests to verify no regressions