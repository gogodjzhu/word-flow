package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <subcommand>",
		Short: "Manage configuration",
		Long:  "View, get, set, and initialize wordflow configuration.",
	}

	cmd.AddCommand(newCmdConfigView(f))
	cmd.AddCommand(newCmdConfigGet(f))
	cmd.AddCommand(newCmdConfigSet(f))
	cmd.AddCommand(newCmdConfigPath(f))
	cmd.AddCommand(newCmdConfigInit(f))

	return cmd
}

func newCmdConfigView(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			out, err := cfg.ToString()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(f.IOStreams.Out, out)
			return nil
		},
	}
}

func newCmdConfigGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			val, err := getConfigValue(cfg, args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(f.IOStreams.Out, val)
			return nil
		},
	}
}

func newCmdConfigSet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			if err := validateConfigKey(args[0]); err != nil {
				return err
			}
			return config.PatchYAMLFile(cfg.Common.ConfigFilename, args[0], args[1])
		},
	}
}

func validateConfigKey(key string) error {
	parts := strings.Split(key, ".")
	t := reflect.TypeOf(config.Config{})
	for i, part := range parts {
		field, found := findFieldTypeByYAMLTag(t, part)
		if !found {
			return fmt.Errorf("config key %q not found. Valid keys can be found with 'wordflow config view'", key)
		}
		if i == len(parts)-1 {
			return nil
		}
		if field.Type.Kind() == reflect.Ptr {
			field.Type = field.Type.Elem()
		}
		t = field.Type
	}
	return nil
}

func findFieldTypeByYAMLTag(t reflect.Type, tag string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		name := strings.SplitN(yamlTag, ",", 2)[0]
		if name == tag {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

func newCmdConfigPath(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(f.IOStreams.Out, cfg.Common.ConfigFilename)
			return nil
		},
	}
}

func newCmdConfigInit(f *cmdutil.Factory) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file with defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			configPath := cfg.Common.ConfigFilename
			if _, err := os.Stat(configPath); err == nil && !force {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "Config file already exists at %s. Use --force to overwrite.\n", configPath)
				return nil
			}
			dir := filepath.Dir(configPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}
			if err := os.WriteFile(configPath, []byte(config.ConfigTemplate()), 0644); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Config file created at %s\n", configPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config file")
	return cmd
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	val := reflect.ValueOf(cfg).Elem()
	parts := strings.Split(key, ".")
	current := val

	for i, part := range parts {
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return "", fmt.Errorf("config key %q not found (nil pointer at %q)", key, strings.Join(parts[:i+1], "."))
			}
			current = current.Elem()
		}
		if current.Kind() != reflect.Struct {
			return "", fmt.Errorf("config key %q not found", key)
		}
		field, found := findFieldByYAMLTag(current, part)
		if !found {
			return "", fmt.Errorf("config key %q not found", key)
		}
		current = field
	}

	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return "", nil
		}
		current = current.Elem()
	}

	switch v := current.Interface().(type) {
	case config.Duration:
		return time.Duration(v).String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func findFieldByYAMLTag(v reflect.Value, tag string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		name := strings.SplitN(yamlTag, ",", 2)[0]
		if name == tag {
			return field, true
		}
	}
	return reflect.Value{}, false
}