package feeder

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/p1ass/feeds"
	"github.com/pkg/errors"
)

type atomCrawler struct {
	URL string
}

// NewAtomCrawler returns atomCrawler
func NewAtomCrawler(url string) Crawler {
	return &atomCrawler{URL: url}
}

// Crawl is crawl entry items from atom file
func (crawler *atomCrawler) Crawl() ([]*Item, error) {
	resp, err := http.Get(crawler.URL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get response from rss.")
	}
	defer resp.Body.Close()

	var atom feeds.AtomFeed
	err = xml.NewDecoder(resp.Body).Decode(&atom)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode response body.")
	}

	items := []*Item{}

	for _, e := range atom.Entries {
		item, err := convertAtomEntryToItem(e)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to convert RSSItem to Item.")
		}
		items = append(items, item)
	}
	return items, nil
}

func convertAtomEntryToItem(e *feeds.AtomEntry) (*Item, error) {
	p, err := time.Parse(time.RFC3339, e.Published)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to parse published time. published=%v", e.Published))
	}
	u, err := time.Parse(time.RFC3339, e.Updated)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to parse updated time. updated=%v", e.Updated))
	}
	var description string
	if e.Summary != nil{
		description = e.Summary.Content
	}
	i := &Item{
		Title:       e.Title,
		Description: description,
		ID:          e.Id,
		Created:     &p,
		Updated:     &u,
	}

	var name, email string
	if e.Author != nil {
		name, email = e.Author.Name, e.Author.Email
	}
	if len(name) > 0 || len(email) > 0 {
		i.Author = &Author{
			Name:  e.Author.Name,
			Email: e.Author.Email,
		}
	}

	if e.Content != nil {
		i.Content = e.Content.Content
	}

	for _, link := range e.Links {
		if link.Rel == "enclosure" {
			i.Enclosure = &Enclosure{
				URL:    link.Href,
				Length: link.Length,
				Type:   link.Type,
			}
		} else {
			i.Link = &Link{
				Href:   link.Href,
				Rel:    link.Rel,
				Type:   link.Type,
				Length: link.Length,
			}
		}
	}
	return i, nil
}
