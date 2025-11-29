package rabbitmq

import (
	"go.uber.org/fx"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewClient),
		fx.Invoke(func(c *Client) {}),
	)
}

type Config struct {
	URL               string        `mapstructure:"url"`
	ReconnectAttempts int           `mapstructure:"reconnect_attempts"`
	ReconnectDelay    time.Duration `mapstructure:"reconnect_delay"`
	PrefetchCount     int           `mapstructure:"prefetch_count"`
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

type Publisher struct {
	*rabbitmq.Publisher
}

func (c *Client) NewPublisher() *Publisher {
	publisher, err := rabbitmq.NewPublisher(c.Conn, rabbitmq.WithPublisherOptionsLogging)
	if err != nil {
		panic(err)
	}

	return &Publisher{
		Publisher: publisher,
	}
}

type Consumer struct {
	*rabbitmq.Consumer
}

func (c *Client) NewConsumer(queue string) *Consumer {
	consumer, err := rabbitmq.NewConsumer(
		c.Conn,
		queue,
		rabbitmq.WithConsumerOptionsQOSPrefetch(c.cfg.PrefetchCount),
		rabbitmq.WithConsumerOptionsLogging,
		rabbitmq.WithConsumerOptionsQueueDurable,
	)
	if err != nil {
		panic(err)
	}

	return &Consumer{
		Consumer: consumer,
	}
}
