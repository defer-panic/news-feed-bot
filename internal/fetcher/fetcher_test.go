package fetcher_test

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/defer-panic/news-feed-bot/internal/fetcher"
	"github.com/defer-panic/news-feed-bot/internal/fetcher/mocks"
	"github.com/defer-panic/news-feed-bot/internal/model"
)

//go:embed testdata/feed1.xml
var feed1 []byte

//go:embed testdata/feed2.xml
var feed2 []byte

func TestFetcher_Fetch(t *testing.T) {
	var (
		source1Server   = setupFeedSever(feed1)
		source2Server   = setupFeedSever(feed2)
		sourcesProvider = &mocks.SourcesProviderMock{
			SourcesFunc: func(ctx context.Context) ([]model.Source, error) {
				return []model.Source{
					{
						ID:       1,
						Name:     "dev.to",
						FeedURL:  source1Server.URL,
						Priority: 10,
					},
					{
						ID:       2,
						Name:     "Go Time Podcast",
						FeedURL:  source2Server.URL,
						Priority: 100,
					},
				}, nil
			},
		}
	)

	t.Run("should fetch articles from all sources", func(t *testing.T) {
		var (
			articles       = make(map[string]model.Article)
			articleStorage = &mocks.ArticleStorageMock{
				StoreFunc: func(ctx context.Context, article model.Article) error {
					articles[article.Link] = article
					return nil
				},
			}
			fetcher = fetcher.New(articleStorage, sourcesProvider, 0, nil)
		)

		require.NoError(t, fetcher.Fetch(context.Background()))
		assert.Len(t, articles, 4)
	})

	t.Run("should filter articles by keywords", func(t *testing.T) {
		var (
			articles       = make(map[string]model.Article)
			articleStorage = &mocks.ArticleStorageMock{
				StoreFunc: func(ctx context.Context, article model.Article) error {
					articles[article.Link] = article
					return nil
				},
			}
			filterKeywords = []string{"leetcode"}
			fetcher        = fetcher.New(articleStorage, sourcesProvider, 0, filterKeywords)
		)

		require.NoError(t, fetcher.Fetch(context.Background()))
		assert.Len(t, articles, 3)
	})
}

func setupFeedSever(feed []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/xml; charset=utf-8")
		_, _ = w.Write(feed)
	}))
}
