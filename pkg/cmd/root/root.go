package root

import (
	"fmt"

	"github.com/gogodjzhu/word-flow/internal/config"
	configcmd "github.com/gogodjzhu/word-flow/pkg/cmd/config"
	"github.com/gogodjzhu/word-flow/pkg/cmd/dict"
	"github.com/gogodjzhu/word-flow/pkg/cmd/server"
	"github.com/gogodjzhu/word-flow/pkg/cmd/trans"
	versioncmd "github.com/gogodjzhu/word-flow/pkg/cmd/version"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var version = "dev"

func NewCmdRoot(f *cmdutil.Factory) (*cobra.Command, error) {
	var configPath string

	cmd := &cobra.Command{
		Use:   "wordflow <command> <subcommand> [flags]",
		Short: "wordflow",
		Long:  `wordflow is a terminal-based dictionary and vocabulary learning tool.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if configPath != "" {
				f.SetConfigPath(configPath)
			}
			// Skip version validation for config subcommands so config init can regenerate
			for p := cmd; p != nil; p = p.Parent() {
				if p.Name() == "config" {
					return nil
				}
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			if cfg.Version != config.DefaultConfigVersion {
				return fmt.Errorf("unsupported config version: %q. Run 'wordflow config init' to regenerate your config", cfg.Version)
			}
			return nil
		},

		Annotations: map[string]string{
			"version": version,
			"website": "https://github.com/gogodjzhu/word-flow",
		},
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "Specify config file path")

	cmd.AddCommand(versioncmd.NewCmdVersion(f))
	cmd.AddCommand(configcmd.NewCmdConfig(f))

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