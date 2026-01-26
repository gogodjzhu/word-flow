package cmdutil

import (
	"github.com/gogodjzhu/word-flow/internal/config"
	"io"
	"os"
	"runtime"
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
	In       io.Reader
	Out      io.Writer
	Renderer *Renderer
}

func ioStreams() *IOStreams {
	// 默认启用颜色渲染（除非环境明确不支持）
	colorEnabled := isTerminal(os.Stdout)

	return &IOStreams{
		In:       os.Stdin,
		Out:      os.Stdout,
		Renderer: NewRenderer(colorEnabled),
	}
}

// isTerminal 检查输出是否为终端
func isTerminal(w io.Writer) bool {
	// 简单检查：如果是 Windows，默认不启用颜色以避免兼容性问题
	// 在实际生产环境中，可以使用更复杂的终端检测
	return runtime.GOOS != "windows"
}
