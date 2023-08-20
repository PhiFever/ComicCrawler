package eh

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGalleryInfo(t *testing.T) {
	testCases := []struct {
		url             string
		expectedTitle   string
		expectedMaxPage int
	}{
		{
			url:             "https://e-hentai.org/g/2569708/4bd9316841/",
			expectedTitle:   "[中信出版社] 流浪地球2电影制作手记 The Wandering Earth II FLIM HAND BOOK",
			expectedMaxPage: 468,
		},
		{
			url:             "https://zhuanlan.zhihu.com/p/375530785",
			expectedTitle:   "",
			expectedMaxPage: 0,
		},
	}

	c := colly.NewCollector()

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			galleryInfo := GetGalleryInfo(c, tc.url)
			assert.Equal(t, tc.expectedTitle, galleryInfo.Title, "Title mismatch")
			assert.Equal(t, tc.expectedMaxPage, galleryInfo.TotalImage, "TotalImage mismatch")
		})
	}
}

func TestGenerateIndexURL(t *testing.T) {
	url := "https://xxx/yyy"
	testCases := []struct {
		page     int
		expected string
	}{
		{
			page:     0,
			expected: "https://xxx/yyy",
		},
		{
			page:     8,
			expected: "https://xxx/yyy?p=8",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("page=%d", tc.page), func(t *testing.T) {
			result := GenerateIndexURL(url, tc.page)
			assert.Equal(t, tc.expected, result)
		})
	}
}
