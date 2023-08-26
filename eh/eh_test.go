package eh

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getGalleryInfo(t *testing.T) {
	testCases := []struct {
		url                 string
		expectedGalleryInfo GalleryInfo
	}{
		{
			url: "https://e-hentai.org/g/2569708/4bd9316841/",
			expectedGalleryInfo: GalleryInfo{
				URL:        "https://e-hentai.org/g/2569708/4bd9316841/",
				Title:      "[中信出版社] 流浪地球2电影制作手记 The Wandering Earth II FLIM HAND BOOK",
				TotalImage: 468,
				TagList: map[string][]string{
					"language": {"chinese"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			galleryInfo := getGalleryInfo(tc.url)
			assert.Equal(t, tc.expectedGalleryInfo, galleryInfo)
		})
	}
}

func Test_generateIndexURL(t *testing.T) {
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
			result := generateIndexURL(url, tc.page)
			assert.Equal(t, tc.expected, result)
		})
	}
}
