package dict

import (
	"fmt"
	"strings"

	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/gogodjzhu/word-flow/pkg/dict"
	"github.com/spf13/cobra"
)

func NewCmdDict(f *cmdutil.Factory) (*cobra.Command, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	var notebookDefault string
	var dictionaryDefault string
	var list bool
	cmd := &cobra.Command{
		Use:   "dict <word>",
		Short: "Look up the word in the dictionary",
		Long:  "Look up the word in the dictionary, you can specify the dictionary, including youdao, ecdict, etymonline etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If list flag is provided, print available dictionary types and exit.
			if list {
				dicts := dict.AvailableDictionaries()
				for _, d := range dicts {
					_, _ = fmt.Fprintf(f.IOStreams.Out, "%-12s - %s\n", d.Name, d.Description)
				}
				return nil
			}
			/* lookup the word in the dictionary */
			dictionary, err := dict.NewDict(cfg.Dict)
			if err != nil {
				return err
			}
			wordItem, err := dictionary.Search(strings.TrimSpace(strings.Join(args, " ")))
			if err != nil {
				return err
			}
			// Use the new rendering system
			segments := wordItem.Format()
			renderErr := f.IOStreams.Renderer.RenderToWriter(segments, f.IOStreams.Out)
			if renderErr != nil {
				return renderErr
			}

			/* mark the word as learning in the notebook */
			notebook, err := dict.OpenNotebook(cfg.Notebook)
			if err != nil {
				return err
			}
			if _, err := notebook.Mark(wordItem.Word, dict.Learning, wordItem); err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cfg.Notebook.Default = notebookDefault
			cfg.Dict.Default = dictionaryDefault
			return cfg.Save()
		},
	}
	cmd.Flags().StringVarP(&notebookDefault, "notebook", "n", cfg.Notebook.Default, "Specify the notebook")
	cmd.Flags().StringVarP(&dictionaryDefault, "dictionary", "d", cfg.Dict.Default, "Specify the dictionary")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "List available dictionary types")
	return cmd, nil
}
