package baidu

import (
	"encoding/json"
	"testing"

	"github.com/gogodjzhu/word-flow/internal/config"
)

func TestGenerateSign_DifferentInputs(t *testing.T) {
	appid := "test_app"
	text := "test_text"
	salt := "12345"
	secret := "test_secret"

	sign1 := generateSign(appid, text, salt, secret)
	sign2 := generateSign(appid, text, salt, secret)

	if sign1 != sign2 {
		t.Error("sign should be deterministic")
	}

	sign3 := generateSign(appid, "different_text", salt, secret)
	if sign1 == sign3 {
		t.Error("sign should change with different text")
	}
}

func TestBaiduResponse_Parsing(t *testing.T) {
	jsonStr := `{"from":"en","to":"zh","trans_result":[{"src":"hello","dst":"你好"},{"src":"world","dst":"世界"}]}`

	var resp baiduResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.From != "en" {
		t.Errorf("expected from 'en', got %s", resp.From)
	}
	if resp.To != "zh" {
		t.Errorf("expected to 'zh', got %s", resp.To)
	}
	if len(resp.TransResult) != 2 {
		t.Errorf("expected 2 trans results, got %d", len(resp.TransResult))
	}
	if resp.TransResult[0].Src != "hello" {
		t.Errorf("expected src 'hello', got %s", resp.TransResult[0].Src)
	}
	if resp.TransResult[0].Dst != "你好" {
		t.Errorf("expected dst '你好', got %s", resp.TransResult[0].Dst)
	}
}

func TestNewTranslatorBaidu(t *testing.T) {
	cfg := &config.TransBaiduConfig{
		AppID:  "test_app_id",
		Secret: "test_secret",
	}

	trans := NewTranslatorBaidu(cfg)
	if trans == nil {
		t.Fatal("expected non-nil translator")
	}
	if trans.cfg != cfg {
		t.Error("expected config to be set")
	}
}

func TestTransBaiduConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.TransBaiduConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     config.TransBaiduConfig{AppID: "app_id", Secret: "secret"},
			wantErr: false,
		},
		{
			name:    "missing app_id",
			cfg:     config.TransBaiduConfig{AppID: "", Secret: "secret"},
			wantErr: true,
		},
		{
			name:    "missing secret",
			cfg:     config.TransBaiduConfig{AppID: "app_id", Secret: ""},
			wantErr: true,
		},
		{
			name:    "both missing",
			cfg:     config.TransBaiduConfig{AppID: "", Secret: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}