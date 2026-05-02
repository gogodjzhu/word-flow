## Context

Word-flow is a terminal-based dictionary and translation tool. Currently:

- **Dictionary lookup** (`wordflow dict`) has a clean `Dict` interface with 5 providers (Youdao, Etymonline, ECDICT, Merriam-Webster, LLM), a factory pattern, and per-provider config.
- **Text translation** (`wordflow trans`) has no abstraction layer — it directly creates an `llm.Client` and calls `TranslateWithStream()`. This means only LLM can be used for translation.
- Google Translate's free API (`translate.googleapis.com`) is verified to work for both translation and dictionary lookups without any API key.

The change adds Google Translate as both a dictionary provider and a translation provider, while introducing a `Translator` interface to decouple the `trans` command from the LLM implementation.

## Goals / Non-Goals

**Goals:**
- Introduce a `Translator` interface with unified `Translate(text, io.Writer, opts)` method signature
- Implement Google Translate as a translation provider with streaming simulation via sentence segmentation
- Implement LLM translator with private streaming/non-streaming/ref logic, removing `TranslateWithStream` from `llm.Client`
- Add Google as a dictionary provider using `dt=bd` API data
- Create independent `trans` config section separate from `dict`
- Support `--ref` (side-by-side) mode for both LLM and Google translators

**Non-Goals:**
- Multi-language support (English-to-Chinese only for now)
- Google Cloud Translation API (official paid API) — free API only
- Combining dict and trans config for shared providers (they remain independent)
- Changing the existing `Dict` interface

## Decisions

### D1: Translator interface uses io.Writer, not separate streaming/sync methods

**Decision**: Single `Translate(text string, out io.Writer, opts *TransOptions) error` method.

**Alternatives considered**:
- Separate `Translate()` (sync) and `TranslateStream()` (streaming) methods — rejected because sync translation is just "write everything to Writer at once" vs streaming "write incrementally". The Writer abstraction unifies both.
- Returning `(string, error)` — rejected because it forces buffering the entire response, losing the streaming experience for LLM providers.

**Rationale**: `io.Writer` lets the caller control buffering behavior. A `bytes.Buffer` for sync, `os.Stdout` for streaming, `io.Pipe` for ref-mode rendering. The translator implementation doesn't need to know.

### D2: Google streaming simulation via segment batching with threshold

**Decision**: Google translator segments input text by sentences via `util.SegmentText`, then batches segments into groups that fit within a character threshold (hardcoded). Each batch is sent as a single Google Translate API call. Segments within a batch are concatenated with spaces, translated together, and results are written to `io.Writer` per-batch.

**Rationale**: Sending one API request per sentence is wasteful and increases latency. Batching multiple sentences into a single request reduces network overhead while still providing incremental output. A character threshold (e.g., 500 chars) keeps batches small enough for streaming perception while reducing total requests. The threshold is hardcoded for simplicity; it can be exposed to config later if needed.

**Batching algorithm**:
1. Call `util.SegmentText(text)` to split into sentences
2. Iterate through segments, accumulating into the current batch until adding the next segment would exceed the threshold
3. Send each batch as a single API call, write the batch result to `out` immediately
4. For ref mode, map each batch's original segments to their translated results

### D3: Google ref mode outputs JSON segment pairs matching LLM format

**Decision**: When `opts.Ref == true`, Google translator outputs JSON array `[{"raw": "...", "translation": "..."}]` matching the format LLM produces, so the existing `TranslationRenderer` works unchanged.

**Rationale**: The `--ref` rendering pipeline (`TranslationRenderer`, `SimpleStreamReader`) already expects this format. Reusing it avoids duplicating rendering logic.

### D4: Independent trans config section

**Decision**: New top-level `trans` section in config YAML, with its own `default`, `llm`, and `google` sub-configs. Independent from `dict.llm`.

**Alternatives considered**:
- Share `dict.llm` config — rejected because users may want different LLM settings for dictionary vs translation (e.g., a cheaper model for dict, a better model for trans).
- Elevate `llm` to top-level — rejected because it's a bigger refactor and changes the mental model.

**Rationale**: Independent configs are clearer. A user might use `dict.default: youdao` + `trans.default: google` as their everyday setup, with `trans.llm` configured for when they need higher quality.

### D5: TransOptions carries cross-cutting rendering options

**Decision**: `TransOptions` struct contains `Ref bool` and `NoStream bool` — both are cross-cutting rendering concerns that affect all providers, not provider-specific options.

**Alternatives considered**:
- Only `Ref` in TransOptions, provider-specific settings in config — rejected because `NoStream` is not provider-specific; it is a rendering mode determined by the command's `--no-stream` flag that all translators must respect (LLM switches API mode, Google writes segments all-at-once vs incrementally).
- Separate `Translate()` (sync) and `TranslateStream()` (streaming) methods — already rejected in D1.

**Rationale**: `NoStream` is a cross-cutting concern: the `--no-stream` command flag controls whether output should be delivered incrementally or all-at-once. It maps to `stream=false` in the LLM API request, and to batch-vs-incremental writing in the Google adapter. Unlike temperature/max_tokens (which are truly provider-specific), `NoStream` is set by the user at invocation time and affects provider behavior uniformly.

### D6: Google Dict implementation maps dt=bd response to WordItem

**Decision**: Parse the `dt=bd` response array (second element containing word class definitions) into `WordMeaning` structs. No phonetic data (Google free API doesn't provide it). Source field set to `"google"`.

**Rationale**: The `dt=bd` response provides part-of-speech tags, translations, and synonyms — enough for a useful dictionary entry. Missing phonetics is acceptable (different providers have different capabilities, as established in the Dict interface contract).

### D7: LLM translator integrates into Translate interface, removes TranslateWithStream

**Decision**: The `pkg/translator/llm/trans_llm.go` module does NOT wrap `llm.Client.TranslateWithStream()`. Instead, it directly incorporates the LLM translation logic (prompt building, streaming, non-streaming, response parsing) as private methods. The public `llm.Client` no longer exposes `TranslateWithStream`; `Translate(text, out, opts)` on the `Translator` interface is the sole public API for translation.

**Mode mapping via `TransOptions`**:
- `opts == nil || (!Ref && !NoStream)` → streaming plain translation: stream SSE to `out`
- `opts.NoStream == true && !Ref` → non-streaming: full response to `out` in one write
- `opts.Ref == true && !NoStream` → streaming ref: stream SSE via `io.Pipe`, produce `[{"raw":"...","translation":"..."}]`
- `opts.Ref == true && NoStream` → non-streaming ref: full response as JSON segment pairs to `out`

**Rationale**: This project should have a single clean translation API surface. The old `TranslateWithStream` had an awkward signature `(text, disableStream, out, ref) → (string, error)` that mixed concerns (streaming control, ref mode, return value vs writer). The new `Translator.Translate(text, out, opts) → error` interface is cleaner and more extensible. Keeping the old method creates two code paths for the same functionality. The LLM adapter should own its streaming/ref/non-streaming logic privately, not delegate to a public method on `llm.Client`. This means `llm.Client` is reduced to only `TranslateAndExplain()` for the dict command; all translation-specific code moves to `pkg/translator/llm/`.

**Implementation note**: `pkg/translator/llm/trans_llm.go` can reference `internal/llm` for HTTP helpers, types (`ChatRequest`, `StreamResponse`, etc.), and `PromptBuilder`, but the `translate` logic itself is implemented within the translator package. Alternatively, `internal/llm` can expose a lower-level `SendChatRequest` that the translator calls directly.

### D8: --endpoint flag overrides config default

**Decision**: The `wordflow trans` command accepts an `--endpoint` / `-e` flag that overrides `trans.default` for the current invocation, consistent with the `--dictionary` / `-d` flag pattern in the `dict` command.

**Rationale**: Users need a quick way to switch translation endpoints without editing config (e.g., default to Google for speed, but `--endpoint=llm` for quality on demand). Using `--endpoint` matches the `--dictionary` flag pattern and the `Endpoint` type already used in `pkg/dict`, keeping the vocabulary consistent across dict and trans commands.

### D9: Sentence segmentation with abbreviation awareness

**Decision**: Abstract sentence segmentation into a reusable `pkg/util.SegmentText(text string) []string` function that uses a two-step algorithm: (1) split on sentence-ending punctuation (`.!?`) followed by whitespace, then (2) re-merge segments where the word before the punctuation is a known abbreviation (e.g., "Mr", "Dr", "U.S", "e.g", "i.e", "vs", "etc") or a single capital letter. The abbreviation list covers common English abbreviations.

**Rationale**: A naive regex split on `.!?` breaks on abbreviations like "U.S.", "Mr. Smith went home." Creating a reusable utility enables both the Google translator and potential future consumers (e.g., text preprocessing, other segmenters) to share correctly-implemented sentence splitting.

### D10: Default translator behavior change from LLM to Google

**Decision**: The new `trans.default: "google"` default changes the behavior for existing users who previously got LLM translation. This is acceptable because: (1) Google works without configuration while LLM requires an API key, so new and unconfigured users get a working default; (2) existing LLM users can set `trans.default: "llm"` or use `--endpoint=llm` to restore previous behavior; (3) the `wordflow config init` template documents both options clearly.

**Rationale**: Zero-configuration defaults provide the best first-run experience. Users who have already configured LLM keys likely have a config file and can change the default.

## Risks / Trade-offs

- **[Google free API instability]** → The free `translate.googleapis.com` endpoint is undocumented and may break or get rate-limited. Mitigation: clear documentation that this is a free/unofficial endpoint; the `GoogleConfig` is empty now but can be extended with an official API key field later.
- **[Sentence segmentation quality]** → Splitting on sentence-ending punctuation can fail on abbreviations (e.g., "U.S.", "Mr. Smith"). Mitigation: `SegmentText` uses a two-step algorithm that re-merges segments after known abbreviations (see D9). Remaining edge cases (uncommon abbreviations) are acceptable for now.
- **[Config duplication]** → Users configuring both `dict.llm` and `trans.llm` must enter API key twice. Mitigation: environment variables can set both; future enhancement could add config references.
- **[Streaming latency for Google]** → Sequential batch-by-batch API calls add latency, but less than per-sentence calls. With a 500-char batch threshold, a typical paragraph (3-5 sentences) becomes 1-2 requests. Acceptable for a free provider.