package main

import (
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
	if ToSafeFilename(title) != expected {
		t.Errorf("ToSafeFilename() = %s; want %s", ToSafeFilename(title), expected)
	}
}

func TestSaveImages(t *testing.T) {
	c := initCollector()
	var imageDataList []map[string]string
	imageDataList = append(imageDataList, map[string]string{
		"imageName": "1.jpg",
		//"imageUrl":  `https://klibhro.vutxylhuqpcu.hath.network/h/0831b75aa55ca2952676ac3ae86ca44bc5fe500a-718271-1280-1808-jpg/keystamp=1692102000-1766c97dab;fileindex=87912021;xres=1280/MJK_20_Z2477CE_1656_001.jpg`})
		"imageUrl": `https://th.bing.com/th/id/OIP.SQmqQt18WUcWYyuX8fGGGAHaE8?pid=ImgDet&rs=1`})
	imageDataList = append(imageDataList, map[string]string{
		"imageName": "2.jpg",
		//"imageUrl":  `https://jvqboaw.pkjrvmcjplqf.hath.network:2047/h/b48d7f5206c03b112d957184a44af44e8a3894ec-688714-1280-1808-jpg/keystamp=1692102000-1794e07515;fileindex=87912022;xres=1280/MJK_20_Z2477CE_1656_002.jpg`})
		"imageUrl": `https://th.bing.com/th/id/OIP.6L7shpwxVAIr279rA0B1JQHaE7?pid=ImgDet&rs=1`})
	saveDir := "test"
	if SaveImages(c, imageDataList, saveDir) != nil {
		t.Errorf("SaveImages() = %s; want nil", SaveImages(c, imageDataList, "test"))
	}
}

func TestSaveFile(t *testing.T) {
	data := []byte(cast.ToString(time.Now()))
	filePath, _ := filepath.Abs("./test/saveFile.txt")

	err := SaveFile(filePath, data)
	if err != nil {
		t.Errorf("SaveFile() = %s; want nil", err)
	}
}

func TestBuildCache(t *testing.T) {
	var imageDataList []map[string]string
	imageDataList = append(imageDataList, map[string]string{
		"imageName": "1.jpg",
		//"imageUrl":  `https://klibhro.vutxylhuqpcu.hath.network/h/0831b75aa55ca2952676ac3ae86ca44bc5fe500a-718271-1280-1808-jpg/keystamp=1692102000-1766c97dab;fileindex=87912021;xres=1280/MJK_20_Z2477CE_1656_001.jpg`})
		"imageUrl": `https://th.bing.com/th/id/OIP.SQmqQt18WUcWYyuX8fGGGAHaE8?pid=ImgDet&rs=1`})
	imageDataList = append(imageDataList, map[string]string{
		"imageName": "2.jpg",
		//"imageUrl":  `https://jvqboaw.pkjrvmcjplqf.hath.network:2047/h/b48d7f5206c03b112d957184a44af44e8a3894ec-688714-1280-1808-jpg/keystamp=1692102000-1794e07515;fileindex=87912022;xres=1280/MJK_20_Z2477CE_1656_002.jpg`})
		"imageUrl": `https://th.bing.com/th/id/OIP.6L7shpwxVAIr279rA0B1JQHaE7?pid=ImgDet&rs=1`})
	saveDir := "test"
	err := buildCache(saveDir, imageDataList)
	if err != nil {
		t.Errorf("buildCache() = %s; want nil", err)
	}
}
