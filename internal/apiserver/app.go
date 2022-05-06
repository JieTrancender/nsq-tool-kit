package apiserver

import (
	"github.com/JieTrancender/kbm-iam/internal/apiserver/options"
	"github.com/JieTrancender/kbm-iam/pkg/app"
)

const commandDesc = `The API Server services REST operations to do the api
objects management.`

// NewApp creates a App object with default parameters.
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp("IAM API Server",
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
		return Run()
	}
}
