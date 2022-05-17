package nsqconsumer

import (
	"github.com/marmotedu/iam/pkg/app"
	"github.com/marmotedu/iam/pkg/log"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/config"
	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/options"
)

const commandDesc = `Consumers nsq messages and pubs to outputs.`

// NewApp creates a App object with default parameters.
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp("nsq consumer",
		basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}

func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		log.Init(opts.Log)
		defer log.Flush()
		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		return Run(cfg)
	}
}
