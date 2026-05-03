## Why

Google Translate (current default) is unstable in China. Adding Baidu Translate as a reliable API-based alternative with official free quota (2M chars/month after identity verification).

## What Changes

- Add `pkg/translator/baidu/` package implementing `Translator` interface
- Add `TransBaiduConfig` to config with `AppID` and `Secret` fields
- Add `baidu` as a new translator endpoint option
- Set `baidu` as new default translator (replacing `google`)
- Keep `google` as a fallback/failsafe option
- Support `Ref` and `NoStream` modes matching existing translator behavior

## Capabilities

### New Capabilities

- `baidu-translator`: API-based translation using Baidu Translate API with app_id/secret authentication

### Modified Capabilities

- `translator-interface`: Add `baidu` to AvailableTranslators and factory switch case
- `trans-config`: Add `trans.baidu` section with `app_id` and `secret` fields

## Impact

- New package: `pkg/translator/baidu/`
- Config changes: `internal/config/config.go` - TransBaiduConfig, TransConfig additions
- Change default from `google` to `baidu`