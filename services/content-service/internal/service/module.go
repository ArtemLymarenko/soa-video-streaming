package service

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewCategoryService,
			NewMediaContentService,
			NewRecommendations,
			NewS3Mock,
			NewBucketHandler,
		),
		fx.Invoke(RunSagaConsumers),
	)
}
