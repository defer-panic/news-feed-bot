package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/samber/lo"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type OsnovaSource struct {
	SourceName string
	URL        string

	apiToken string
}

func NewOsnovaSource(sourceName, url, apiToken string) OsnovaSource {
	return OsnovaSource{
		SourceName: sourceName,
		URL:        url,
		apiToken:   apiToken,
	}
}

func (o OsnovaSource) Fetch(ctx context.Context) ([]model.Item, error) {
	//TODO implement me
	const urlFormat = "%s/v1.9/news/default/recent"

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(urlFormat, o.URL),
		nil,
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiToken))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyJSON, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var feed osnovaFeedResponse
	if err := json.Unmarshal(bodyJSON, &feed); err != nil {
		return nil, err
	}

	return lo.Map(feed.Result, func(item osnovaFeedItem, index int) model.Item {
		return model.Item{
			Title:      item.Title,
			Link:       item.URL,
			Date:       time.Unix(item.Date, 0),
			SourceName: o.SourceName,
		}
	}), nil
}

func (o OsnovaSource) Name() string {
	return o.SourceName
}

type osnovaFeedResponse struct {
	Result []osnovaFeedItem `json:"result"`
}

type osnovaFeedItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Date  int64  `json:"date"`
	URL   string `json:"url"`
}
