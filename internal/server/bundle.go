package server

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/stillunbuilt/mutefeed/internal/feed"
)

// FetchBundle retrieves and aggregates all feeds in a bundle concurrently.
func FetchBundle(ctx context.Context, bundleID string) (*gofeed.Feed, error) {
	return feed.FetchBundle(ctx, bundleID, nil)
}
