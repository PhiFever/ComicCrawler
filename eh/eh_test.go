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
			url:             "https://e-hentai.org/g/1838806/61460acecb/",
			expectedTitle:   "[Homunculus] Bye-Bye Sister (COMIC Kairakuten 2021-02) [Chinese] [CE家族社×無邪気無修宇宙分組] [Digital]",
			expectedMaxPage: 24,
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
