package api

import (
	"context"

	"github.com/wetor/AnimeGo/internal/models"
)

type Filter interface {
	Filter([]*models.FeedItem) []*models.FeedItem
}

type FilterManager interface {
	Update(ctx context.Context, items []*models.FeedItem)
}
