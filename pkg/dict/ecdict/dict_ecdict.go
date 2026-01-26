package dict_ecdict

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DictEcdict struct {
	db *gorm.DB
}

func NewDictEcdit(config *config.EcdictConfig) (*DictEcdict, error) {
	if config == nil {
		return nil, errors.New("ecdict config is required")
	}
	err := PrepareDBFile(config.DBFilename)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(config.DBFilename), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to open db file")
	}
	return &DictEcdict{
		db: db,
	}, nil
}

func PrepareDBFile(dbfilename string) error {
	_, err := os.Stat(dbfilename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		url := "https://github.com/skywind3000/ECDICT/releases/download/1.0.28/ecdict-sqlite-28.zip"
		tmpZip := filepath.Join(os.TempDir(), "ecdict-sqlite-28.zip")

		if err := downloadWithProgress(url, tmpZip); err != nil {
			return errors.Wrap(err, "failed to download ecdict zip")
		}
		if err := unzipSqliteTo(tmpZip, dbfilename); err != nil {
			return errors.Wrap(err, "failed to extract sqlite from zip")
		}
		return nil
	}
	return errors.Wrap(err, "failed to prepare db file")
}

func (d *DictEcdict) Search(word string) (*entity.WordItem, error) {
	var wordItem Word
	d.db.Raw("select * from stardict where word = ?", word).Scan(&wordItem)
	result := &entity.WordItem{
		ID:            entity.WordId(wordItem.Word),
		Word:          wordItem.Word,
		WordPhonetics: make([]*entity.WordPhonetic, 0),
		WordMeanings:  make([]*entity.WordMeaning, 0),
	}
	result.WordPhonetics = append(result.WordPhonetics, &entity.WordPhonetic{
		Text: wordItem.Phonetic,
	})
	definitions := strings.Split(wordItem.Translation, "\n")
	for _, definition := range definitions {
		var partOfSpeechStr string
		var definitionStr string
		dotPos := strings.Index(definition, ".")
		if dotPos != -1 {
			partOfSpeechStr = definition[:dotPos+1]
		}
		definitionStr = definition[dotPos+1:]
		result.WordMeanings = append(result.WordMeanings, &entity.WordMeaning{
			PartOfSpeech: partOfSpeechStr,
			Definitions:  definitionStr,
		})
	}
	return result, nil
}

type Word struct {
	gorm.Model
	Id          int64  `gorm:"column:id"`
	Word        string `gorm:"column:word"`
	ShortWorld  string `gorm:"column:sw"`
	Phonetic    string
	Definition  string
	Translation string
	Pos         string
	Collins     string
	Oxford      string
	Tag         string
	Bnc         string
	Frq         string
	Exchange    string
	Detail      string
	Audio       string
}

func (Word) TableName() string {
	return "stardict"
}

// downloadWithProgress downloads a file from url to dest with inline progress.
func downloadWithProgress(url, dest string) error {
	fmt.Printf("Download %s to %s\n", url, dest)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	var downloaded int64 = 0
	lastPrinted := time.Now()
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			// Print progress every 300ms
			if time.Since(lastPrinted) > 300*time.Millisecond {
				printProgress(downloaded, total)
				lastPrinted = time.Now()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}
	// Final progress 100%
	printProgress(downloaded, total)
	fmt.Print("\n")
	return nil
}

func printProgress(done, total int64) {
	if total <= 0 {
		fmt.Printf("Downloading... %d bytes\n", done)
		return
	}
	percent := float64(done) / float64(total) * 100.0
	// Inline progress: carriage return without newline
	fmt.Printf("\rDownloading: %.1f%% (%d/%d bytes)", percent, done, total)
}

// unzipSqliteTo extracts the first .sqlite or .db file from zipPath to destDb.
func unzipSqliteTo(zipPath, destDb string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var srcFile *zip.File
	for _, f := range r.File {
		nameLower := strings.ToLower(f.Name)
		if strings.HasSuffix(nameLower, ".sqlite") || strings.HasSuffix(nameLower, ".db") {
			srcFile = f
			break
		}
	}
	if srcFile == nil {
		return errors.New("no sqlite/db file found in zip")
	}

	rc, err := srcFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destDb), 0755); err != nil {
		return err
	}
	out, err := os.Create(destDb)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return err
	}
	return nil
}
