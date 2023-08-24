package dmzj

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"path/filepath"
	"regexp"
	"strings"
)

const cookiesPath = `cookies.json`

type GalleryInfo struct {
	URL            string              `json:"gallery_url"`
	Title          string              `json:"gallery_title"`
	LastChapter    string              `json:"last_chapter"`
	LastUpdateTime string              `json:"last_update_time"`
	TagList        map[string][]string `json:"tag_list"`
}

func GetGalleryInfo(doc *goquery.Document, galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.TagList = make(map[string][]string)
	galleryInfo.URL = galleryUrl

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
		row.Find("td").Each(func(index int, cell *goquery.Selection) {
			cell.Find("a").Each(func(index int, a *goquery.Selection) {
				galleryInfo.TagList[localKey] = append(galleryInfo.TagList[localKey], strings.TrimSpace(a.Text()))
			})
			//找到最后更新时间
			cell.Find("span").Each(func(index int, span *goquery.Selection) {
				galleryInfo.LastUpdateTime = strings.TrimSpace(span.Text())
			})
		})
	})

	rp = strings.NewReplacer("第", "", "话", "")
	lastChapter, _ := regexp.MatchString(`第(\d+)话`, galleryInfo.TagList["最新收录"][0])
	if lastChapter {
		galleryInfo.LastChapter = rp.Replace(galleryInfo.TagList["最新收录"][0])
	} else {
		galleryInfo.LastChapter = "未知"
	}
	return galleryInfo
}

func GetAllImagePageUrl(doc *goquery.Document) []string {
	var imagePageUrl []string
	//找到<div class="cartoon_online_border">
	doc.Find("div.cartoon_online_border").Each(func(i int, s *goquery.Selection) {

	})
	return imagePageUrl
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	beginIndex := 0
	doc := client.GetHtmlDoc(cookiesPath, galleryUrl)
	//获取画廊信息
	galleryInfo := GetGalleryInfo(doc, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		//读取缓存文件
		var lastGalleryInfo GalleryInfo
		err := utils.LoadCache(filepath.Join(safeTitle, infoJsonPath), &lastGalleryInfo)
		utils.ErrorCheck(err)
		if lastGalleryInfo.LastChapter == galleryInfo.LastChapter {
			fmt.Println("本gallery已经下载完毕")
			return
		} else {
			fmt.Println("发现更新，继续下载更新部分")
			beginIndex = cast.ToInt(lastGalleryInfo.LastChapter)
		}
	} else {
		//生成缓存文件
		err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
		utils.ErrorCheck(err)
		if onlyInfo {
			fmt.Println("画廊信息获取完毕，程序自动退出。")
			return
		}
	}

	fmt.Println(beginIndex)
}
