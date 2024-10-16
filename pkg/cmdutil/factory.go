package cmdutil

import (
	"github.com/gogodjzhu/word-flow/internal/config"
	"io"
	"os"
)

type Factory struct {
	IOStreams *IOStreams

	Config func() (*config.Config, error)
}

func NewFactory() *Factory {
	f := &Factory{
		IOStreams: ioStreams(),
		Config:    config.ReadConfig,
	}
	return f
}

type IOStreams struct {
	In  io.Reader
	Out io.Writer
}

func ioStreams() *IOStreams {
	return &IOStreams{
		In:  os.Stdin,
		Out: os.Stdout,
	}
}
