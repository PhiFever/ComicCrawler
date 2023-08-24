package dmzj

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const cookiesPath = `cookies.json`

type GalleryInfo struct {
	URL            string              `json:"gallery_url"`
	Title          string              `json:"gallery_title"`
	LastChapter    string              `json:"last_chapter"`
	LastUpdateTime string              `json:"last_update_time"`
	TagList        map[string][]string `json:"tag_list"`
}

// GetGalleryInfo 从主目录页获取画廊信息
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

// GetAllImagePageInfo 从主目录页获取所有图片页地址
func GetAllImagePageInfo(doc *goquery.Document) []map[int]string {
	var imagePageInfoMap []map[int]string
	// 找到<div class="cartoon_online_border">
	doc.Find("div.cartoon_online_border").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				imageName := strings.TrimSpace(a.Text())

				indexStr, err := utils.ExtractNumberFromText(`第(\d+)话`, imageName)
				utils.ErrorCheck(err)
				//cast库在转换时字符串若是以 "0" 开头，"07" 转换后得到整型 7，而 "08" 转换后得到整型 0
				//https://iokde.com/post/golang-cast64-snare.html
				imageIndex, _ := strconv.Atoi(indexStr)

				imageInfo := map[int]string{
					imageIndex: "https://manhua.dmzj.com" + href,
				}
				imagePageInfoMap = append(imagePageInfoMap, imageInfo)
			}
		})
	})

	return imagePageInfoMap
}

// GetImageUrlFromPage 从单个图片页获取图片地址
func GetImageUrlFromPage(doc *goquery.Document) string {
	var imageUrl string
	//找到<div class="comic_wraCon autoHeight"
	doc.Find("div.comic_wraCon.autoHeight").Each(func(i int, s *goquery.Selection) {
		//找到<img>的src属性
		src, exists := s.Find("img").Attr("src")
		if exists {
			imageUrl = src
		}
	})
	return imageUrl
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	beginIndex := 0

	cookies, err := client.ReadCookiesFromFile(cookiesPath)
	utils.ErrorCheck(err)
	cookiesParam := client.ConvertCookies(cookies)
	menuDoc := client.GetHtmlDoc(cookiesParam, galleryUrl)

	//获取画廊信息
	galleryInfo := GetGalleryInfo(menuDoc, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		//读取缓存文件
		var lastGalleryInfo GalleryInfo
		err := utils.LoadCache(filepath.Join(safeTitle, infoJsonPath), &lastGalleryInfo)
		utils.ErrorCheck(err)
		//获取已经下载的图片数量
		downloadedImageCount := utils.GetFileTotal(safeTitle, []string{".jpg", ".png"})
		fmt.Println("Downloaded image count:", downloadedImageCount)
		//计算剩余图片数量
		remainImageCount := cast.ToInt(galleryInfo.LastChapter) - downloadedImageCount
		if remainImageCount == 0 && lastGalleryInfo.LastChapter == galleryInfo.LastChapter {
			fmt.Println("本gallery已经下载完毕")
			return
		} else if remainImageCount < 0 || lastGalleryInfo.LastChapter > galleryInfo.LastChapter {
			fmt.Println("下载记录有误！")
			return
		} else {
			fmt.Println("剩余图片数量:", remainImageCount)
			beginIndex = downloadedImageCount
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

	fmt.Println("beginIndex= ", beginIndex)
	imagePageInfoMap := GetAllImagePageInfo(menuDoc)[beginIndex:]

	////测试用
	//for _, info := range imagePageInfoMap {
	//	for index, url := range info {
	//		fmt.Println(index, url)
	//	}
	//}
	//log.Println("imagePageInfoMap length:", len(imagePageInfoMap))

	tasks := make(chan map[int]string, len(imagePageInfoMap))
	for _, info := range imagePageInfoMap {
		tasks <- info
	}
	close(tasks)

	numWorkers := 5 // 可以根据需求调整并发的数量
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for info := range tasks {
				for index, url := range info {
					fmt.Println(index, url)
					pageDoc := client.GetHtmlDoc(cookiesParam, url)
					//获取图片地址
					imageUrl := GetImageUrlFromPage(pageDoc)
					fmt.Println(imageUrl)
				}
			}
		}()
	}

	wg.Wait() // 等待所有任务完成

}
