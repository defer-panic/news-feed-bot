package model

import (
	"encoding/json"
	"time"

	"github.com/tomakado/containers/set"
)

type Item struct {
	Title       string
	Description string
	Link        string
	Date        time.Time
	SourceName  string
}

type FetchInput struct {
	Since time.Time
}

type ClassifiedItem struct {
	ID    int64
	Item  Item
	Slug  string
	Score int
}

type Class struct {
	Name  string  `json:"name"`
	Slug  string  `json:"slug"`
	Icon  string  `json:"icon"`
	Stems StemSet `json:"stems"`
}

type ClassifyResult struct {
	Class *Class
	Score int
}

type StemSet struct {
	Stems set.HashSet[string]
}

func (s *StemSet) UnmarshalJSON(data []byte) error {
	var stems []string
	if err := json.Unmarshal(data, &stems); err != nil {
		return err
	}

	*s = StemSet{Stems: set.New(stems...)}

	return nil
}

type Chat struct {
	ID         int64
	TelegramID int64
	Topics     set.HashSet[string]
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
