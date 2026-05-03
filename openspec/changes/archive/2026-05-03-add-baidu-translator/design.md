## Context

Word-flow currently uses Google Translate as the default translator. Google Translate's free API (`translate.googleapis.com`) is unstable in China due to network issues. This change adds Baidu Translate as a reliable alternative using their official API with free quota (2M chars/month after identity verification).

## Goals / Non-Goals

**Goals:**
- Add Baidu Translate API as a new translator provider
- Maintain backward compatibility with Google translator
- Support `--ref` (side-by-side) and `--no-stream` modes
- Change default translator from `google` to `baidu`

**Non-Goals:**
- Multi-language support (English-to-Chinese only for now)
- Official Baidu Cloud Translation API (uses direct API, not cloud)
- Combining dict and trans config
- Removing Google translator (keep as fallback)

## Decisions

### D1: Baidu API authentication via appid + secret + md5 sign

**Decision**: Baidu Translate API requires `appid`, `secret`, and a signature computed as `md5(appid + text + salt + secret)`.

**Rationale**: This is Baidu's standard authentication mechanism. The signature prevents tampering and ensures only registered apps can use the API.

### D2: Baidu API endpoint and request format

**Decision**: Use `https://api.fanyi.baidu.com/api/trans/vip/translate` with POST parameters:
- `q`: text to translate
- `from`: source language
- `to`: target language
- `appid`: application ID
- `salt`: random number (timestamp-based)
- `sign`: md5 signature

**Response format**:
```json
{
  "from": "en",
  "to": "zh",
  "trans_result": [
    {"src": "hello", "dst": "你好"}
  ]
}
```

**Rationale**: Direct API call with proper authentication. Baidu supports auto language detection (`from: "auto"`).

### D3: No sentence batching needed for Baidu

**Decision**: Unlike Google Translate's free API which requires batching to simulate streaming, Baidu's API handles larger texts efficiently and returns complete translations. We can send the full text in a single request.

**Rationale**: Baidu's API is designed to handle real translation workloads without the streaming simulation complexity that Google required.

### D4: Ref mode maps to src/dst pairs

**Decision**: Baidu's response provides `src` (original) and `dst` (translation) per segment. For ref mode, output JSON array `[{raw: src, translation: dst}]` matching existing format.

**Rationale**: Existing `--ref` rendering pipeline expects this format. Baidu's src/dst structure maps directly.

### D5: TransBaiduConfig with app_id and secret fields

**Decision**: `TransBaiduConfig` struct contains:
```go
type TransBaiduConfig struct {
    AppID  string `yaml:"app_id"`
    Secret string `yaml:"secret"`
}
```

**Rationale**: Baidu requires both AppID and Secret for authentication. These are user-specific credentials from the Baidu developer console.

### D6: Default translator changes from google to baidu

**Decision**: `trans.default: "baidu"` becomes the new default, replacing `google`.

**Rationale**: Baidu is more reliable in China. Google remains available via `--endpoint=google` or by setting `trans.default: google` in config.

## Risks / Trade-offs

- **[Baidu quota]** → 2M chars/month free tier. Exceeding requires paid upgrade. Mitigation: monitor usage, switch to Google if needed.
- **[Network to Baidu]** → Baidu API may have connectivity issues in some regions. Mitigation: Google remains as fallback.
- **[Config migration]** → Existing users with default Google will automatically switch to Baidu. Google users should explicitly set `trans.default: google` if needed.
- **[Baidu newline handling]** → Baidu API treats newlines as spaces when translating multi-line input. The API returns translated segments that we concatenate with spaces. Use Google endpoint (`--endpoint=google`) if newline preservation is important.

## Open Questions

- Should we validate AppID/Secret format during config validation?
- Do we need to handle Baidu's specific error codes (52003, etc.) with user-friendly messages?