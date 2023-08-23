package dmzj

import (
	"ComicDownloader/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"os"
	"strings"
)

type GalleryInfo struct {
	URL   string `json:"gallery_url"`
	Title string `json:"gallery_title"`
	//最新章节
	LastChapter string              `json:"last_chapter"`
	TagList     map[string][]string `json:"tag_list"`
}

func GetGalleryInfo(galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.TagList = make(map[string][]string)
	galleryInfo.URL = galleryUrl

	//cookies, err := client.ReadCookiesFromFile("../cookies.json")
	//utils.ErrorCheck(err)
	//htmlContent, err := client.GetRenderedPage(galleryUrl, client.ConvertCookies(cookies))
	//// 将 []byte 转换为 io.Reader
	//reader := bytes.NewReader(htmlContent)
	//doc, err := goquery.NewDocumentFromReader(reader)

	htmlContent, _ := os.Open("../static/dmzj_chromedp.html")
	doc, err := goquery.NewDocumentFromReader(htmlContent)

	rp := strings.NewReplacer(">>", "")
	//找到其中<div class="wrap">下的<div class="path_lv3">元素中的最后一个文本节点即为标题
	doc.Find("div.wrap div.path_lv3").Each(func(i int, s *goquery.Selection) {
		galleryInfo.Title = strings.TrimSpace(rp.Replace(s.Contents().Last().Text()))
	})

	//找到<div class="anim-main_list">
	rp = strings.NewReplacer("：", "")
	doc.Find(".anim-main_list table tbody tr").Each(func(index int, row *goquery.Selection) {
		key := strings.TrimSpace(row.Find("th").Text())
		localKey := rp.Replace(key)
		value := strings.TrimSpace(row.Find("td").Text())
		if _, ok := galleryInfo.TagList[localKey]; ok {
			galleryInfo.TagList[localKey] = append(galleryInfo.TagList[localKey], value)
		} else {
			//TODO 有些标签是多值的，需要处理
			galleryInfo.TagList[localKey] = []string{value}
		}
	})
	utils.ErrorCheck(err)

	return galleryInfo
}

func DownloadGallery(infoJson string, galleryUrl string, onlyInfo bool) {
	//获取画廊信息
	galleryInfo := GetGalleryInfo(galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)
}
