package trans

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/gogodjzhu/word-flow/internal/buzz_error"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/gogodjzhu/word-flow/pkg/translator"
	"github.com/gogodjzhu/word-flow/pkg/translator/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCmdTrans(f *cmdutil.Factory) (*cobra.Command, error) {
	var noStream bool
	var ref bool
	var endpoint string

	cfg, err := f.Config()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}
	originalEndpoint := cfg.Trans.Default

	cmd := &cobra.Command{
		Use:   "trans [text]",
		Short: "Translate English text to Chinese",
		Long: `Translate English text to Chinese.
Supports both command line arguments and stdin (pipe) input.
When --ref is enabled, shows original and translation in segment pairs.
Use --no-stream to get formatted output with --ref.
Use --endpoint to override the default translator (google, llm).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return errors.Wrap(err, "failed to get config")
			}

			if endpoint != "" {
				cfg.Trans.Default = endpoint
			}

			if err := config.ValidateForTrans(cfg); err != nil {
				return err
			}

			t, err := translator.NewTranslator(cfg.Trans)
			if err != nil {
				return errors.Wrap(err, "failed to create translator")
			}

			var text string
			if len(args) > 0 {
				text = strings.Join(args, " ")
			} else {
				text, err = readFromStdin(f.IOStreams.In)
				if err != nil {
					return errors.Wrap(err, "failed to read from stdin")
				}
				if strings.TrimSpace(text) == "" {
					return buzz_error.InvalidInput("No input text provided")
				}
			}

			trimmedText := strings.TrimSpace(text)
			if trimmedText == "" {
				return buzz_error.InvalidInput("No input text provided")
			}

			opts := &types.TransOptions{
				Ref:      ref,
				NoStream: noStream,
			}

			if noStream {
				var buf bytes.Buffer
				err := t.Translate(text, &buf, opts)
				if err != nil {
					return errors.Wrap(err, "failed to translate text")
				}
				translation := buf.String()

				if ref {
					err = renderWithRef(f.IOStreams.Renderer, translation, f.IOStreams.Out)
				} else {
					err = f.IOStreams.Renderer.RenderText(translation, f.IOStreams.Out)
				}
				if err != nil {
					return errors.Wrap(err, "failed to render translation")
				}
			} else if ref {
				pr, pw := io.Pipe()
				go func() {
					defer pw.Close()
					err := t.Translate(text, pw, opts)
					if err != nil {
						_ = pw.CloseWithError(err)
					}
				}()

				streamReader := NewSimpleStreamReader(pr, f.IOStreams.Renderer, f.IOStreams.Out)
				err := streamReader.Process()
				if err != nil {
					return errors.Wrap(err, "failed to process stream")
				}
			} else {
				err := t.Translate(text, f.IOStreams.Out, opts)
				if err != nil {
					return errors.Wrap(err, "failed to translate text")
				}
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if endpoint != "" && endpoint != originalEndpoint {
				return config.PatchYAMLFile(cfg.Common.ConfigFilename, "trans.default", endpoint)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&noStream, "no-stream", false, "Disable streaming output")
	cmd.Flags().BoolVar(&ref, "ref", false, "Show original text with translation in segment pairs")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Override default translator (google, llm)")

	return cmd, nil
}

func readFromStdin(r io.Reader) (string, error) {
	var result strings.Builder
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		result.WriteString(scanner.Text())
		result.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		if err == io.EOF {
			return "", nil
		}
		return "", errors.Wrap(err, "failed to read from stdin")
	}

	text := strings.TrimSuffix(result.String(), "\n")

	return text, nil
}

func renderWithRef(renderer *cmdutil.Renderer, translation string, out io.Writer) error {
	translationRenderer := NewTranslationRenderer(renderer, out)
	return translationRenderer.RenderTranslationWithRef(translation)
}