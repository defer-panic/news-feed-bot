package classifier_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/defer-panic/news-feed-bot/internal/classifier"
)

//go:embed data/classes.json
var classesData []byte

//go:embed testdata/finance.txt
var financeText string

//go:embed testdata/high-tech.txt
var highTechText string

//go:embed testdata/crypto.txt
var cryptoText string

//go:embed testdata/pop-culture.txt
var popCultureText string

//go:embed testdata/science.txt
var scienceText string

//go:embed testdata/sport.txt
var sportText string

func TestClassifier_Classify(t *testing.T) {
	testCases := []struct {
		name         string
		text         string
		expectedSlug string
	}{
		{
			name:         "finance",
			text:         financeText,
			expectedSlug: "finance",
		},
		{
			name:         "high-tech",
			text:         highTechText,
			expectedSlug: "high-tech",
		},
		{
			name:         "crypto",
			text:         cryptoText,
			expectedSlug: "crypto",
		},
		{
			name:         "pop-culture",
			text:         popCultureText,
			expectedSlug: "pop-culture",
		},
		{
			name:         "science",
			text:         scienceText,
			expectedSlug: "science",
		},
		{
			name:         "sport",
			text:         sportText,
			expectedSlug: "sport",
		},
	}

	classes, err := classifier.LoadClasses(bytes.NewReader(classesData))
	require.NoError(t, err)

	classifier := classifier.New(classes)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			class, err := classifier.Classify(tc.text)
			require.NoError(t, err)

			require.Equal(t, tc.expectedSlug, class.Slug)
		})
	}
}
