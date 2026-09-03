package news_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/news"
)

func TestRSSNewsProviderMapsAndFiltersItems(t *testing.T) {
	published := time.Date(2026, time.September, 2, 10, 30, 0, 0, time.UTC)
	server := rssServer(t, http.StatusOK, `<?xml version="1.0"?>
<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>Crypto Test Wire</title>
    <item>
      <guid>article-1</guid>
      <title>Bitcoin &amp; markets rise</title>
      <content:encoded><![CDATA[Markets react positively to new demand.]]></content:encoded>
      <link>/article-1#comments</link>
      <pubDate>Wed, 02 Sep 2026 10:30:00 +0000</pubDate>
      <source>Test Desk</source>
    </item>
    <item>
      <guid>old-article</guid>
      <title>Old article</title>
      <description>Too old for this fetch.</description>
      <link>/old</link>
      <pubDate>Tue, 01 Sep 2026 08:00:00 +0000</pubDate>
    </item>
    <item>
      <guid>malformed</guid>
      <title>Missing date</title>
      <description>This item must be skipped.</description>
    </item>
  </channel>
</rss>`)
	defer server.Close()

	provider := news.NewRSSNewsProvider([]string{server.URL}, server.Client())
	items, err := provider.Fetch(context.Background(), published.Add(-time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d, want 1", len(items))
	}
	item := items[0]
	if item.Title != "Bitcoin & markets rise" || item.Text != "Markets react positively to new demand." {
		t.Fatalf("mapped text = %q / %q", item.Title, item.Text)
	}
	if item.Source != "Test Desk" || item.URL != server.URL+"/article-1" || item.PublishedAt != published.UnixMilli() {
		t.Fatalf("mapped metadata = %+v", item)
	}
	if !strings.HasPrefix(item.ID, "rss-") {
		t.Fatalf("ID=%q, want stable rss prefix", item.ID)
	}

	again, err := provider.Fetch(context.Background(), published.Add(-time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 1 || again[0].ID != item.ID {
		t.Fatalf("stable ID changed: %q -> %+v", item.ID, again)
	}
}

func TestRSSNewsProviderDeduplicatesArticleURLs(t *testing.T) {
	server := rssServer(t, http.StatusOK, `
<rss version="2.0"><channel><title>Duplicate Wire</title>
  <item><guid>one</guid><title>Same story</title><description>First copy.</description><link>/same</link><pubDate>Wed, 02 Sep 2026 10:30:00 +0000</pubDate></item>
  <item><guid>two</guid><title>Same story updated</title><description>Second copy.</description><link>/same#latest</link><pubDate>Wed, 02 Sep 2026 10:31:00 +0000</pubDate></item>
</channel></rss>`)
	defer server.Close()

	provider := news.NewRSSNewsProvider([]string{server.URL}, server.Client())
	items, err := provider.Fetch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d, want duplicate URL collapsed to 1", len(items))
	}
}

func TestRSSNewsProviderReturnsValidItemsWhenAnotherFeedFails(t *testing.T) {
	valid := rssServer(t, http.StatusOK, `
<rss version="2.0"><channel><title>Working Wire</title>
  <item><guid>one</guid><title>Working story</title><description>Usable description.</description><link>/one</link><pubDate>Wed, 02 Sep 2026 10:30:00 +0000</pubDate></item>
</channel></rss>`)
	defer valid.Close()
	failing := rssServer(t, http.StatusServiceUnavailable, "temporarily unavailable")
	defer failing.Close()

	provider := news.NewRSSNewsProvider([]string{failing.URL, valid.URL}, valid.Client())
	items, err := provider.Fetch(context.Background(), 0)
	if err == nil {
		t.Fatal("expected partial fetch error")
	}
	var fetchErr *news.RSSFetchError
	if !errors.As(err, &fetchErr) || fetchErr.FailedFeedCount() != 1 {
		t.Fatalf("error=%v, want one failed RSS feed", err)
	}
	if len(items) != 1 || items[0].Title != "Working story" {
		t.Fatalf("items=%+v, want one item from working feed", items)
	}
}

func TestRSSNewsProviderRejectsInvalidRSS(t *testing.T) {
	server := rssServer(t, http.StatusOK, `<feed><entry>not RSS 2.0</entry></feed>`)
	defer server.Close()

	provider := news.NewRSSNewsProvider([]string{server.URL}, server.Client())
	items, err := provider.Fetch(context.Background(), 0)
	if err == nil {
		t.Fatal("expected invalid RSS error")
	}
	if len(items) != 0 {
		t.Fatalf("items=%d, want 0", len(items))
	}
}

func TestRSSNewsProviderReportsNetworkError(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection reset")
	})}
	provider := news.NewRSSNewsProvider([]string{"https://feed.example/rss"}, client)

	items, err := provider.Fetch(context.Background(), 0)
	if err == nil || !strings.Contains(err.Error(), "connection reset") {
		t.Fatalf("error=%v, want deterministic network error", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d, want 0", len(items))
	}
}

func rssServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
