package news

import (
	"strings"
	"testing"
)

func TestNewsUpsertStatementBuildsBatchAndEncodesCoins(t *testing.T) {
	items := []NewsItem{
		{ID: "news-1", Title: "Bitcoin update", Text: "Content", Source: "Wire", URL: "https://example.com/1", PublishedAt: 100, RelatedCoins: []string{"BTC"}},
		{ID: "news-2", Title: "Market update", Text: "More content", Source: "Wire", PublishedAt: 200},
	}

	query, args, err := newsUpsertStatement(items, 300)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"($1,$2,$3,$4,$5,$6,$7,$8)",
		"($9,$10,$11,$12,$13,$14,$15,$16)",
		"ON CONFLICT (id) DO UPDATE",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query missing %q: %s", fragment, query)
		}
	}
	if len(args) != 16 {
		t.Fatalf("args=%d, want 16", len(args))
	}
	if args[6] != `["BTC"]` || args[14] != `[]` {
		t.Fatalf("related coin JSON = %v / %v", args[6], args[14])
	}
	if args[7] != int64(300) || args[15] != int64(300) {
		t.Fatalf("created_at args = %v / %v", args[7], args[15])
	}
}
