## ADDED Requirements

### Requirement: Google Translate provider implementation
The system SHALL implement a `TranslatorGoogle` struct in `pkg/translator/google/trans_google.go` that satisfies the `Translator` interface.

#### Scenario: Simple translation
- **WHEN** `Translate("hello world", out, nil)` is called
- **THEN** the translator SHALL call the Google Translate free API with `client=gtx&sl=en&tl=zh-CN&dt=t` and write the translated text to `out`

#### Scenario: Translation with streaming simulation
- **WHEN** `Translate` is called with multi-sentence text, an io.Writer, and `opts.NoStream == false` (or nil opts)
- **THEN** the translator SHALL split text into sentences via `util.SegmentText`, batch segments within the character threshold (sending each batch as one API call), and write each batch's translation to the io.Writer immediately after receiving it

#### Scenario: Translation with NoStream mode
- **WHEN** `Translate` is called with `opts.NoStream == true`
- **THEN** the translator SHALL batch all segments within the threshold, send each batch as one API call, concatenate all batch results, and write the complete translation to `out` in a single write

#### Scenario: Ref mode translation
- **WHEN** `Translate` is called with `opts.Ref == true`
- **THEN** the translator SHALL split text into sentences, batch segments within the threshold, translate each batch, split batch results back into per-segment pairs, and write JSON `[{"raw": "<original segment>", "translation": "<translated segment>"}]` to the io.Writer

#### Scenario: Empty translation response
- **WHEN** the Google API returns an empty result for valid input
- **THEN** the translator SHALL return an error indicating translation failed

#### Scenario: Network error handling
- **WHEN** the Google API request fails due to network issues
- **THEN** the translator SHALL return a wrapped error with context about the failure

### Requirement: Google Translate API URL construction
The `TranslatorGoogle` SHALL construct API URLs using the base `https://translate.googleapis.com/translate_a/single` with parameters: `client=gtx`, `sl=en`, `tl=zh-CN`, `dt=t`, and `q=<URL-encoded text>`.

#### Scenario: URL encoding
- **WHEN** the input text contains spaces or special characters
- **THEN** the text SHALL be properly URL-encoded in the `q` parameter

### Requirement: Google translation response parsing
The `TranslatorGoogle` SHALL parse the JSON array response from the Google Translate API and extract translated text from the first element.

#### Scenario: Parse standard response
- **WHEN** the API returns `[[["你好","hello",null,null,10]],null,"en",...]`
- **THEN** the translator SHALL concatenate all translation segments (the first element of each inner array) into the full translated text

### Requirement: Google segment batching
The `TranslatorGoogle` SHALL batch sentence segments into groups for API calls, rather than sending one request per sentence. Batching reduces network overhead while preserving streaming perception.

**Algorithm**: After splitting text into segments via `util.SegmentText`, iterate through segments accumulating into a batch. Add a segment to the current batch if the combined length does not exceed a character threshold (default: 500 characters). When adding the next segment would exceed the threshold, send the current batch as one API call, then start a new batch. A single segment that exceeds the threshold is sent as its own batch.

**Ref mode batching**: When `opts.Ref == true`, the translator SHALL map batch results back to individual segment pairs. The Google API returns translated text that corresponds to the batch input; the translator SHALL map the original segments to their translated portions within the batch result.

#### Scenario: Short text within threshold
- **WHEN** input text segments total length is under 500 characters
- **THEN** all segments SHALL be sent as a single API request

#### Scenario: Long text exceeding threshold
- **WHEN** input text segments total length exceeds 500 characters
- **THEN** segments SHALL be batched into multiple API requests, each batch under 500 characters, processed sequentially

#### Scenario: Ref mode with batching
- **WHEN** ref mode is enabled and text requires multiple batches
- **THEN** each batch SHALL be mapped back to per-segment `{raw, translation}` pairs, preserving the correspondence between original and translated segments

### Requirement: Google sentence segmentation
The `TranslatorGoogle` SHALL use `util.SegmentText(text)` from `pkg/util` to split input text into sentences for streaming simulation. The segmentation algorithm SHALL:
1. Split text after sentence-ending punctuation (`.`, `!`, `?`) followed by whitespace
2. Re-merge segments where the word before the punctuation is a known abbreviation (e.g., "Mr", "Mrs", "Dr", "U.S", "e.g", "i.e", "vs", "etc", "Inc", "Corp", "Jr", "Sr", "St" — case-insensitive matching)
3. Treat text with no sentence-ending punctuation as a single segment

#### Scenario: Multiple sentences
- **WHEN** input is "Hello world. How are you? I am fine."
- **THEN** `SegmentText` SHALL return `["Hello world.", " How are you?", " I am fine."]`

#### Scenario: Single word or phrase
- **WHEN** input has no sentence-ending punctuation
- **THEN** `SegmentText` SHALL return the entire text as a single segment

#### Scenario: Abbreviation handling
- **WHEN** input is "Mr. Smith went to the U.S. capital. He arrived."
- **THEN** `SegmentText` SHALL return `["Mr. Smith went to the U.S. capital.", " He arrived."]`

#### Scenario: Abbreviation list covers common cases
- **WHEN** text contains abbreviations like "Dr.", "e.g.", "i.e.", "vs.", "etc.", "Inc."
- **THEN** `SegmentText` SHALL NOT split after these abbreviations

### Requirement: Google translator config
The system SHALL define a `TransGoogleConfig` struct in `internal/config/trans_configs.go` with no required fields (empty struct). It SHALL implement `TransEndpointConfig` (i.e., `Validate() error` returning nil).

#### Scenario: Default Google config
- **WHEN** config is loaded without explicit google settings
- **THEN** the system SHALL use a default `TransGoogleConfig` with zero values