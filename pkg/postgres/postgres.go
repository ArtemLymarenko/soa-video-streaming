package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewClient),
		fx.Provide(
			fx.Annotate(
				ProvidePGXPool,
				fx.As(new(DB)),
			),
		),
		fx.Invoke(Run),
	)
}

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Config struct {
	DSN            string        `mapstructure:"dsn"`
	MaxConns       int32         `mapstructure:"max_conns"`
	MinConns       int32         `mapstructure:"min_conns"`
	ConnMaxIdle    time.Duration `mapstructure:"conn_max_idle"`
	HealthCheckInt time.Duration `mapstructure:"health_check_int"`
}

type Client struct {
	Pool *pgxpool.Pool
	cfg  *Config
}

func NewClient(cfg *Config) (*Client, error) {
	return &Client{
		cfg: cfg,
	}, nil
}

func Run(lc fx.Lifecycle, c *Client) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			config, err := pgxpool.ParseConfig(c.cfg.DSN)
			if err != nil {
				return fmt.Errorf("parse DSN: %w", err)
			}

			config.MaxConns = c.cfg.MaxConns
			config.MinConns = c.cfg.MinConns
			config.MaxConnIdleTime = c.cfg.ConnMaxIdle
			config.HealthCheckPeriod = c.cfg.HealthCheckInt

			pool, err := pgxpool.NewWithConfig(ctx, config)
			if err != nil {
				return fmt.Errorf("pgx connect: %w", err)
			}

			c.Pool = pool

			logrus.Info("Connected to PostgreSQL")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("Closing PostgreSQL connection")
			c.Pool.Close()
			return nil
		},
	})
}

func ProvidePGXPool(c *Client) *pgxpool.Pool {
	return c.Pool
}

func (c *Client) Tx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
