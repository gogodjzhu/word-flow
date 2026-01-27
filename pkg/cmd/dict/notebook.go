package dict

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
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

	var op string
	var notebook string
	cmd := &cobra.Command{
		Use:   "notebook <word>",
		Short: "Learning words in notebook",
		RunE: func(cmd *cobra.Command, args []string) error {
			notebook, err := dict.OpenNotebook(cfg.Notebook)
			if err != nil {
				return err
			}
			var model tea.Model
			switch op {
			case "review":
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
				model = tui_list.NewApp("Words review", initOptions, []tui_list.CallbackFunc{
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
			case "exam":
				// Get due words for review
				dueWords, err := notebook.GetDueWords()
				if err != nil {
					return err
				}

				if len(dueWords) == 0 {
					_, _ = fmt.Fprintln(f.IOStreams.Out, "🎉 No words due for review!")
					_, _ = fmt.Fprintln(f.IOStreams.Out, "💡 Add some words to your notebook first using 'wordflow notebook -o review'")
					return nil
				}

				// Limit session size based on configuration
				maxReviews := 50 // default
				if max, ok := cfg.Notebook.Parameters["fsrs.max_reviews_per_session"].(int); ok {
					maxReviews = max
				}

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

					// Save updated FSRS cards
					for _, word := range results.Words {
						if word != nil && word.FSRSCard != nil {
							if err := notebook.UpdateFSRSCard(word.WordItemId, word.FSRSCard); err != nil {
								_, _ = fmt.Fprintf(f.IOStreams.Out, "[Err] Failed to save FSRS card for %s: %v\n", word.Word, err)
							}
						}
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
			default:
				return fmt.Errorf("unknown operation: %s", op)
			}

			// Only run TUI program if model is not nil (for "review" operation)
			if model != nil {
				_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
				return err
			}

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			if notebook == cfg.Notebook.Default {
				return nil
			}
			cfg.Notebook.Default = notebook
			return cfg.Save()
		},
	}

	cmd.Flags().StringVarP(&op, "operation", "o", "review", "Specify operation, exam or review")
	cmd.Flags().StringVarP(&notebook, "notebook", "n", cfg.Notebook.Default, "Specify notebook name")
	return cmd, nil
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
