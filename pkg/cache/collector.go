package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const ResyncGapSeconds = 300

type CollectorCache[K comparable, V any] struct {
	items  sync.Map
	lastTS int64
	name   string
}

func NewCollectorCache[K comparable, V any](name string) *CollectorCache[K, V] {
	return &CollectorCache[K, V]{
		name: name,
	}
}

func (c *CollectorCache[K, V]) Set(key K, value V) {
	c.items.Store(key, value)
}

func (c *CollectorCache[K, V]) SetAll(items map[K]V) {
	for key, value := range items {
		c.items.Store(key, value)
	}
}

func (c *CollectorCache[K, V]) Get(key K) (V, bool) {
	value, ok := c.items.Load(key)
	if !ok {
		var zeroValue V
		return zeroValue, false
	}

	return value.(V), true
}

func (c *CollectorCache[K, V]) ForEach(process func(k K, v V)) {
	c.items.Range(func(k, v any) bool {
		key, okKey := k.(K)
		val, okVal := v.(V)
		if okKey && okVal {
			process(key, val)
		}

		return true
	})
}

type CollectorFunc[K comparable, V any] func(prev, next int64) (map[K]V, error)
type MaxTSFunc func() (int64, error)

func (c *CollectorCache[K, V]) RunCollector(
	ctx context.Context,
	collectorFunc CollectorFunc[K, V],
	maxTSFunc MaxTSFunc,
	postCollectionFuncs ...func() error,
) {
	logrus.Infof("Starting cache collector (%s)", c.name)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.runCollect(collectorFunc, maxTSFunc, postCollectionFuncs...)
			if err != nil {
				logrus.WithField("entity", c.name).WithError(err).Error("Failed to collect items")
			}
		}
	}
}

func (c *CollectorCache[K, V]) runCollect(
	collectorFunc CollectorFunc[K, V],
	maxTSFunc MaxTSFunc,
	postCollectionFuncs ...func() error,
) error {
	lastTs := c.lastTS
	maxTs, err := maxTSFunc()
	if err != nil {
		return fmt.Errorf("failed to get latest timestamp: %w", err)
	}

	lastTsWithGap := lastTs
	if lastTs > ResyncGapSeconds && time.Now().Unix()-lastTs < ResyncGapSeconds {
		lastTsWithGap -= ResyncGapSeconds
	}

	if maxTs == lastTsWithGap {
		return fmt.Errorf("no change in cache. Revision: %d: err: %v", lastTs, err)
	}

	start := time.Now()
	items, err := collectorFunc(lastTsWithGap, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("error collecting cache items: %v", err)
	}

	logrus.WithField("entity", c.name).Infof(
		"Rows: %d. Took: %s, revision: %d / %d. Now: %v.",
		len(items), time.Since(start), lastTs, maxTs, time.Now().Unix(),
	)

	c.SetAll(items)

	now := time.Now().Unix()
	if maxTs > now {
		maxTs = now
	}

	c.lastTS = maxTs

	for _, postFunc := range postCollectionFuncs {
		err = postFunc()
		if err != nil {
			logrus.WithField("entity", c.name).WithError(err).Errorf("Error post collection function failed")
		}
	}

	return nil
}
