package main

import (
	"EH_downloader/eh"
	"fmt"
	"github.com/spf13/cast"
	"path/filepath"
	"testing"
	"time"
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

func TestToSafeFilename(t *testing.T) {
	title := `[sfs]\24r/f4?*<q>|:`
	expected := "[sfs]_24r_f4___q___"
	if eh.ToSafeFilename(title) != expected {
		t.Errorf("ToSafeFilename() = %s; want %s", eh.ToSafeFilename(title), expected)
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
	c := eh.InitCollector()
	saveDir := "test"
	if saveImages(c, imageDataList, saveDir) != nil {
		t.Errorf("saveImages() = %s; want nil", saveImages(c, imageDataList, "test"))
	}
}

func TestSaveFile(t *testing.T) {
	data := []byte(cast.ToString(time.Now()))
	filePath, _ := filepath.Abs("./test/saveFile.txt")

	err := eh.SaveFile(filePath, data)
	if err != nil {
		t.Errorf("SaveFile() = %s; want nil", err)
	}
}

func TestBuildCache(t *testing.T) {
	saveDir := "test"
	cacheFile := "cache.json"
	err := eh.BuildCache(saveDir, cacheFile, imageDataList)
	if err != nil {
		t.Errorf("buildCache() = %s; want nil", err)
	}
}

func TestLoadCache(t *testing.T) {
	saveDir := "test"
	cacheFile := "cache.json"
	imageDataList, err := eh.LoadCache(filepath.Join(saveDir, cacheFile))
	for _, data := range imageDataList {
		fmt.Println(data["imageName"])
		fmt.Println(data["imageUrl"])
	}
	if err != nil {
		t.Errorf("loadCache() = %s; want nil", err)
	}

}
