package dict

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil/tui/tui_exam"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil/tui/tui_list"
	"github.com/gogodjzhu/word-flow/pkg/dict"
	"github.com/gogodjzhu/word-flow/pkg/dict/fsrs"
	"github.com/spf13/cobra"
)

func NewCmdNotebook(f *cmdutil.Factory) (*cobra.Command, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	var notebook string
	originalNotebook := cfg.Notebook.Default
	cmd := &cobra.Command{
		Use:   "notebook",
		Short: "Learning words in notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg.Notebook.Default = notebook
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if notebook == originalNotebook {
				return nil
			}
			cfg.Notebook.Default = notebook
			return cfg.Save()
		},
	}

	cmd.PersistentFlags().StringVarP(&notebook, "notebook", "n", cfg.Notebook.Default, "Specify notebook name")
	cmd.AddCommand(newCmdNotebookReview(f, cfg))
	cmd.AddCommand(newCmdNotebookExam(f, cfg))
	cmd.AddCommand(newCmdNotebookImport(f, cfg))
	return cmd, nil
}

func newCmdNotebookReview(f *cmdutil.Factory, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review words in notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			notebookConfig, err := cfg.Notebook.GetConfig()
			if err != nil {
				return err
			}
			notebook, err := dict.OpenNotebook(notebookConfig)
			if err != nil {
				return err
			}
			notes, err := notebook.ListNotes()
			if err != nil {
				return err
			}
			initOptions := make([]tui_list.OptionEntity, len(notes))
			for i, note := range notes {
				hint := fmt.Sprintf("lookupTimes:%d", note.LookupTimes)
				if note.LastLookupTime > 0 {
					hint = fmt.Sprintf("%s, last: %s", hint, humanize.Time(time.Unix(note.LastLookupTime, 0)))
				}
				initOptions[i] = tui_list.NewOption(&wordItemOptions{
					item:  note.WordItemId,
					title: note.Word,
					hint:  hint,
				})
			}
			model := tui_list.NewApp("Words review", initOptions, []tui_list.CallbackFunc{
				{
					Keys:            []string{"enter"},
					FullDescription: "look up selected word",
					Callback: func(selectedOption tui_list.OptionEntity) []tui_list.OptionEntity {
						words, err := notebook.ListNotes()
						if err != nil {
							_, _ = fmt.Fprintln(f.IOStreams.Out, "[Err] list words failed")
							return nil
						}
						updateOptions := make([]tui_list.OptionEntity, len(words))
						for i, word := range words {
							if word.WordItemId != selectedOption.Entity().(string) {
								hint := fmt.Sprintf("lookupTimes:%d", word.LookupTimes)
								if word.LastLookupTime > 0 {
									hint = fmt.Sprintf("%s, last: %s", hint, humanize.Time(time.Unix(word.LastLookupTime, 0)))
								}
								updateOptions[i] = tui_list.NewOption(&wordItemOptions{
									item:  word.WordItemId,
									title: word.Word,
									hint:  hint,
								})
							} else {
								// Check if cached translation exists
								if word.Translation != "" {
									updateOptions[i] = tui_list.NewOption(&wordItemOptions{
										item:  word.WordItemId,
										title: word.Word,
										hint:  word.Translation,
									})
								} else {
									// Fallback to live API call if no cached translation
									dictionary, err := dict.NewDict(cfg.Dict)
									if err != nil {
										_, _ = fmt.Fprintln(f.IOStreams.Out, "[Err] init dictionary failed")
										return nil
									}
									wordItem, err := dictionary.Search(word.Word)
									if err != nil {
										_, _ = fmt.Fprintln(f.IOStreams.Out, "[Err] search word failed")
										hint := fmt.Sprintf("lookupTimes:%d", word.LookupTimes)
										if word.LastLookupTime > 0 {
											hint = fmt.Sprintf("%s, last: %s", hint, humanize.Time(time.Unix(word.LastLookupTime, 0)))
										}
										updateOptions[i] = tui_list.NewOption(&wordItemOptions{
											item:  word.WordItemId,
											title: word.Word,
											hint:  hint,
										})
									} else {
										// Cache the translation for future use
										if _, err := notebook.Mark(word.Word, dict.Learning, wordItem); err != nil {
											_, _ = fmt.Fprintln(f.IOStreams.Out, "[Err] cache translation failed")
										}
										updateOptions[i] = tui_list.NewOption(&wordItemOptions{
											item:  word.WordItemId,
											title: word.Word,
											hint:  wordItem.RawString(),
										})
									}
								}
							}
						}
						return updateOptions
					},
				},
				{
					Keys:             []string{"x"},
					ShortDescription: "delete",
					FullDescription:  "delete selected word",
					Callback: func(selectedOption tui_list.OptionEntity) []tui_list.OptionEntity {
						if _, err := notebook.Mark(selectedOption.Title(), dict.Delete, nil); err != nil {
							fmt.Fprintln(f.IOStreams.Out, "[Err] mark word failed")
						}
						words, err := notebook.ListNotes()
						if err != nil {
							fmt.Fprintln(f.IOStreams.Out, "[Err] list words failed")
							return nil
						}
						updateOptions := make([]tui_list.OptionEntity, len(words))
						for i, word := range words {
							hint := fmt.Sprintf("lookupTimes:%d", word.LookupTimes)
							if word.LastLookupTime > 0 {
								hint = fmt.Sprintf("%s, last: %s", hint, humanize.Time(time.Unix(word.LastLookupTime, 0)))
							}
							updateOptions[i] = tui_list.NewOption(&wordItemOptions{
								item:  word.WordItemId,
								title: word.Word,
								hint:  hint,
							})
						}
						return updateOptions
					},
				},
			})

			_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
			return err
		},
	}
	return cmd
}

func newCmdNotebookExam(f *cmdutil.Factory, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exam",
		Short: "Review due words with FSRS",
		RunE: func(cmd *cobra.Command, args []string) error {
			notebookConfig, err := cfg.Notebook.GetConfig()
			if err != nil {
				return err
			}
			notebook, err := dict.OpenNotebook(notebookConfig)
			if err != nil {
				return err
			}
			// Get due words for review
			dueWords, err := notebook.GetDueWords()
			if err != nil {
				return err
			}

			if len(dueWords) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "🎉 No words due for review!")
				_, _ = fmt.Fprintln(f.IOStreams.Out, "💡 Add some words to your notebook first using 'wordflow notebook review'")
				return nil
			}

			// Limit session size based on configuration
			maxReviews := notebookConfig.MaxReviews

			if len(dueWords) > maxReviews {
				dueWords = dueWords[:maxReviews]
				_, _ = fmt.Fprintf(f.IOStreams.Out, "Limited to %d words for this session\n", maxReviews)
			}

			// Initialize FSRS scheduler
			scheduler := fsrs.NewScheduler()

			// Create exam TUI model
			examModel := tui_exam.NewModel(dueWords, scheduler)

			// Run the exam
			program := tea.NewProgram(examModel, tea.WithAltScreen())
			result, err := program.Run()
			if err != nil {
				return err
			}

			// Get results and save updated cards
			if examResult, ok := result.(tui_exam.Model); ok {
				results := examResult.GetResults()

				// Save updated notes and FSRS cards
				if err := notebook.SaveExamResults(results.Words); err != nil {
					_, _ = fmt.Fprintf(f.IOStreams.Out, "[Err] Failed to save exam results: %v\n", err)
				}

				// Show summary
				fmt.Fprintf(f.IOStreams.Out, "\n✅ Session completed!\n")
				fmt.Fprintf(f.IOStreams.Out, "   Reviewed: %d words\n", results.Completed)
				if results.Skipped > 0 {
					fmt.Fprintf(f.IOStreams.Out, "   Skipped: %d words\n", results.Skipped)
				}
				fmt.Fprintf(f.IOStreams.Out, "   Duration: %s\n", results.Duration.Round(time.Second))
			} else {
				fmt.Fprintf(f.IOStreams.Out, "[Warning] Failed to get exam results\n")
			}
			return nil
		},
	}
	return cmd
}

func newCmdNotebookImport(f *cmdutil.Factory, cfg *config.Config) *cobra.Command {
	var importFile string
	var importFormat string
	var dictionary string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import words into notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			notebookConfig, err := cfg.Notebook.GetConfig()
			if err != nil {
				return err
			}
			notebook, err := dict.OpenNotebook(notebookConfig)
			if err != nil {
				return err
			}
			if importFormat != "tsv" {
				return fmt.Errorf("unsupported import format: %s", importFormat)
			}
			if strings.TrimSpace(importFile) == "" {
				return fmt.Errorf("import file is required")
			}
			words, err := readImportWordsTSV(importFile)
			if err != nil {
				return err
			}
			if len(words) == 0 {
				return fmt.Errorf("no valid words found in %s", importFile)
			}
			dictConfig := *cfg.Dict
			dictConfig.Default = dictionary
			dictionaryClient, err := dict.NewDict(&dictConfig)
			if err != nil {
				return err
			}

			imported := 0
			for _, word := range words {
				wordItem, err := dictionaryClient.Search(word)
				if err != nil {
					_, _ = fmt.Fprintf(f.IOStreams.Out, "[Err] search word failed: %s (%v)\n", word, err)
					continue
				}
				if _, err := notebook.Mark(wordItem.Word, dict.Learning, wordItem); err != nil {
					_, _ = fmt.Fprintf(f.IOStreams.Out, "[Err] save word failed: %s (%v)\n", wordItem.Word, err)
					continue
				}
				imported++
			}
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Imported %d/%d words\n", imported, len(words))
			return nil
		},
	}
	cmd.Flags().StringVarP(&importFile, "input", "i", "", "Specify input file for import")
	cmd.Flags().StringVarP(&importFormat, "format", "f", "tsv", "Specify import format (tsv)")
	cmd.Flags().StringVarP(&dictionary, "dictionary", "d", string(dict.Youdao), "Specify dictionary for import")
	return cmd
}

type wordItemOptions struct {
	item  string
	title string
	hint  string
}

func (w *wordItemOptions) Entity() interface{} {
	return w.item
}

func (w *wordItemOptions) Title() string {
	return w.title
}

func (w *wordItemOptions) Description() string {
	return w.hint
}

func readImportWordsTSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	var words []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		word := strings.TrimSpace(fields[0])
		if word == "" {
			continue
		}
		words = append(words, word)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}
