## ADDED Requirements

### Requirement: Baidu Translate provider implementation
The system SHALL implement a `TranslatorBaidu` struct in `pkg/translator/baidu/trans_baidu.go` that satisfies the `Translator` interface.

#### Scenario: Simple translation
- **WHEN** `Translate("hello world", out, nil)` is called
- **THEN** the translator SHALL call the Baidu Translate API with proper authentication and write the translated text to `out`

#### Scenario: Translation with NoStream mode
- **WHEN** `Translate` is called with `opts.NoStream == true`
- **THEN** the translator SHALL send the full text to Baidu API and write the complete translation to `out` in a single write

#### Scenario: Ref mode translation
- **WHEN** `Translate` is called with `opts.Ref == true`
- **THEN** the translator SHALL output JSON array `[{raw: "<original>", translation: "<translated>"}]` to the io.Writer

#### Scenario: Network error handling
- **WHEN** the Baidu API request fails due to network issues
- **THEN** the translator SHALL return a wrapped error with context about the failure

### Requirement: Baidu Translate API URL construction
The `TranslatorBaidu` SHALL use the API endpoint `https://api.fanyi.baidu.com/api/trans/vip/translate` with POST parameters: `q`, `from`, `to`, `appid`, `salt`, `sign`.

#### Scenario: Signature generation
- **WHEN** a translation request is made
- **THEN** the signature SHALL be computed as `md5(appid + text + salt + secret)`

#### Scenario: Language auto-detection
- **WHEN** `from` parameter is set to `"auto"`
- **THEN** the Baidu API SHALL automatically detect the source language

### Requirement: Baidu translation response parsing
The `TranslatorBaidu` SHALL parse the JSON response from the Baidu Translate API and extract translations from `trans_result`.

#### Scenario: Parse standard response
- **WHEN** the API returns `{"from": "en", "to": "zh", "trans_result": [{"src": "hello", "dst": "你好"}]}`
- **THEN** the translator SHALL extract "你好" as the translated text

#### Scenario: Multiple translation segments
- **WHEN** the API returns multiple segments in `trans_result`
- **THEN** the translator SHALL concatenate all `dst` values in order

### Requirement: Baidu translator config
The system SHALL define a `TransBaiduConfig` struct in `internal/config/config.go` with fields:
- `AppID string` — Baidu application ID
- `Secret string` — Baidu application secret

#### Scenario: Default Baidu config
- **WHEN** config is loaded without explicit baidu settings
- **THEN** the system SHALL use a default `TransBaiduConfig` with empty strings

#### Scenario: Config validation
- **WHEN** Baidu is selected as the translator and `AppID` or `Secret` is empty
- **THEN** the validator SHALL return an error indicating missing credentials