package config

import (
	"os"
	"reflect"
	"testing"
)

func TestInitConfig(t *testing.T) {
	tmpConfigFilename := "/tmp/config.yaml"
	t.Cleanup(func() {
		_ = os.Remove(tmpConfigFilename)
	})
	type args struct {
		configFilename string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		prevFunc func(args args) bool
		postFunc func(args args) bool
	}{
		{
			name: "Success",
			args: args{
				configFilename: tmpConfigFilename,
			},
			wantErr: false,
			prevFunc: func(args args) bool {
				return true
			},
			postFunc: func(args args) bool {
				_, err := os.ReadFile(args.configFilename)
				return err == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ret := tt.prevFunc(tt.args); !ret {
				t.Errorf("InitConfig() prevFunc = %v", ret)
			}
			if err := InitConfig(tt.args.configFilename); (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ret := tt.postFunc(tt.args); !ret {
				t.Errorf("InitConfig() postFunc = %v", ret)
			}
		})
	}
}

func TestReadConfigSpecified(t *testing.T) {
	tmpConfigFilename := "/tmp/config.yaml"
	t.Cleanup(func() {
		_ = os.Remove(tmpConfigFilename)
	})
	type args struct {
		configFilename string
	}
	conf := defaultConfig
	tests := []struct {
		name     string
		args     args
		want     *Config
		wantErr  bool
		prevFunc func(args args) bool
		postFunc func(got *Config, args args) bool
	}{
		{
			name: "Success",
			args: args{
				configFilename: tmpConfigFilename,
			},
			want:    &conf,
			wantErr: false,
			prevFunc: func(args args) bool {
				conf.Common.ConfigFilename = args.configFilename
				return true
			},
			postFunc: func(got *Config, args args) bool {
				return true
			},
		},
		{
			name: "Update",
			args: args{
				configFilename: tmpConfigFilename,
			},
			want:    &conf,
			wantErr: false,
			prevFunc: func(args args) bool {
				conf.Common.ConfigFilename = args.configFilename
				conf.Common.BasePath = "/tmp/update"
				return conf.Save() == nil
			},
			postFunc: func(got *Config, args args) bool {
				return reflect.DeepEqual(got, &conf)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ret := tt.prevFunc(tt.args); !ret {
				t.Errorf("ReadConfigSpecified() prevFunc = %v", ret)
			}
			got, err := ReadConfigSpecified(tt.args.configFilename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfigSpecified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfigSpecified() got = %v, want %v", got, tt.want)
			}
			if ret := tt.postFunc(got, tt.args); !ret {
				t.Errorf("ReadConfigSpecified() postFunc = %v", ret)
			}
		})
	}
}
