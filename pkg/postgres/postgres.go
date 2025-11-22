package postgres

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewClient),
		fx.Provide(ProvidePGXPool),
		fx.Provide(
			fx.Annotate(
				ProvidePGXPool,
				fx.As(new(DB)),
			),
		),
		fx.Invoke(func(c *Client) {}),
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

func NewClient(lc fx.Lifecycle, cfg *Config) (*Client, error) {
	config, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnIdleTime = cfg.ConnMaxIdle
	config.HealthCheckPeriod = cfg.HealthCheckInt

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("pgx connect: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	logrus.Info("Connected to PostgreSQL")

	c := &Client{
		Pool: pool,
		cfg:  cfg,
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logrus.Info("Closing PostgreSQL connection")
			c.Pool.Close()
			return nil
		},
	})

	return c, nil
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
