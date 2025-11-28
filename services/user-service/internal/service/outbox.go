package service

import (
	"context"
	"database/sql"
	"fmt"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/oagudo/outbox"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type OutboxPublisher struct {
	client *rabbitmq.Client
	ch     *amqp.Channel
}

func NewOutboxPublisher(lc fx.Lifecycle, client *rabbitmq.Client) (*OutboxPublisher, error) {
	p := &OutboxPublisher{
		client: client,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ch, err := client.NewChannel()
			if err != nil {
				return err
			}

			p.ch = ch
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if p.ch != nil {
				return p.ch.Close()
			}
			return nil
		},
	})

	return p, nil
}

func (p *OutboxPublisher) Publish(ctx context.Context, msg *outbox.Message) error {
	if p.ch == nil {
		return fmt.Errorf("outbox publisher: channel is nil")
	}

	queueName := string(msg.Metadata)
	if queueName == "" {
		return fmt.Errorf("outbox publisher: queue name (metadata) is empty")
	}

	return p.ch.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg.Payload,
		},
	)
}

func RunOutboxReader(lc fx.Lifecycle, pool *postgres.Client, publisher *OutboxPublisher) {
	var db *sql.DB
	var reader *outbox.Reader

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			db = stdlib.OpenDBFromPool(pool.Pool)

			dbCtx := outbox.NewDBContext(db, outbox.SQLDialectPostgres)

			reader = outbox.NewReader(
				dbCtx,
				publisher,
				outbox.WithInterval(1*time.Second),
				outbox.WithReadBatchSize(10),
			)

			reader.Start()
			logrus.Info("Outbox reader started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if reader != nil {
				return reader.Stop(ctx)
			}

			if db != nil {
				return db.Close()
			}

			return nil
		},
	})
}
