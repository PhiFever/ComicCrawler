package main

import (
	"EH_downloader/client"
	"EH_downloader/utils"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"net/http"
	"path/filepath"
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
			title, maxPage := getGalleryInfo(c, tc.url)
			assert.Equal(t, tc.expectedTitle, title, "Title mismatch")
			assert.Equal(t, tc.expectedMaxPage, maxPage, "MaxPage mismatch")
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
		t.Run(cast.ToString(tc.page), func(t *testing.T) {
			result := generateIndexURL(url, tc.page)
			assert.Equal(t, tc.expected, result)
		})
	}
}

var imageDataList = []map[string]string{
	{
		"imageName": "1.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP.SQmqQt18WUcWYyuX8fGGGAHaE8?pid=ImgDet&rs=1`,
	},
	{
		"imageName": "2.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP.6L7shpwxVAIr279rA0B1JQHaE7?pid=ImgDet&rs=1`,
	},
	{
		"imageName": "3.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP.i242SBVfAPAhfxY5omlfgQHaLP?pid=ImgDet&rs=1`,
	},
	{
		"imageName": "4.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP._0UYsgLTgJ8WAUYXFXKHRQHaEK?pid=ImgDet&rs=1`,
	},
}

func TestSaveImages(t *testing.T) {
	headers := make(http.Header)
	headers.Set(`User-Agent`, `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.82`)
	//headers.Set("Accept", "image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	//headers.Set("Upgrade-Insecure-Requests", "1")
	c := client.InitCollector(headers)
	saveDir := "./test"
	absPath, err := filepath.Abs(saveDir)
	fmt.Println(absPath)
	if err != nil {
		t.Errorf("filepath.Abs() = %s; want nil", err)
	}
	err = saveImages(c, imageDataList, absPath)
	if err != nil {
		t.Errorf("saveImages() = %s; want nil", err)
	}
	for _, data := range imageDataList {
		if !utils.CacheFileExists(filepath.Join(saveDir, data["imageName"])) {
			t.Errorf("saveImages() = %s; want %s", filepath.Join(saveDir, data["imageName"]), "exist")
		}
	}
}
