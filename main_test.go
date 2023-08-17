package main

import (
	"EH_downloader/client"
	"EH_downloader/utils"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/http"
	"path/filepath"
	"testing"
)

func TestGetGalleryInfo(t *testing.T) {
	urlStr := "https://e-hentai.org/g/1838806/61460acecb/"
	c := colly.NewCollector()
	title, maxPage := getGalleryInfo(c, urlStr)

	expectedTitle := "[Homunculus] Bye-Bye Sister (COMIC Kairakuten 2021-02) [Chinese] [CE家族社×無邪気無修宇宙分組] [Digital]"
	expectedMaxPage := 24
	if title != expectedTitle || maxPage != expectedMaxPage {
		t.Errorf("getGalleryInfo() = %s, %d; want %s, %d", title, maxPage, expectedTitle, expectedMaxPage)
	}

	urlStr = "https://zhuanlan.zhihu.com/p/375530785"
	title, maxPage = getGalleryInfo(c, urlStr)
	expectedTitle = ""
	expectedMaxPage = 0
	if title != expectedTitle || maxPage != expectedMaxPage {
		t.Errorf("getGalleryInfo() = %s, %d; want %s, %d", title, maxPage, expectedTitle, expectedMaxPage)
	}
}

func TestGenerateIndexURL(t *testing.T) {
	urlStr := "https://xxx/yyy"
	page := 8

	expected := "https://xxx/yyy?p=8"
	if generateIndexURL(urlStr, page) != expected {
		t.Errorf("generateIndexURL() = %s; want %s", generateIndexURL(urlStr, page), expected)
	}

	page = 0
	expected = "https://xxx/yyy"
	if generateIndexURL(urlStr, page) != expected {
		t.Errorf("generateIndexURL() = %s; want %s", generateIndexURL(urlStr, page), expected)
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
