// Package happymh m.happymh.com
package happymh

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"ComicCrawler/utils/stack"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"path/filepath"
	"strings"
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

func buildRequestHeaders() http.Header {
	headers := http.Header{
		"authority": []string{"ruicdn.happymh.com"},
		"method":    []string{"GET"},
		//"path":      []string{"/1f290a226753ed7e0c3d3689e1c84102/0d1ecb53f86000d7d0f95d23cfd2015e.jpg"},
		"scheme":          []string{"https"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Encoding": []string{"gzip, deflate, br"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en;q=0.8"},
		"Cache-Control":   []string{"no-cache"},
		//"Cookie": []string{"sf_token=B88E1644-1FB1-FB83-2945-19E734C2F16F; _ga=GA1.1.127078344.1694264199; cf_clearance=rLcfj4LdjlT0LIzmXwvjqrSNc9JPvZN4pHpxNEbOEjE-1694265652-0-1-1a7f072b.875d48f.4d7e58cb-250.2.1694265652; _ga_E2G2LX2SKZ=GS1.1.1694264198.1.1.1694267583.15.0.0; __cf_bm=CkK4E0mHt5DVBFl7sUxpWxe4bzd477uP27j1A6huwRU-1694269904-0-AaGCFWMXrXyQmDsW5rIBSNtr1ER353tp997QxtDtoX5Lz02r3XdytdKmn2hFWr219fSfmQExchbITvdtU87PjWQ="},
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

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	beginIndex := 0

	cookies := client.ReadCookiesFromFile(cookiesPath)
	cookiesParam := client.ConvertCookies(cookies)
	// 初始化 Chromedp 上下文
	ctx, cancel := client.InitChromedpContext(false)
	defer cancel()
	menuDoc := client.GetHtmlDoc(client.GetClickedRenderedPage(ctx, galleryUrl, cookiesParam, "#expandButton"))

	//获取画廊信息
	galleryInfo := getGalleryInfo(menuDoc, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
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

	//获取所有图片页面的url
	imagePageInfoList, indexToTitleMapList := getImagePageInfoList(menuDoc)
	imagePageInfoList = imagePageInfoList[beginIndex:]
	fmt.Println(len(imagePageInfoList))
	fmt.Println(len(indexToTitleMapList))
	err := utils.BuildCache(safeTitle, "menu.json", indexToTitleMapList)
	utils.ErrorCheck(err)
}
