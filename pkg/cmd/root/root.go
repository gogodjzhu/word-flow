package root

import (
	"github.com/gogodjzhu/word-flow/pkg/cmd/dict"
	versioncmd "github.com/gogodjzhu/word-flow/pkg/cmd/version"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdRoot(f *cmdutil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "github.com/gogodjzhu/word-flow <command> <subcommand> [flags]",
		Short: "github.com/gogodjzhu/word-flow",
		Long:  `github.com/gogodjzhu/word-flow is a tool collection for bash environments.`,

		Annotations: map[string]string{
			"version": "0.0.1",
			"website": "www.github.com/gogodjzhu/word-flow.xyz",
		},
	}

	cmd.AddCommand(versioncmd.NewCmdVersion(f))

	if cmdDict, err := dict.NewCmdDict(f); err != nil {
		return nil, err
	} else {
		cmd.AddCommand(cmdDict)
	}

	if cmdNotebook, err := dict.NewCmdNotebook(f); err != nil {
		return nil, err
	} else {
		cmd.AddCommand(cmdNotebook)
	}

	return cmd, nil
}
