package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewClient),
		fx.Invoke(func(c *Client) {}),
	)
}

type Config struct {
	URL            string        `mapstructure:"url"`
	ReconnectDelay time.Duration `mapstructure:"reconnect_delay"`
}

type Client struct {
	Conn *rabbitmq.Conn
	cfg  *Config
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := rabbitmq.NewConn(
		cfg.URL,
		rabbitmq.WithConnectionOptionsReconnectInterval(cfg.ReconnectDelay),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn: conn,
		cfg:  cfg,
	}, nil
}

func (c *Client) withRawChannel(fn func(ch *amqp.Channel) error) error {
	conn, err := amqp.Dial(c.cfg.URL)
	if err != nil {
		return fmt.Errorf("dial raw amqp: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	return fn(ch)
}

func (c *Client) CreateExchange(exchangeName string, kind string) error {
	p, err := rabbitmq.NewPublisher(
		c.Conn,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeName(exchangeName),
		rabbitmq.WithPublisherOptionsExchangeKind(kind),
	)
	if err != nil {
		return fmt.Errorf("failed to init DLX/DLQ publisher: %w", err)
	}

	p.Close()
	return nil
}

func (c *Client) CreateQueue(name string, durable, autoDelete, exclusive bool) error {
	return c.withRawChannel(func(ch *amqp.Channel) error {
		_, err := ch.QueueDeclare(
			name,
			durable,
			autoDelete,
			exclusive,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("queue declare %q: %w", name, err)
		}
		return nil
	})
}

func (c *Client) BindQueue(queue, exchange, routingKey string) error {
	return c.withRawChannel(func(ch *amqp.Channel) error {
		if err := ch.QueueBind(
			queue,
			routingKey,
			exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("queue bind %q -> %q: %w", queue, exchange, err)
		}
		return nil
	})
}
