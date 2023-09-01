package zuimh

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/http"
	"path/filepath"
)

const (
	cookiesPath = `cookies.json`
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

func getAllImagePageInfo(c *colly.Collector, indexUrl string) (imagePageInfoList []map[int]string, indexToNameMap []map[int]string) {
	//<ul id="chapter-list-1" data-sort="asc">
	c.OnHTML("ul#chapter-list-1", func(e *colly.HTMLElement) {
		//对每个<li>标签进行处理
		e.ForEach("li", func(index int, li *colly.HTMLElement) {
			//<a>标签的href属性即为图片页面的url
			imagePageUrl := li.ChildAttr("a", "href")
			//<span>标签的text即为图片页面的标题
			imagePageTitle := li.ChildText("span")
			imagePageInfo := map[int]string{
				index: "https://www.zuimh.com" + imagePageUrl,
			}
			imagePageInfoList = append(imagePageInfoList, imagePageInfo)
			indexToNameMap = append(indexToNameMap, map[int]string{
				index: imagePageTitle,
			})
			index++
		})
	})

	err := c.Visit(indexUrl)
	utils.ErrorCheck(err)
	return imagePageInfoList, indexToNameMap
}

func buildJpegRequestHeaders() http.Header {
	headers := http.Header{
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
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"},
	}

	return headers
}

// getImageUrlFromPage 从单个图片页获取图片地址
func getImageUrlFromPage(c *colly.Collector, imagePageUrl string) []string {
	var imageUrlList []string
	c.OnHTML("script", func(e *colly.HTMLElement) {
		scriptContent := e.Text
		fmt.Println(scriptContent)

		// 在这里继续解析 scriptContent，提取 var chapterImages 的值
		// 使用与前面示例中相似的方法来提取数组
	})

	err := c.Visit(imagePageUrl)
	utils.ErrorCheck(err)
	return imageUrlList
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	//目录号
	beginIndex := 0
	//needUpdate := false

	baseCollector := client.InitCookiesCollector(client.ReadCookiesFromFile(cookiesPath), "https://www.zuimh.com/")

	//获取画廊信息
	galleryInfo := getGalleryInfo(baseCollector, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		//读取缓存文件
		var lastGalleryInfo GalleryInfo
		err := utils.LoadCache(filepath.Join(safeTitle, infoJsonPath), &lastGalleryInfo)
		utils.ErrorCheck(err)

		//needUpdate = utils.CheckUpdate(lastGalleryInfo.LastUpdateTime, galleryInfo.LastUpdateTime)
		//if needUpdate {
		//	fmt.Println("发现新章节，更新下载记录")
		//	err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
		//	utils.ErrorCheck(err)
		//} else {
		//	fmt.Println("无需更新下载记录")
		//}
		mainImagePath, err := filepath.Abs(safeTitle)
		utils.ErrorCheck(err)
		beginIndex = utils.GetBeginIndex(mainImagePath, []string{".jpg", ".png"})
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
	c := client.InitJPEGCollector(buildJpegRequestHeaders())

	//获取所有图片页面的url
	imagePageUrlList, indexToNameMap := getAllImagePageInfo(c, galleryUrl+"#chapters")
	fmt.Println("Total Chapter:", len(imagePageUrlList))
	fmt.Println(len(indexToNameMap))

	for _, imagePage := range imagePageUrlList {
		for index, url := range imagePage {
			imageUrlList := getImageUrlFromPage(c, url)
			fmt.Println(indexToNameMap[index])
			fmt.Println(imageUrlList)
		}
		break
	}
}
