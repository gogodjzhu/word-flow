package root

import (
	"github.com/gogodjzhu/word-flow/pkg/cmd/dict"
	"github.com/gogodjzhu/word-flow/pkg/cmd/server"
	"github.com/gogodjzhu/word-flow/pkg/cmd/trans"
	versioncmd "github.com/gogodjzhu/word-flow/pkg/cmd/version"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var version = "0.4.4"

func NewCmdRoot(f *cmdutil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "wordflow <command> <subcommand> [flags]",
		Short: "wordflow",
		Long:  `wordflow is a terminal-based dictionary and vocabulary learning tool.`,

		Annotations: map[string]string{
			"version": version,
			"website": "https://github.com/gogodjzhu/word-flow",
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

	if cmdTrans, err := trans.NewCmdTrans(f); err != nil {
		return nil, err
	} else {
		cmd.AddCommand(cmdTrans)
	}

	if cmdServer, err := server.NewCmdServer(f); err != nil {
		return nil, err
	} else {
		cmd.AddCommand(cmdServer)
	}

	return cmd, nil
}
