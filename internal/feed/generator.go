package feed

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

// GenerateRSS creates standard RSS 2.0 XML from the original feed metadata and filtered items
func GenerateRSS(original *gofeed.Feed, items []*gofeed.Item, feedSelfURL string) (string, error) {
	title := original.Title
	if title == "" {
		title = "Filtered Feed"
	} else {
		title = fmt.Sprintf("[Clean] %s", title)
	}

	feedLink := original.Link
	if feedLink == "" {
		feedLink = feedSelfURL
	}

	description := original.Description
	if description == "" {
		description = "Filtered by MuteFeed - AI slop, clickbait & paywall-free feed."
	}

	now := time.Now()
	outFeed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: feedLink},
		Description: description,
		Created:     now,
	}

	for _, item := range items {
		pubTime := now
		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			pubTime = *item.UpdatedParsed
		}

		authorName := ""
		if item.Author != nil {
			authorName = item.Author.Name
		}

		desc := item.Description
		if desc == "" {
			desc = item.Content
		}

		id := item.GUID
		if id == "" {
			id = item.Link
		}

		outFeed.Items = append(outFeed.Items, &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: item.Link},
			Description: desc,
			Author:      &feeds.Author{Name: authorName},
			Created:     pubTime,
			Id:          id,
		})
	}

	rssXml, err := outFeed.ToRss()
	if err != nil {
		return "", fmt.Errorf("failed to generate RSS xml: %w", err)
	}

	return rssXml, nil
}
