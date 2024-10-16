package dict

import (
	"fmt"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/gogodjzhu/word-flow/pkg/dict"
	"github.com/spf13/cobra"
	"strings"
)

func NewCmdDict(f *cmdutil.Factory) (*cobra.Command, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	var notebookDefault string
	var dictionaryDefault string
	cmd := &cobra.Command{
		Use:   "dict <word>",
		Short: "Look up the word in the dictionary",
		Long:  "Look up the word in the dictionary, you can specify the dictionary by option",
		RunE: func(cmd *cobra.Command, args []string) error {
			/* lookup the word in the dictionary */
			dictionary, err := dict.NewDict(cfg.Dict)
			if err != nil {
				return err
			}
			wordItem, err := dictionary.Search(strings.TrimSpace(strings.Join(args, " ")))
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(f.IOStreams.Out, wordItem.RenderString())

			/* mark the word as learning in the notebook */
			notebook, err := dict.OpenNotebook(cfg.Notebook)
			if err != nil {
				return err
			}
			if _, err := notebook.Mark(wordItem.Word, dict.Learning); err != nil {
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
	return cmd, nil
}
