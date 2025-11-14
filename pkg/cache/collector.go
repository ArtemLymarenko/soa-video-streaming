package cache

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CollectorCache[K comparable, V any] struct {
	items   map[K]V
	itemsMX sync.RWMutex

	prev int64
	next int64
}

func NewCollectorCache[K comparable, V any]() *CollectorCache[K, V] {
	return &CollectorCache[K, V]{
		items: make(map[K]V),
	}
}

func (c *CollectorCache[K, V]) Set(key K, value V) {
	c.itemsMX.Lock()
	defer c.itemsMX.Unlock()
	c.items[key] = value
}

func (c *CollectorCache[K, V]) SetAll(items map[K]V) {
	c.itemsMX.Lock()
	defer c.itemsMX.Unlock()

	for key, value := range items {
		c.items[key] = value
	}
}

func (c *CollectorCache[K, V]) Get(key K) (V, bool) {
	c.itemsMX.RLock()
	defer c.itemsMX.RUnlock()
	t, ok := c.items[key]
	return t, ok
}

type CollectorFunc[K comparable, V any] func(prev, next int64) (map[K]V, error)

func (c *CollectorCache[K, V]) RunCollector(ctx context.Context, collectorFunc CollectorFunc[K, V]) {
	logrus.Infof("Running cache collector")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := c.collect(collectorFunc)
		if err != nil {
			logrus.WithError(err).Errorf("Failed collec cache items")
		}

		time.Sleep(30 * time.Second)
	}
}

func (c *CollectorCache[K, V]) collect(collectorFunc CollectorFunc[K, V]) error {
	c.next = time.Now().Unix()

	items, err := collectorFunc(c.prev, c.next)
	if err != nil {
		return err
	}

	c.prev = c.next

	c.SetAll(items)

	return nil
}
