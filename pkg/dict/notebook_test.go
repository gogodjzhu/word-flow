package dict

import (
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"testing"
)

func TestFileNotebook_readNote(t *testing.T) {
	type fields struct {
		filename string
	}
	tests := []struct {
		name     string
		fields   fields
		want     []*entity.WordNote
		wantErr  bool
		preFunc  func() bool
		postFunc func() bool
	}{
		{
			name: "test",
			fields: fields{
				filename: "/tmp/test_notebook.yaml",
			},
			want: []*entity.WordNote{
				{
					WordItemId:     "12a032ce9179c32a6c7ab397b9d871fa",
					Word:           "people",
					LookupTimes:    2,
					LastLookupTime: 1696082553,
				},
				{
					WordItemId:     "af238f7fa55be68cec87e800e3606c04",
					Word:           "cheek",
					LookupTimes:    1,
					LastLookupTime: 1696087202,
				},
			},
			wantErr: false,
			preFunc: func() bool {
				return os.WriteFile("/tmp/test_notebook.yaml", []byte(
					`
- word_id: 12a032ce9179c32a6c7ab397b9d871fa
  word: people
  lookup_times: 2
  last_lookup_time: 1696082553
- word_id: af238f7fa55be68cec87e800e3606c04
  word: cheek
  lookup_times: 1
  last_lookup_time: 1696087202
`), 0644) == nil
			},
			postFunc: func() bool {
				return os.Remove("/tmp/test_notebook.yaml") == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fileNotebook{
				directory: tt.fields.filename,
			}
			if tt.preFunc != nil && !tt.preFunc() {
				t.Error("preFunc() failed")
			}
			got, err := f.readNote()
			if (err != nil) != tt.wantErr {
				t.Errorf("readNote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readNote() got = %v, want %v", got, tt.want)
			}
			if tt.postFunc != nil && !tt.postFunc() {
				t.Error("postFunc failed")
			}
		})
	}
}

func TestFileNotebook_writeNote(t *testing.T) {
	type fields struct {
		filename string
	}
	type args struct {
		notes []*entity.WordNote
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		preFunc  func() bool
		postFunc func(args args) bool
	}{
		{
			name: "test",
			fields: fields{
				filename: "/tmp/test_notebook.yaml",
			},
			args: args{
				notes: []*entity.WordNote{
					{
						WordItemId:     "12a032ce9179c32a6c7ab397b9d871fa",
						Word:           "people",
						LookupTimes:    2,
						LastLookupTime: 1696082553,
					},
				},
			},
			wantErr: false,
			preFunc: func() bool {
				_ = os.Remove("/tmp/test_notebook.yaml")
				return true
			},
			postFunc: func(args args) bool {
				b, err := os.ReadFile("/tmp/test_notebook.yaml")
				if err != nil {
					return false
				}
				var notes []*entity.WordNote
				if err := yaml.Unmarshal(b, &notes); err != nil {
					return false
				}
				return reflect.DeepEqual(notes, args.notes)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fileNotebook{
				directory: tt.fields.filename,
			}
			if tt.preFunc != nil && !tt.preFunc() {
				t.Error("preFunc() failed")
			}
			if err := f.writeNote(tt.args.notes); (err != nil) != tt.wantErr {
				t.Errorf("writeNote() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.postFunc != nil && !tt.postFunc(tt.args) {
				t.Error("postFunc failed")
			}
		})
	}
}
