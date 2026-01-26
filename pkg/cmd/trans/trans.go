package trans

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/gogodjzhu/word-flow/internal/buzz_error"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/llm"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCmdTrans(f *cmdutil.Factory) (*cobra.Command, error) {
	var useStdin bool
	var noStream bool
	var ref bool

	cmd := &cobra.Command{
		Use:   "trans [text]",
		Short: "Translate English text to Chinese using LLM",
		Long: `Translate English text to Chinese using LLM.
Supports both command line arguments and stdin (pipe) input.
When --ref is enabled, shows original and translation in segment pairs.
Use --no-stream to get formatted output with --ref.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return errors.Wrap(err, "failed to get config")
			}

			llmConfig, err := cfg.Dict.GetConfigForEndpoint("llm")
			if err != nil {
				return errors.Wrap(err, "failed to get LLM config")
			}

			// Get input text
			var text string
			if len(args) > 0 {
				// Command line arguments
				text = strings.Join(args, " ")
			} else {
				// Read from stdin (pipe)
				text, err = readFromStdin(f.IOStreams.In)
				if err != nil {
					return errors.Wrap(err, "failed to read from stdin")
				}
				if strings.TrimSpace(text) == "" {
					return buzz_error.InvalidInput("No input text provided")
				}
			}

			// Trim whitespace for validation
			trimmedText := strings.TrimSpace(text)
			if trimmedText == "" {
				return buzz_error.InvalidInput("No input text provided")
			}

			// Create LLM client
			llmCfg := llmConfig.(*config.LLMConfig)
			client := llm.NewClient(
				llmCfg.ApiKey,
				llmCfg.URL,
				llmCfg.Model,
				llmCfg.Timeout,
				llmCfg.MaxTokens,
				llmCfg.Temperature,
			)

			// Translate text
			if noStream {
				// Non-streaming mode - capture translation first
				var buf bytes.Buffer
				translation, err := client.TranslateWithStream(text, true, &buf, ref)
				if err != nil {
					return errors.Wrap(err, "failed to translate text")
				}

				// Output with formatting
				if ref {
					err = renderWithRef(f.IOStreams.Renderer, translation, f.IOStreams.Out)
				} else {
					err = f.IOStreams.Renderer.RenderText(translation, f.IOStreams.Out)
				}
				if err != nil {
					return errors.Wrap(err, "failed to render translation")
				}
			} else if ref {
				// True streaming mode with ref - capture and process in real-time
				pr, pw := io.Pipe()
				go func() {
					defer pw.Close()
					_, err := client.TranslateWithStream(text, false, pw, ref)
					if err != nil {
						_ = pw.CloseWithError(err)
					}
				}()

				// Process the stream as it comes
				streamReader := NewSimpleStreamReader(pr, f.IOStreams.Renderer, f.IOStreams.Out)
				err := streamReader.Process()
				if err != nil {
					return errors.Wrap(err, "failed to process stream")
				}
			} else {
				// Regular streaming mode
				_, err := client.TranslateWithStream(text, false, f.IOStreams.Out, ref)
				if err != nil {
					return errors.Wrap(err, "failed to translate text")
				}
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return errors.Wrap(err, "failed to get config")
			}
			return cfg.Save()
		},
	}

	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read from stdin (pipe)")
	cmd.Flags().BoolVar(&noStream, "no-stream", false, "Disable streaming output")
	cmd.Flags().BoolVar(&ref, "ref", false, "Show original text with translation in segment pairs")

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

	// Remove trailing newline if present
	text := strings.TrimSuffix(result.String(), "\n")

	return text, nil
}

func renderWithRef(renderer *cmdutil.Renderer, translation string, out io.Writer) error {
	translationRenderer := NewTranslationRenderer(renderer, out)
	return translationRenderer.RenderTranslationWithRef(translation)
}
