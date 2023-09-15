// Package happymh m.happymh.com
package happymh

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"ComicCrawler/utils/stack"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"log"
	"path/filepath"
	"strings"
	"time"
)

const (
	cookiesPath = `cookies.json`
)

type GalleryInfo struct {
	URL    string `json:"gallery_url"`
	Title  string `json:"gallery_title"`
	Author string `json:"author"`
}

// getGalleryInfo 从主目录页获取画廊信息
func getGalleryInfo(doc *goquery.Document, galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.URL = galleryUrl

	//找到<h1>标签,即为文章标题
	galleryInfo.Title = strings.TrimSpace(doc.Find("h2").Text())
	return galleryInfo
}

// getImageUrlListFromPage 从单个图片页获取图片地址
func getImageUrlListFromPage(doc *goquery.Document) []string {
	var imageUrlList []string
	doc.Find("#root > div > article.jss17").Each(func(i int, s *goquery.Selection) {
		s.Find("img").Each(func(j int, img *goquery.Selection) {
			src, exists := img.Attr("src")
			if exists {
				imageUrlList = append(imageUrlList, src)
			}
		})
	})
	return imageUrlList
}

// getImagePageInfoList 从主目录页获取所有图片页的url
func getImagePageInfoList(doc *goquery.Document) (imagePageInfoList []map[int]string, indexToTitleMapList []map[int]string) {
	imageInfoStack := stack.Stack{}
	// 找到<div class="cartoon_online_border">
	doc.Find("#limitList > div").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				imagePageTitle := strings.TrimSpace(a.Text())
				imagePageInfo := map[string]string{
					imagePageTitle: href,
				}
				imageInfoStack.Push(imagePageInfo)
			}
		})
	})

	index := 0
	//直接处理得到的是逆序序列，通过栈转换为正序
	for !imageInfoStack.IsEmpty() {
		item := imageInfoStack.Pop()
		imageInfo := item.(map[string]string)
		for imagePageTitle, imagePageUrl := range imageInfo {
			imagePageInfo := map[int]string{
				index: imagePageUrl,
			}
			imagePageInfoList = append(imagePageInfoList, imagePageInfo)
			indexToTitleMapList = append(indexToTitleMapList, map[int]string{index: imagePageTitle})
			index++
		}
	}
	return imagePageInfoList, indexToTitleMapList
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) error {
	beginIndex := 0

	cookies := client.ReadCookiesFromFile(cookiesPath)
	cookiesParam := client.ConvertCookies(cookies)
	// 初始化 Chromedp 上下文
	chromeCtx, cancel := client.InitChromedpContext(true)
	defer cancel()
	menuDoc := client.GetHtmlDoc(client.GetClickedRenderedPage(chromeCtx, cookiesParam, galleryUrl, "#expandButton"))

	//获取画廊信息
	galleryInfo := getGalleryInfo(menuDoc, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		mainImagePath, err := filepath.Abs(safeTitle)
		if err != nil {
			return err
		}
		beginIndex = utils.GetBeginIndex(mainImagePath, []string{".jpg", ".png"})
	} else {
		//生成缓存文件
		err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
		if err != nil {
			return err
		}
	}
	if onlyInfo {
		fmt.Println("画廊信息获取完毕，程序自动退出。")
		return nil
	}

	fmt.Println("beginIndex=", beginIndex)

	//获取所有图片页面的url
	imagePageInfoList, indexToTitleMapList := getImagePageInfoList(menuDoc)
	imagePageInfoList = imagePageInfoList[beginIndex:]

	err := utils.BuildCache(safeTitle, "menu.json", indexToTitleMapList)
	if err != nil {
		return err
	}

	fmt.Println("正在下载图片...")
	//FIXME: 此处需要初始化一个新的chromedp上下文，可能是因为浏览器缓存的原因，如果不初始化新的上下文，会导致后续的页面异常加载
	pageCtx, pageCancel := client.InitChromedpContext(true)
	defer pageCancel()
	for _, pageInfo := range imagePageInfoList {
		var imageInfoList []map[string]string
		for index, url := range pageInfo {
			pageDoc := client.GetHtmlDoc(client.GetScrolledRenderedPage(pageCtx, cookiesParam, url))
			client.ChromedpClearCash(pageCtx)
			//获取图片地址
			imageUrlList := getImageUrlListFromPage(pageDoc)
			for k, imageUrl := range imageUrlList {
				imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
				imageInfo := map[string]string{
					"imageTitle": cast.ToString(index) + "_" + cast.ToString(k) + imageSuffix,
					"imageUrl":   imageUrl,
				}
				imageInfoList = append(imageInfoList, imageInfo)
			}
		}
		if len(imageInfoList) == 0 {
			cancel()
			pageCancel()
			return fmt.Errorf("imageInfoList is empty, please check browser")
		}
		//防止被ban，每处理一篇目录就sleep 5-10 seconds
		sleepTime := client.TrueRandFloat(5, 10)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)

		//TODO:这块理论上可以设置成异步下载，chromeCtx下载图片，pageCtx解析页面
		for _, imageInfo := range imageInfoList {
			client.ChromedpDownloadImage(chromeCtx, cookiesParam, imageInfo, safeTitle)
			time.Sleep(time.Millisecond * time.Duration(client.DelayMs))
		}
	}
	return nil
}
