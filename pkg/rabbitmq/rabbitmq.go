package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewClient),
	)
}

type Config struct {
	URL               string        `mapstructure:"url"`
	ReconnectAttempts int           `mapstructure:"reconnect_attempts"`
	ReconnectDelay    time.Duration `mapstructure:"reconnect_delay"`
	PrefetchCount     int           `mapstructure:"prefetch_count"`
}

type Client struct {
	conn *amqp.Connection
	cfg  *Config
}

func NewClient(lc fx.Lifecycle, cfg *Config) (*Client, error) {
	c := &Client{
		cfg: cfg,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return c.connectWithRetry(ctx)
		},
		OnStop: func(ctx context.Context) error {
			if c.conn != nil {
				logrus.Info("Closing RabbitMQ connection")
				return c.conn.Close()
			}
			return nil
		},
	})

	return c, nil
}

func (c *Client) connectWithRetry(ctx context.Context) error {
	var lastErr error
	for i := 0; i <= c.cfg.ReconnectAttempts; i++ {
		conn, err := amqp.Dial(c.cfg.URL)
		if err == nil {
			c.conn = conn
			logrus.Info("Connected to RabbitMQ", zap.String("url", c.cfg.URL))
			return nil
		}

		lastErr = err
		logrus.WithFields(logrus.Fields{
			"url":     c.cfg.URL,
			"attempt": i + 1,
		}).WithError(err).Warn("Failed to connect to RabbitMQ, retrying")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.cfg.ReconnectDelay):
		}
	}

	return fmt.Errorf("rabbitmq: failed to connect after %d attempts: %w", c.cfg.ReconnectAttempts, lastErr)
}

func (c *Client) newChannel() (*amqp.Channel, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("rabbitmq: connection is nil")
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	if c.cfg.PrefetchCount > 0 {
		if err = ch.Qos(c.cfg.PrefetchCount, 0, false); err != nil {
			_ = ch.Close()
			return nil, err
		}
	}

	return ch, nil
}

type QueueOptions struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

func (c *Client) DeclareQueue(name string, opts QueueOptions) (amqp.Queue, *amqp.Channel, error) {
	ch, err := c.newChannel()
	if err != nil {
		return amqp.Queue{}, nil, err
	}

	q, err := ch.QueueDeclare(
		name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		opts.Args,
	)

	if err != nil {
		_ = ch.Close()
		return amqp.Queue{}, nil, err
	}

	return q, ch, nil
}

type Publisher struct {
	ch         *amqp.Channel
	exchange   string
	routingKey string
	mandatory  bool
	immediate  bool
}

func (c *Client) NewPublisher(exchange, routingKey string, mandatory, immediate bool) (*Publisher, error) {
	ch, err := c.newChannel()
	if err != nil {
		return nil, err
	}
	return &Publisher{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
		mandatory:  mandatory,
		immediate:  immediate,
	}, nil
}

func (p *Publisher) PublishJSON(ctx context.Context, body []byte) error {
	return p.Publish(ctx, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (p *Publisher) Publish(ctx context.Context, msg amqp.Publishing) error {
	return p.ch.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		p.mandatory,
		p.immediate,
		msg,
	)
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

type ConsumeOptions struct {
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

type ConsumerHandler func(ctx context.Context, d amqp.Delivery) error

func (c *Client) ConsumeQueue(
	ctx context.Context,
	queueName string,
	opts ConsumeOptions,
	handler ConsumerHandler,
) error {
	ch, err := c.newChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		queueName,
		opts.Consumer,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq: channel closed")
			}

			msgCtx, cancel := context.WithCancel(ctx)
			err := handler(msgCtx, d)
			cancel()

			if !opts.AutoAck {
				if err != nil {
					_ = d.Nack(false, true)
				} else {
					_ = d.Ack(false)
				}
			}
		}
	}
}
