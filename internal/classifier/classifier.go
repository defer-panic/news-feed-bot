package classifier

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/kljensen/snowball"
	"github.com/tomakado/containers/set"

	"github.com/defer-panic/news-feed-bot/internal/model"
)

type Classifier struct {
	classes map[string]*model.Class
}

func New(classes []*model.Class) *Classifier {
	classMap := make(map[string]*model.Class, len(classes))
	for _, class := range classes {
		classMap[class.Slug] = class
	}

	return &Classifier{classes: classMap}
}

func (c *Classifier) Classify(text string) (*model.ClassifyResult, error) {
	stems, err := c.getStems(text)
	if err != nil {
		return nil, err
	}

	scores := c.initScoreMap()

	c.collectScores(scores, stems)

	return c.findHighestScoreClass(scores), nil
}

func (c *Classifier) Classes(_ context.Context) ([]*model.Class, error) {
	classes := make([]*model.Class, 0, len(c.classes))
	for _, class := range c.classes {
		classes = append(classes, class)
	}

	return classes, nil
}

func (c *Classifier) getStems(text string) ([]string, error) {
	var (
		tokens = strings.Split(text, " ")
		stems  = make([]string, 0, len(tokens))
	)

	for _, token := range tokens {
		if strings.TrimSpace(token) == "" {
			continue
		}

		stemmed, err := snowball.Stem(token, "russian", true)
		if err != nil {
			return nil, err
		}

		stems = append(stems, stemmed)
	}

	return stems, nil
}

func (c *Classifier) initScoreMap() map[string]classScore {
	scores := make(map[string]classScore)
	for slug := range c.classes {
		scores[slug] = classScore{
			numHits:    0,
			hitVariety: set.New[string](),
		}
	}

	return scores
}

func (c *Classifier) collectScores(scores map[string]classScore, stems []string) {
	for _, stem := range stems {
		for slug, class := range c.classes {
			if class.Stems.Stems.Contains(stem) {
				score := scores[slug]

				score.numHits++
				score.hitVariety.Add(stem)

				scores[slug] = score
			}
		}
	}
}

func (c *Classifier) findHighestScoreClass(scores map[string]classScore) *model.ClassifyResult {
	var (
		finalScores = make(map[int]*model.Class)
		scoreValues = make([]int, 0, len(finalScores))
	)

	for slug, score := range scores {
		finalScores[score.score()] = c.classes[slug]
		scoreValues = append(scoreValues, score.score())
	}

	sort.Ints(scoreValues)

	highestScore := scoreValues[len(scoreValues)-1]
	return &model.ClassifyResult{
		Class: finalScores[highestScore],
		Score: highestScore,
	}
}

func LoadClasses(r io.Reader) ([]*model.Class, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var classes []*model.Class
	if err := json.Unmarshal(data, &classes); err != nil {
		return nil, err
	}

	var errs error

	for _, class := range classes {
		var (
			rawStems = class.Stems.Stems.Slice()
			stems    = make([]string, 0, len(rawStems))
		)

		for _, stem := range rawStems {
			stemmed, err := snowball.Stem(stem, "russian", true)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			stems = append(stems, stemmed)
		}

		class.Stems.Stems = set.New(stems...)
	}

	return classes, nil
}

type classScore struct {
	numHits    int
	hitVariety set.HashSet[string]
}

func (c classScore) score() int {
	return c.numHits * len(c.hitVariety)
}
