package cache

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewCategoryCache,
			NewCategoryCollector,
		),
		fx.Invoke(RunCategoryCollector),
	)
}
