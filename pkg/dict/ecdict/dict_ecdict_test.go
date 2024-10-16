package dict_ecdict

import (
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestDictEcdict_Search(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("/tmp/stardict.db"), &gorm.Config{})
	if err != nil {
		t.Error(err)
	}
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		word string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		postFunc func(got *entity.WordItem, args args) bool
	}{
		{
			name: "Test1",
			fields: fields{
				db: db,
			},
			args: args{
				word: "note",
			},
			wantErr: false,
			postFunc: func(got *entity.WordItem, args args) bool {
				if got == nil || got.ID == "" || len(got.WordMeanings) == 0 || len(got.WordPhonetics) == 0 {
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DictEcdict{
				db: tt.fields.db,
			}
			got, err := d.Search(tt.args.word)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.postFunc != nil && !tt.postFunc(got, tt.args) {
				t.Errorf("Search() postFunc failed")
			}
		})
	}
}
