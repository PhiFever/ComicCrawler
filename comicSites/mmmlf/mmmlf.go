package mmmlf

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type GalleryInfo struct {
	URL            string              `json:"gallery_url"`
	Title          string              `json:"gallery_title"`
	LastChapter    string              `json:"last_chapter"`
	LastUpdateTime string              `json:"last_update_time"`
	TagList        map[string][]string `json:"tag_list"`
}

func getGalleryInfo(c *colly.Collector, url string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.TagList = make(map[string][]string)
	galleryInfo.URL = url

	//找到<h1>标签,即为文章标题
	c.OnHTML("h1", func(e *colly.HTMLElement) {
		galleryInfo.Title = e.Text
	})

	//TODO:其他信息待续

	err := c.Visit(url)
	utils.ErrorCheck(err)
	return galleryInfo
}

func buildJpegRequestHeaders() http.Header {
	headers := http.Header{
		"authority": []string{"16t.765567.xyz"},
		"method":    []string{"GET"},
		//"path":                      []string{"/cdn/copymgpic/tianmijiayuan/1/e6b7c744-b0d5-11eb-af55-024352452ce0.jpg"},
		"scheme":                    []string{"https"},
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Encoding":           []string{"gzip, deflate, br"},
		"Accept-Language":           []string{"zh-CN,zh;q=0.9"},
		"Cache-Control":             []string{"no-cache"},
		"Dnt":                       []string{"1"},
		"Pragma":                    []string{"no-cache"},
		"Sec-Ch-Ua":                 []string{"\"Chromium\";v=\"116\", \"Not)A;Brand\";v=\"24\", \"Google Chrome\";v=\"116\""},
		"Sec-Ch-Ua-Mobile":          []string{"?0"},
		"Sec-Ch-Ua-Platform":        []string{"\"Windows\""},
		"Sec-Fetch-Dest":            []string{"document"},
		"Sec-Fetch-Mode":            []string{"navigate"},
		"Sec-Fetch-Site":            []string{"none"},
		"Sec-Fetch-User":            []string{"?1"},
		"Sec-Gpc":                   []string{"1"},
		"Upgrade-Insecure-Requests": []string{"1"},
		"User-Agent":                []string{client.ChromeUserAgent},
	}

	return headers
}

func getImageUrlListFromPage(c *colly.Collector, url string) []string {
	var imageUrlList []string
	//找到 <div class="comiclist">
	c.OnHTML("div.comiclist", func(e *colly.HTMLElement) {
		//对其中的每个 <div class="comicpage"> 进行处理
		e.ForEach("div.comicpage", func(_ int, div *colly.HTMLElement) {
			//获取每个 <img> 的src属性
			src := div.ChildAttr("img", "src")
			//有时候会有奇怪的失效链接，需要去除
			if !strings.Contains(src, "https://mirror.mangafunc.fun/comic/") {
				imageUrlList = append(imageUrlList, src)
			}
		})
	})
	err := c.Visit(url)
	utils.ErrorCheck(err)
	return imageUrlList
}

func getImagePageInfoList(c *colly.Collector, url string) (imagePageInfoList []map[int]string, indexToTitleMapList []map[int]string) {
	//找到 <ul class="chapter-list clearfix" id="chapterList">
	c.OnHTML("ul#chapterList", func(e *colly.HTMLElement) {
		//对每个<li>标签进行处理
		//FIXME:此处可能需要根据实际目录情况变动，因为有时候是从第1话开始有时候是第0话开始
		index := 0
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			//<a>标签的href属性即为图片页面的url
			imagePageUrl := li.ChildAttr("a", "href")
			//<a>标签的text即为图片页面的标题
			imagePageTitle := li.ChildText("a")
			imagePageInfo := map[int]string{
				index: "https://mmmlf.com" + imagePageUrl,
			}
			imagePageInfoList = append(imagePageInfoList, imagePageInfo)
			indexToTitleMapList = append(indexToTitleMapList, map[int]string{
				index: imagePageTitle,
			})
			index++
		})
	})

	err := c.Visit(url)
	utils.ErrorCheck(err)
	return imagePageInfoList, indexToTitleMapList
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	beginIndex := 0

	baseCollector := colly.NewCollector(
		colly.UserAgent(client.ChromeUserAgent),
	)
	//获取画廊信息
	galleryInfo := getGalleryInfo(baseCollector, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		imagePath, err := filepath.Abs(safeTitle)
		utils.ErrorCheck(err)
		beginIndex = utils.GetBeginIndex(imagePath, []string{".jpg", ".png"})
	} else {
		//生成缓存文件
		err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
		utils.ErrorCheck(err)
	}
	if onlyInfo {
		fmt.Println("画廊信息获取完毕，程序自动退出。")
		return
	}

	fmt.Println("beginIndex=", beginIndex)
	jpegCollector := client.InitJPEGCollector(buildJpegRequestHeaders())

	//获取所有图片页面的url
	imagePageInfoList, indexToTitleMapList := getImagePageInfoList(jpegCollector, galleryUrl)
	//FIXME:为什么要减1？因为beginIndex是从0开始的，而imagePageUrlList的index是从1开始的
	//所以当beginIndex=0时程序会报错
	//if beginIndex != 0 {
	//	imagePageInfoList = imagePageInfoList[beginIndex-1:]
	//}
	imagePageInfoList = imagePageInfoList[beginIndex:]

	err := utils.BuildCache(safeTitle, "menu.json", indexToTitleMapList)
	utils.ErrorCheck(err)

	//对每话的页面进行处理
	for _, pageInfo := range imagePageInfoList {
		var imageInfoList []map[string]string
		for key, pageUrl := range pageInfo {
			imageUrlList := getImageUrlListFromPage(baseCollector, pageUrl)
			for i, imageUrl := range imageUrlList {
				imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
				imageInfo := map[string]string{
					"imageTitle": cast.ToString(key) + "_" + cast.ToString(i) + imageSuffix,
					"imageUrl":   imageUrl,
				}
				imageInfoList = append(imageInfoList, imageInfo)
			}

			utils.SaveImages(jpegCollector, imageInfoList, safeTitle)
			//防止被ban，每处理一篇目录就sleep 5-10 seconds
			sleepTime := client.TrueRandFloat(5, 10)
			log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
}
