## ADDED Requirements

### Requirement: Translator interface definition
The system SHALL define a `Translator` interface in `pkg/translator/translator.go` with the method signature `Translate(text string, out io.Writer, opts *TransOptions) error`.

#### Scenario: Successful translation with default options
- **WHEN** `Translate` is called with text, an io.Writer, and nil or default TransOptions
- **THEN** the translator SHALL write the translated text to the io.Writer and return nil error

#### Scenario: Translation with ref mode enabled
- **WHEN** `Translate` is called with `opts.Ref == true`
- **THEN** the translator SHALL write JSON segment pairs `[{"raw": "...", "translation": "..."}]` to the io.Writer

#### Scenario: Empty input text
- **WHEN** `Translate` is called with empty or whitespace-only text
- **THEN** the translator SHALL return an error

### Requirement: TransOptions struct
The system SHALL define a `TransOptions` struct in `pkg/translator/translator.go` containing:
- `Ref bool` — when true, translators output side-by-side (reference) JSON segment pairs
- `NoStream bool` — when true, translators that support streaming SHALL produce the complete output at once (non-streaming mode), rather than delivering output incrementally

#### Scenario: Ref mode disabled (default)
- **WHEN** TransOptions is nil or Ref is false
- **THEN** translators SHALL output plain translated text

#### Scenario: Ref mode enabled
- **WHEN** TransOptions.Ref is true
- **THEN** translators SHALL output JSON segment pairs format `[{"raw": "...", "translation": "..."}]`

#### Scenario: NoStream mode
- **WHEN** TransOptions.NoStream is true
- **THEN** streaming-capable translators (LLM) SHALL use their non-streaming API endpoint and write the complete result to `out` at once; synchronous translators (Google) SHALL collect all segments and write them together to `out`

#### Scenario: Streaming mode (default)
- **WHEN** TransOptions is nil or NoStream is false
- **THEN** streaming-capable translators SHALL deliver output incrementally as it becomes available; synchronous translators SHALL write each segment to `out` as it completes

### Requirement: Translator factory function
The system SHALL provide a `NewTranslator(conf *config.TransConfig) (Translator, error)` function that creates the appropriate translator based on `conf.Default`.

#### Scenario: Create Google translator
- **WHEN** `conf.Default` is `"google"`
- **THEN** the factory SHALL return a `*TranslatorGoogle` instance

#### Scenario: Create LLM translator
- **WHEN** `conf.Default` is `"llm"`
- **THEN** the factory SHALL return a `*TranslatorLLM` instance

#### Scenario: Unknown translator endpoint
- **WHEN** `conf.Default` is not a recognized translator endpoint name
- **THEN** the factory SHALL return an `InvalidEndpoint` error

#### Scenario: Validate translator config
- **WHEN** the selected translator's config is invalid (e.g., LLM missing api_key)
- **THEN** the factory SHALL return a validation error before constructing the translator

### Requirement: AvailableTranslators listing
The system SHALL provide an `AvailableTranslators() []TranslatorInfo` function returning name and description for each translator.

#### Scenario: List available translators
- **WHEN** `AvailableTranslators()` is called
- **THEN** it SHALL return entries for both `"llm"` and `"google"` with descriptions

### Requirement: TransConfig in root config
The system SHALL add a `TransConfig` struct as a top-level `trans` field in the root `Config` struct.

#### Scenario: Default trans config
- **WHEN** config is loaded without a `trans` section
- **THEN** the system SHALL apply defaults: `default: "google"` and an empty `GoogleConfig`

#### Scenario: Trans config validation
- **WHEN** `trans.default` is `"llm"`
- **THEN** the system SHALL validate that `trans.llm` (NOT `dict.llm`) has required fields (api_key, url, model)

#### Scenario: Trans config env overrides
- **WHEN** environment variables with `WORDFLOW_TRANS_` prefix are set
- **THEN** they SHALL override corresponding YAML values in the `trans` section

#### Scenario: Trans config template
- **WHEN** a new config is initialized
- **THEN** the template SHALL include a `trans` section with `default: google`, an empty `google: {}`, and commented-out `llm` settings

### Requirement: trans command --endpoint flag
The `wordflow trans` command SHALL accept an `--endpoint` / `-e` flag that overrides the `trans.default` config value for the current invocation, consistent with the `--dictionary` / `-d` flag in the `dict` command.

#### Scenario: Endpoint flag overrides default
- **WHEN** `wordflow trans --endpoint=llm "text"` is run and `trans.default` is `"google"`
- **THEN** the command SHALL use the LLM translator instead of Google

#### Scenario: Endpoint flag with unknown value
- **WHEN** `wordflow trans --endpoint=unknown "text"` is run
- **THEN** the command SHALL return an `InvalidEndpoint` error

#### Scenario: No endpoint flag
- **WHEN** `wordflow trans "text"` is run without `--endpoint`
- **THEN** the command SHALL use the translator specified by `trans.default` config

### Requirement: trans command uses Translator interface
The `wordflow trans` command SHALL use the `Translator` interface instead of directly creating `llm.Client`.

#### Scenario: Stdout streaming (default)
- **WHEN** `wordflow trans "text"` is run with a streaming-capable translator (LLM)
- **THEN** the command SHALL call `translator.Translate(text, os.Stdout, nil)` and output SHALL stream to terminal

#### Scenario: No-stream mode
- **WHEN** `wordflow trans --no-stream "text"` is run
- **THEN** the command SHALL call `translator.Translate(text, &buf, &TransOptions{NoStream: true})` then render the buffered result

#### Scenario: Ref mode streaming
- **WHEN** `wordflow trans --ref "text"` is run
- **THEN** the command SHALL call `translator.Translate(text, pw, &TransOptions{Ref: true})` via io.Pipe, and render with `SimpleStreamReader` and `TranslationRenderer`

#### Scenario: Ref mode no-stream
- **WHEN** `wordflow trans --no-stream --ref "text"` is run
- **THEN** the command SHALL call `translator.Translate(text, &buf, &TransOptions{Ref: true, NoStream: true})` then render with `TranslationRenderer`

#### Scenario: Trans validation
- **WHEN** the selected translator's config is invalid
- **THEN** the `trans` command SHALL return a validation error before attempting translation