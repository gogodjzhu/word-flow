## 1. Config Foundation

- [x] 1.1 Add `TransBaiduConfig` struct to `internal/config/config.go` with `AppID` and `Secret` fields
- [x] 1.2 Add `Baidu *TransBaiduConfig` field to `TransConfig` struct
- [x] 1.3 Update `GetEndpointConfig()` in `TransConfig` to handle `"baidu"` case
- [x] 1.4 Update `applyDefaults()` to initialize `Trans.Baidu` with default `TransBaiduConfig`
- [x] 1.5 Update `ConfigTemplate()` to include `baidu: {}` section with `app_id` and `secret` fields (commented out)
- [x] 1.6 Update `ValidateForTrans()` to validate Baidu config (check AppID and Secret are not empty)

## 2. Baidu Translator Package

- [x] 2.1 Create `pkg/translator/baidu/trans_baidu.go` with `TranslatorBaidu` struct and constructor `NewTranslatorBaidu(cfg *config.TransBaiduConfig)`
- [x] 2.2 Implement `Translate(text string, out io.Writer, opts *types.TransOptions) error`
- [x] 2.3 Implement signature generation: `md5(appid + text + salt + secret)`
- [x] 2.4 Implement Baidu API HTTP call with proper POST parameters
- [x] 2.5 Implement response parsing: extract `trans_result[].dst` values
- [x] 2.6 Handle ref mode: output JSON `[{raw: src, translation: dst}]` format
- [x] 2.7 Handle no-stream mode: collect all results and write in one operation
- [x] 2.8 Handle network errors and empty responses with wrapped errors

## 3. Translator Factory Updates

- [x] 3.1 Add `TransBaidu Endpoint = "baidu"` constant to `pkg/translator/translator.go`
- [x] 3.2 Add `case TransBaidu:` to `NewTranslator()` switch, returning `trans_baidu.NewTranslatorBaidu()`
- [x] 3.3 Add Baidu entry to `AvailableTranslators()` with name and description
- [x] 3.4 Update `translator_test.go` to include Baidu translator tests

## 4. Command Integration

- [x] 4.1 Update `pkg/cmd/trans/trans.go` to document `--endpoint=baidu` option
- [x] 4.2 Verify existing `--endpoint` flag works with new "baidu" endpoint

## 5. Tests & Verification

- [x] 5.1 Add unit tests for Baidu `Translate()` with various `TransOptions` combinations
- [x] 5.2 Add unit tests for signature generation
- [x] 5.3 Add unit tests for response parsing
- [x] 5.4 Add unit tests for `NewTranslator` factory with Baidu config
- [x] 5.5 Manually test `wordflow trans "hello world" --endpoint=baidu`
- [x] 5.6 Manually test `wordflow trans "hello" --ref --endpoint=baidu`
- [x] 5.7 Run existing tests to verify no regressions