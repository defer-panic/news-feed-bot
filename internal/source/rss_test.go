package source_test

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/defer-panic/news-feed-bot/internal/model"
	"github.com/defer-panic/news-feed-bot/internal/source"
)

//go:embed testdata/feed.xml
var feed []byte

func TestRSSSource_Fetch(t *testing.T) {
	var (
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/xml; charset=utf-8")
			_, _ = w.Write(feed)
		}))
		src = &source.RSSSource{
			URL:        ts.URL,
			SourceName: "dev.to",
		}
		expected = []model.Item{
			{
				Title:      "Climbing Stairs LeetCode 70",
				Categories: []string{"leetcode", "go", "programming", "algorithms"},
				Link:       "https://dev.to/digitebs/climbing-stairs-leetcode-70-4n1j",
				Date:       parseDate(t, "Sun, 19 Mar 2023 07:04:42 +0000"),
				SourceName: "dev.to",
			},
			{
				Title:      "Find the Duplicate Number",
				Categories: []string{"go", "algorithms", "programming", "leetcode"},
				Link:       "https://dev.to/digitebs/find-the-duplicate-number-1l0",
				Date:       parseDate(t, "Sat, 18 Mar 2023 23:34:32 +0000"),
				SourceName: "dev.to",
			},
			{
				Title:      "My Favorite Free Courses to Learn Golang in 2023",
				Categories: []string{"programming", "coding", "go", "development"},
				Link:       "https://dev.to/javinpaul/my-favorite-free-courses-to-learn-golang-in-2023-3mh6",
				Date:       parseDate(t, "Sat, 18 Mar 2023 07:22:06 +0000"),
				SourceName: "dev.to",
			},
		}
	)

	items, err := src.Fetch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, items)
}

func parseDate(t *testing.T, dateStr string) time.Time {
	date, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", dateStr)
	require.NoError(t, err)

	return date
}
