package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/oagudo/outbox"
	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
	"log"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
	"time"
)

type OutboxPublisher struct {
	publisher *rabbitmq.Publisher
}

func NewOutboxPublisher(conn *rabbitmq.Client) (*OutboxPublisher, error) {
	return &OutboxPublisher{
		publisher: conn.NewPublisher(),
	}, nil
}

func (p *OutboxPublisher) Publish(ctx context.Context, msg *outbox.Message) error {
	queueName := string(msg.Metadata)
	if queueName == "" {
		return fmt.Errorf("outbox publisher: queue name metadata is empty")
	}

	return p.publisher.PublishWithContext(ctx, msg.Payload, []string{queueName},
		gorabbit.WithPublishOptionsContentType("application/json"),
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
				outbox.WithInterval(15*time.Second),
				outbox.WithReadBatchSize(10),
			)

			reader.Start()

			go func() {
				for err := range reader.Errors() {
					switch e := err.(type) {
					case *outbox.PublishError:
						logrus.Printf("Failed to publish message | ID: %s | Error: %v",
							e.Message.ID, e.Err)

					case *outbox.UpdateError:
						logrus.Printf("Failed to update message | ID: %s | Error: %v",
							e.Message.ID, e.Err)

					case *outbox.DeleteError:
						logrus.Printf("Batch message deletion failed | Count: %d | Error: %v",
							len(e.Messages), e.Err)
						for _, msg := range e.Messages {
							log.Printf("Failed to delete message | ID: %s", msg.ID)
						}

					case *outbox.ReadError:
						logrus.Printf("Failed to read outbox messages | Error: %v", e.Err)

					default:
						logrus.Printf("Unexpected error occurred | Error: %v", e)
					}
				}
			}()

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
