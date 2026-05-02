## ADDED Requirements

### Requirement: Google dictionary provider implementation
The system SHALL implement a `DictGoogle` struct in `pkg/dict/google/dict_google.go` that satisfies the `Dict` interface with a `Search(word string) (*entity.WordItem, error)` method.

#### Scenario: Look up a word with dictionary data
- **WHEN** `Search("ephemeral")` is called
- **THEN** the implementation SHALL call `https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=zh-CN&dt=t&dt=bd&q=ephemeral` and return a `WordItem` with:
  - `Word`: the searched word
  - `Source`: "google"
  - `WordMeanings`: parsed from the `dt=bd` response (part of speech + translations)
  - `WordPhonetics`: nil (Google free API does not provide phonetics)

#### Scenario: Look up a word without dictionary data
- **WHEN** `Search("serendipity")` is called and the API returns translation but no `dt=bd` dictionary data (second element is null)
- **THEN** the implementation SHALL fall back to using the translation result as a single `WordMeaning` with empty `PartOfSpeech`

#### Scenario: Word not found or empty response
- **WHEN** the API returns an empty or invalid response
- **THEN** the implementation SHALL return an appropriate error

#### Scenario: Network error
- **WHEN** the Google API request fails
- **THEN** the implementation SHALL return a wrapped error with context

### Requirement: Google Dict endpoint registration
The system SHALL register `"google"` as a `Dict` endpoint constant and add the corresponding factory case in `NewDict`.

#### Scenario: Create Google dict instance
- **WHEN** `NewDict` is called with `conf.Default == "google"`
- **THEN** it SHALL return a `*DictGoogle` instance

#### Scenario: Google in AvailableDictionaries
- **WHEN** `AvailableDictionaries()` is called
- **THEN** the list SHALL include an entry for "google" with description "[Free] Online dictionary and translation powered by Google Translate."

### Requirement: Google dictionary response parsing
The `DictGoogle` SHALL parse the `dt=bd` response array into `WordItem` format:
- The second top-level array element (index 1) contains word class entries
- Each word class entry: `[0]` = part of speech (e.g., "adjective"), `[1]` = list of translations, `[2]` = detailed definitions with synonyms
- Map part of speech to `WordMeaning.PartOfSpeech` (abbreviated, e.g., "adjective" → "adj.")
- Map translations to `WordMeaning.Definitions` (joined with "; ")

#### Scenario: Parse multi-pos word
- **WHEN** looking up "ephemeral" which has both adjective and noun definitions
- **THEN** the result SHALL contain multiple `WordMeaning` entries, one for each part of speech

#### Scenario: Parse word with only translation (no bd data)
- **WHEN** looking up "serendipity" which has translation but no dictionary data
- **THEN** the result SHALL contain a single `WordMeaning` with the translation as Definitions and empty PartOfSpeech

### Requirement: GoogleConfig for Dict
The system SHALL add a `GoogleConfig` struct (empty) to `DictConfig` and `internal/config/dict_configs.go`, implementing `DictEndpointConfig` with `Validate()` returning nil.

#### Scenario: Default Google dict config
- **WHEN** config is loaded without explicit google dict settings
- **THEN** the system SHALL use a default `GoogleConfig` with zero values

#### Scenario: Google dict config validation
- **WHEN** `dict.default` is "google"
- **THEN** validation SHALL pass (GoogleConfig requires no fields)