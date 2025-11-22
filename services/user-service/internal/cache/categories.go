package cache

import (
	"context"
	"soa-video-streaming/pkg/cache"
	contentpb "soa-video-streaming/pkg/pb/content"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type CategoryCache = cache.CollectorCache[string, struct{}]

func NewCategoryCache() *CategoryCache {
	return cache.NewCollectorCache[string, struct{}]("categories")
}

func NewCategoryCollector(client contentpb.CategoryServiceClient) cache.CollectorFunc[string, struct{}] {
	return func(prev, next int64) (map[string]struct{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &contentpb.GetCategoriesByTimestampRequest{
			From: prev,
			To:   time.Now().Unix(),
		}

		resp, err := client.GetCategoriesByTimestamp(ctx, req)
		if err != nil {
			return nil, err
		}

		res := make(map[string]struct{}, len(resp.GetCategories()))
		for _, c := range resp.GetCategories() {
			res[c.GetId()] = struct{}{}
		}

		return res, nil
	}
}

func RunCategoryCollector(
	lc fx.Lifecycle,
	cache *CategoryCache,
	collector cache.CollectorFunc[string, struct{}],
	client contentpb.CategoryServiceClient,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go cache.RunCollector(ctx, collector, func() (int64, error) {
				maxTs, err := client.GetMaxTimestamp(ctx, &contentpb.GetMaxTimestampRequest{})
				return maxTs.GetMaxTimestamp(), err
			})

			logrus.Info("Categories cache collector started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("Categories cache collector stopped")
			return nil
		},
	})
}
