package cmdutil

import (
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/gogodjzhu/word-flow/internal/config"
)

type Factory struct {
	IOStreams *IOStreams

	Config     func() (*config.Config, error)
	configOnce sync.Once
	cachedCfg  *config.Config
	configPath string
}

func NewFactory() *Factory {
	f := &Factory{
		IOStreams: ioStreams(),
		Config:    defaultConfigFunc,
	}
	return f
}

func defaultConfigFunc() (*config.Config, error) {
	return config.ReadConfig()
}

func (f *Factory) SetConfigPath(path string) {
	f.configPath = path
	f.Config = func() (*config.Config, error) {
		var cfg *config.Config
		var err error
		f.configOnce.Do(func() {
			cfg, err = config.ReadConfigSpecified(f.configPath)
			if err != nil {
				return
			}
			f.cachedCfg = cfg
		})
		if f.cachedCfg != nil {
			return f.cachedCfg, nil
		}
		return nil, err
	}
}

type IOStreams struct {
	In       io.Reader
	Out      io.Writer
	Renderer *Renderer
}

func ioStreams() *IOStreams {
	colorEnabled := isTerminal(os.Stdout)

	return &IOStreams{
		In:       os.Stdin,
		Out:      os.Stdout,
		Renderer: NewRenderer(colorEnabled),
	}
}

func isTerminal(w io.Writer) bool {
	return runtime.GOOS != "windows"
}