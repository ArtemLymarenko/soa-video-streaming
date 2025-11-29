package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
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
	cfg *Config

	connMu sync.RWMutex
	conn   *amqp.Connection

	pubMu sync.Mutex
	pubCh *amqp.Channel

	stop chan struct{}
}

func NewClient(lc fx.Lifecycle, cfg *Config) (*Client, error) {
	c := &Client{
		cfg:  cfg,
		stop: make(chan struct{}),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := c.connectWithRetry(ctx); err != nil {
				return err
			}

			go c.runReconnectWatcher()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(c.stop)

			c.connMu.Lock()
			if c.conn != nil {
				return c.conn.Close()
			}
			c.connMu.Unlock()

			return nil
		},
	})

	return c, nil
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.cfg.URL)
	if err != nil {
		return err
	}

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	logrus.Infof("Connected to RabbitMQ: %s", c.cfg.URL)
	return nil
}

func (c *Client) connectWithRetry(ctx context.Context) error {
	var lastErr error
	for i := 0; i <= c.cfg.ReconnectAttempts; i++ {
		if err := c.connect(); err == nil {
			return nil
		} else {
			lastErr = err
			logrus.Warnf("RabbitMQ connection attempt %d failed: %v", i+1, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.cfg.ReconnectDelay):
		}
	}
	return fmt.Errorf("failed to connect after retries: %w", lastErr)
}

func (c *Client) runReconnectWatcher() {
	for {
		c.connMu.RLock()
		conn := c.conn
		c.connMu.RUnlock()

		if conn == nil {
			time.Sleep(c.cfg.ReconnectDelay)
			continue
		}

		err, ok := <-conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			return
		}

		logrus.Errorf("RabbitMQ connection closed: %v", err)

		for {
			select {
			case <-c.stop:
				return
			default:
			}

			if err := c.connect(); err == nil {
				logrus.Info("RabbitMQ reconnected")
				c.resetPublisherChannel()
				break
			}

			time.Sleep(c.cfg.ReconnectDelay)
		}
	}
}

func (c *Client) resetPublisherChannel() {
	c.pubMu.Lock()
	defer c.pubMu.Unlock()

	if c.pubCh != nil {
		_ = c.pubCh.Close()
		c.pubCh = nil
	}
}

func (c *Client) newChannel() (*amqp.Channel, error) {
	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return nil, fmt.Errorf("rabbitmq: no active connection")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if c.cfg.PrefetchCount > 0 {
		if err := ch.Qos(c.cfg.PrefetchCount, 0, false); err != nil {
			_ = ch.Close()
			return nil, err
		}
	}

	return ch, nil
}

func (c *Client) getPublisherChannel() (*amqp.Channel, error) {
	c.pubMu.Lock()
	defer c.pubMu.Unlock()

	if c.pubCh != nil && !c.pubCh.IsClosed() {
		return c.pubCh, nil
	}

	ch, err := c.newChannel()
	if err != nil {
		return nil, err
	}

	c.pubCh = ch
	return c.pubCh, nil
}

type PublishParams struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
}

func (c *Client) Publish(ctx context.Context, p PublishParams, msg amqp.Publishing) error {
	ch, err := c.getPublisherChannel()
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx, p.Exchange, p.RoutingKey, p.Mandatory, p.Immediate, msg)
	if err != nil {
		logrus.Warnf("Publish failed, resetting channel: %v", err)
		c.resetPublisherChannel()
	}

	return err
}

func (c *Client) Consume(ctx context.Context, queue string, handler func(ctx context.Context, d amqp.Delivery) error) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn := c.waitForConnection(ctx)
			if conn == nil {
				continue
			}

			ch, err := c.newChannel()
			if err != nil {
				time.Sleep(c.cfg.ReconnectDelay)
				continue
			}

			_, err = ch.QueueDeclarePassive(queue, true, false, false, false, nil)
			if err != nil {
				logrus.Errorf("consumer: queue '%s' not found: %v", queue, err)
				_ = ch.Close()
				time.Sleep(c.cfg.ReconnectDelay)
				continue // пробуємо знову
			}

			msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
			if err != nil {
				logrus.Errorf("consumer: consume error: %v", err)
				_ = ch.Close()
				time.Sleep(c.cfg.ReconnectDelay)
				continue
			}

			logrus.Infof("consumer: started on queue=%s", queue)

			for d := range msgs {
				if err := handler(ctx, d); err != nil {
					_ = d.Nack(false, true)
				} else {
					_ = d.Ack(false)
				}
			}

			logrus.Warn("consumer: msgs channel closed, reconnecting...")
			_ = ch.Close()
		}
	}()
}

func (c *Client) waitForConnection(ctx context.Context) *amqp.Connection {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		c.connMu.RLock()
		conn := c.conn
		c.connMu.RUnlock()

		if conn != nil && !conn.IsClosed() {
			return conn
		}

		time.Sleep(c.cfg.ReconnectDelay)
	}
}
