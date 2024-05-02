package eh

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ybbus/httpretry"
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
		c := httpretry.NewDefaultClient()
		t.Run(tc.url, func(t *testing.T) {
			//galleryInfo := getGalleryInfo(tc.url)
			galleryInfo := getGalleryInfo(c, tc.url)
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

func Test_getImagePageUrlList(t *testing.T) {
	type args struct {
		indexUrl string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "流浪地球2电影制作手记",
			args: args{
				indexUrl: "https://e-hentai.org/g/2569708/4bd9316841/",
			},
			want: []string{
				"https://e-hentai.org/s/e4ee2a1bd1/2569708-1",
				"https://e-hentai.org/s/0196805342/2569708-2",
				"https://e-hentai.org/s/11cff112ba/2569708-3",
				"https://e-hentai.org/s/aa3f8ef141/2569708-4"},
		},
	}
	for _, tt := range tests {
		c := httpretry.NewDefaultClient()
		t.Run(tt.name, func(t *testing.T) {
			//got := getImagePageUrlList(tt.args.c, tt.args.indexUrl)[0:4]
			got := getImagePageUrlList(c, tt.args.indexUrl)[0:4]
			assert.Equalf(t, tt.want, got, "getImagePageUrlList(%v)", tt.args.indexUrl)
		})
	}
}
