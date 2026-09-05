package news

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const maxRSSResponseBytes = 4 << 20

// RSSNewsProvider reads source-neutral NewsItems from configured RSS 2.0 feeds.
type RSSNewsProvider struct {
	feedURLs []string
	client   *http.Client
}

// RSSFetchError reports how many configured feeds failed while allowing items
// from successful feeds to continue through the ingestion pipeline.
type RSSFetchError struct {
	failures []error
}

func (e *RSSFetchError) Error() string {
	return fmt.Sprintf("%d RSS feed(s) failed: %v", len(e.failures), errors.Join(e.failures...))
}

func (e *RSSFetchError) Unwrap() []error {
	return e.failures
}

func (e *RSSFetchError) FailedFeedCount() int {
	if e == nil {
		return 0
	}
	return len(e.failures)
}

func NewRSSNewsProvider(feedURLs []string, client *http.Client) *RSSNewsProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	feeds := make([]string, 0, len(feedURLs))
	for _, feedURL := range feedURLs {
		if feedURL = strings.TrimSpace(feedURL); feedURL != "" {
			feeds = append(feeds, feedURL)
		}
	}
	return &RSSNewsProvider{feedURLs: feeds, client: client}
}

func (p *RSSNewsProvider) Fetch(ctx context.Context, sinceTimestamp int64) ([]NewsItem, error) {
	if p == nil || len(p.feedURLs) == 0 {
		return nil, errors.New("at least one RSS feed URL is required")
	}

	items := make([]NewsItem, 0)
	seen := make(map[string]struct{})
	var fetchErrors []error
	for _, feedURL := range p.feedURLs {
		feedItems, err := p.fetchFeed(ctx, feedURL, sinceTimestamp)
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Errorf("fetch %s: %w", feedURL, err))
			continue
		}
		for _, item := range feedItems {
			if _, duplicate := seen[item.ID]; duplicate {
				continue
			}
			seen[item.ID] = struct{}{}
			items = append(items, item)
		}
	}
	if len(fetchErrors) > 0 {
		return items, &RSSFetchError{failures: fetchErrors}
	}
	return items, nil
}

func (p *RSSNewsProvider) fetchFeed(ctx context.Context, feedURL string, sinceTimestamp int64) ([]NewsItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rss+xml, application/xml;q=0.9, text/xml;q=0.8")
	req.Header.Set("User-Agent", "Crypto-Strategy-Lab/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxRSSResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read RSS: %w", err)
	}
	if len(data) > maxRSSResponseBytes {
		return nil, fmt.Errorf("RSS response exceeds %d bytes", maxRSSResponseBytes)
	}

	var document rssDocument
	if err := xml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse RSS: %w", err)
	}
	if strings.TrimSpace(document.Channel.Title) == "" && len(document.Channel.Items) == 0 {
		return nil, errors.New("RSS channel is empty")
	}

	items := make([]NewsItem, 0, len(document.Channel.Items))
	for _, entry := range document.Channel.Items {
		item, ok := mapRSSItem(feedURL, document.Channel.Title, entry)
		if !ok || item.PublishedAt < sinceTimestamp {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	GUID        string   `xml:"guid"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Content     string   `xml:"encoded"`
	Link        string   `xml:"link"`
	PublishedAt string   `xml:"pubDate"`
	Source      string   `xml:"source"`
	Categories  []string `xml:"category"`
}

func mapRSSItem(feedURL, channelTitle string, entry rssItem) (NewsItem, bool) {
	title := cleanRSSValue(entry.Title)
	text := cleanRSSValue(firstNonEmpty(entry.Content, entry.Description, entry.Title))
	publishedAt, err := parseRSSDate(entry.PublishedAt)
	if err != nil || title == "" || text == "" {
		return NewsItem{}, false
	}

	articleURL := normalizeArticleURL(feedURL, entry.Link)
	stableKey := articleURL
	if stableKey == "" {
		if guid := cleanRSSValue(entry.GUID); guid != "" {
			stableKey = feedURL + "|guid|" + guid
		}
	}
	if stableKey == "" {
		stableKey = feedURL + "|item|" + title + "|" + publishedAt.UTC().Format(time.RFC3339Nano)
	}

	source := cleanRSSValue(firstNonEmpty(entry.Source, channelTitle))
	if source == "" {
		if parsedFeedURL, parseErr := url.Parse(feedURL); parseErr == nil {
			source = parsedFeedURL.Hostname()
		}
	}

	return NewsItem{
		ID:           stableRSSID(stableKey),
		Title:        title,
		Text:         text,
		Source:       source,
		URL:          articleURL,
		PublishedAt:  publishedAt.UnixMilli(),
		RelatedCoins: extractRelatedCoins(append([]string{title, text}, entry.Categories...)...),
	}, true
}

func extractRelatedCoins(values ...string) []string {
	words := make(map[string]struct{})
	for _, value := range values {
		for _, word := range strings.FieldsFunc(value, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}) {
			words[strings.ToUpper(word)] = struct{}{}
		}
	}

	related := make([]string, 0, 3)
	for _, coin := range []struct {
		symbol  string
		aliases []string
	}{
		{symbol: "BTC", aliases: []string{"BTC", "BITCOIN"}},
		{symbol: "ETH", aliases: []string{"ETH", "ETHEREUM"}},
		{symbol: "SOL", aliases: []string{"SOL", "SOLANA"}},
	} {
		for _, alias := range coin.aliases {
			if _, found := words[alias]; found {
				related = append(related, coin.symbol)
				break
			}
		}
	}
	return related
}

func parseRSSDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339Nano,
	} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported RSS date %q", raw)
}

func normalizeArticleURL(feedURL, articleURL string) string {
	raw := strings.TrimSpace(articleURL)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if !parsed.IsAbs() {
		base, baseErr := url.Parse(feedURL)
		if baseErr != nil {
			return ""
		}
		parsed = base.ResolveReference(parsed)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	parsed.Fragment = ""
	return parsed.String()
}

func stableRSSID(key string) string {
	digest := sha256.Sum256([]byte(key))
	return "rss-" + hex.EncodeToString(digest[:16])
}

func cleanRSSValue(value string) string {
	return strings.TrimSpace(html.UnescapeString(value))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var _ NewsProvider = (*RSSNewsProvider)(nil)
