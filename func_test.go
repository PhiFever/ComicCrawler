package main

import (
	"EH_downloader/client"
	"EH_downloader/utils"
	"fmt"
	"path/filepath"
	"testing"
)

//func TestGetGalleryInfo(t *testing.T) {
//	urlStr := "https://e-hentai.org/g/1838806/61460acecb/"
//	c := colly.NewCollector()
//	title, maxPage := getGalleryInfo(c, urlStr)
//
//	expectedTitle := "[Homunculus] Bye-Bye Sister (COMIC Kairakuten 2021-02) [Chinese] [CE家族社×無邪気無修宇宙分組] [Digital]"
//	if title != expectedTitle {
//		t.Errorf("getGalleryInfo() title = %s; want %s", title, expectedTitle)
//	}
//
//	expectedMaxPage := 24
//	if maxPage != expectedMaxPage {
//		t.Errorf("getGalleryInfo() max_page = %d; want %d", maxPage, expectedMaxPage)
//	}
//}

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
	c := client.InitCollector()
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
