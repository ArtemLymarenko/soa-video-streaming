package postgres

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewUsersRepository,
			NewUserInfoRepository,
			NewUserPreference,
		),
	)
}
