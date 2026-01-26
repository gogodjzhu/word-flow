package dict

import (
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
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

func TestNotebookFilenameFix(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "notebook_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a notebook config with default name
	notebookConfig := &config.NotebookConfig{
		Default: "default",
		Parameters: map[string]interface{}{
			config.NotebookConfigNotebookBasepath: tempDir,
		},
	}

	// Open notebook
	notebook, err := OpenNotebook(notebookConfig)
	if err != nil {
		t.Fatalf("Failed to open notebook: %v", err)
	}

	// Cast to fileNotebook to check filename
	fileNotebook, ok := notebook.(*fileNotebook)
	if !ok {
		t.Fatalf("Expected fileNotebook, got %T", notebook)
	}

	// Verify the filename is correctly constructed
	expectedFilename := filepath.Join(tempDir, "default.yaml")
	if fileNotebook.filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, fileNotebook.filename)
	}

	// Verify the notebook name is set correctly
	if fileNotebook.notebookName != "default" {
		t.Errorf("Expected notebook name 'default', got '%s'", fileNotebook.notebookName)
	}
}

func TestFileNotebook_MarkWithTranslation(t *testing.T) {
	// Setup temporary directory for test
	tempDir := "/tmp/test_notebook_" + t.Name()
	defer os.RemoveAll(tempDir)

	// Create test notebook config
	notebookConfig := &config.NotebookConfig{
		Default: "test",
		Parameters: map[string]interface{}{
			"notebook.basepath": tempDir,
		},
	}

	// Open notebook
	notebook, err := OpenNotebook(notebookConfig)
	if err != nil {
		t.Fatalf("Failed to open notebook: %v", err)
	}

	// Create test translation data
	testTranslation := &entity.WordItem{
		ID:   "test-id",
		Word: "test",
		WordMeanings: []*entity.WordMeaning{
			{
				PartOfSpeech: "n.",
				Definitions:  "a test definition",
			},
		},
	}
	expectedTranslationStr := testTranslation.RawString()

	tests := []struct {
		name        string
		word        string
		action      Action
		translation *entity.WordItem
		wantErr     bool
	}{
		{
			name:        "Add word with translation",
			word:        "test",
			action:      Learning,
			translation: testTranslation,
			wantErr:     false,
		},
		{
			name:        "Update existing word without new translation",
			word:        "test",
			action:      Learning,
			translation: nil,
			wantErr:     false,
		},
		{
			name:        "Add another word with translation",
			word:        "example",
			action:      Learning,
			translation: testTranslation,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note, err := notebook.Mark(tt.word, tt.action, tt.translation)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if note == nil && !tt.wantErr {
				t.Error("Mark() returned nil note")
				return
			}
			if !tt.wantErr {
				if note.Word != tt.word {
					t.Errorf("Mark() word = %v, want %v", note.Word, tt.word)
				}
				// Check if translation is preserved
				if tt.translation != nil && note.Translation == "" {
					t.Error("Mark() expected translation to be preserved")
				}
				if tt.translation != nil && note.Translation != "" {
					if note.Translation != expectedTranslationStr {
						t.Errorf("Mark() translation = %v, want %v", note.Translation, expectedTranslationStr)
					}
				}
			}
		})
	}

	// Verify notes are correctly persisted and can be listed
	notes, err := notebook.ListNotes()
	if err != nil {
		t.Fatalf("Failed to list notes: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}

	// Check that translation data is correctly loaded
	var testNote, exampleNote *entity.WordNote
	for _, note := range notes {
		if note.Word == "test" {
			testNote = note
		} else if note.Word == "example" {
			exampleNote = note
		}
	}

	if testNote == nil {
		t.Error("Test note not found")
	} else if testNote.Translation == "" {
		t.Error("Test note translation not found")
	} else if testNote.Translation != expectedTranslationStr {
		t.Errorf("Expected translation '%s', got '%s'", expectedTranslationStr, testNote.Translation)
	}

	if exampleNote == nil {
		t.Error("Example note not found")
	}
}

func TestFileNotebook_BackwardCompatibility(t *testing.T) {
	// Setup temporary directory for test
	tempDir := "/tmp/test_notebook_compat_" + t.Name()
	defer os.RemoveAll(tempDir)

	// Create notebook file with old format (without translation field)
	notebookFile := filepath.Join(tempDir, "old.yaml")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Write old format YAML
	oldData := `- word_id: "098f6bcd4621d373cade4e832627b4f6"
  word: "test"
  lookup_times: 1
  create_time: 1640995200
  last_lookup_time: 1640995200
`
	err = os.WriteFile(notebookFile, []byte(oldData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Open notebook with old format
	notebookConfig := &config.NotebookConfig{
		Default: "old",
		Parameters: map[string]interface{}{
			"notebook.basepath": tempDir,
		},
	}

	notebook, err := OpenNotebook(notebookConfig)
	if err != nil {
		t.Fatalf("Failed to open notebook: %v", err)
	}

	// List notes should work without errors
	notes, err := notebook.ListNotes()
	if err != nil {
		t.Fatalf("Failed to list notes: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	note := notes[0]
	if note.Word != "test" {
		t.Errorf("Expected word 'test', got '%s'", note.Word)
	}

	// Translation should be empty for backward compatibility
	if note.Translation != "" {
		t.Error("Expected empty translation for backward compatibility")
	}

	// Should be able to add translation to existing note
	testTranslation := &entity.WordItem{
		ID:   "test-id",
		Word: "test",
		WordMeanings: []*entity.WordMeaning{
			{
				PartOfSpeech: "n.",
				Definitions:  "a test definition",
			},
		},
	}
	expectedTranslationStr := testTranslation.RawString()

	_, err = notebook.Mark("test", Learning, testTranslation)
	if err != nil {
		t.Fatalf("Failed to mark with translation: %v", err)
	}

	// Verify translation is now saved
	notes, err = notebook.ListNotes()
	if err != nil {
		t.Fatalf("Failed to list notes after update: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note after update, got %d", len(notes))
	}

	updatedNote := notes[0]
	if updatedNote.Translation == "" {
		t.Error("Expected translation to be saved")
	} else if updatedNote.Translation != expectedTranslationStr {
		t.Errorf("Expected translation '%s', got '%s'", expectedTranslationStr, updatedNote.Translation)
	}
}
