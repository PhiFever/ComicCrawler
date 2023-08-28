package dmzj

import (
	"ComicCrawler/client"
	"ComicCrawler/stack"
	"ComicCrawler/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/spf13/cast"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	cookiesPath       = `dmzj_cookies.json`
	numWorkers        = 5   //页面处理的并发量
	batchSize         = 10  //每次下载的图片页面数量，建议为numWorkers的整数倍
	maxImageInOnePage = 100 //单个图片页中最大图片数量，用于初始化channel（感觉应该不会超过50，但是为了保险起见还是设置为100）
	otherDir          = `其他系列`
)

type GalleryInfo struct {
	URL            string              `json:"gallery_url"`
	Title          string              `json:"gallery_title"`
	LastChapter    string              `json:"last_chapter"`
	LastUpdateTime string              `json:"last_update_time"`
	TagList        map[string][]string `json:"tag_list"`
}

// getGalleryInfo 从主目录页获取画廊信息
func getGalleryInfo(doc *goquery.Document, galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.TagList = make(map[string][]string)
	galleryInfo.URL = galleryUrl

	//找到其中<div class="wrap">下的<div class="path_lv3">元素中的最后一个文本节点即为标题
	doc.Find("div.wrap div.path_lv3").Each(func(i int, s *goquery.Selection) {
		galleryInfo.Title = strings.TrimSpace(strings.ReplaceAll(s.Contents().Last().Text(), ">>", ""))
	})

	//找到<div class="anim-main_list">，即为tagList
	doc.Find(".anim-main_list table tbody tr").Each(func(index int, row *goquery.Selection) {
		key := strings.TrimSpace(row.Find("th").Text())
		localKey := strings.ReplaceAll(key, "：", "")
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

	if galleryInfo.TagList["最新收录"] != nil {
		galleryInfo.LastChapter = galleryInfo.TagList["最新收录"][0]
	} else {
		galleryInfo.LastChapter = "未知"
	}
	return galleryInfo
}

func checkUpdate(lastUpdateTime string, newTime string) bool {
	layout := "2006-01-02" //时间格式模板
	parsedDate1, err := time.Parse(layout, lastUpdateTime)
	if err != nil {
		fmt.Println("日期解析错误:", err)
		return true
	}
	parsedDate2, err := time.Parse(layout, newTime)
	if err != nil {
		fmt.Println("日期解析错误:", err)
		return true
	}

	if parsedDate1.Before(parsedDate2) {
		return true
	} else if parsedDate1.After(parsedDate2) {
		fmt.Println("解析的日期晚于当前日期，galleryInfo.json文件异常")
		return true
	} else {
		return false
	}
}

// getAllImagePageInfoBySelector 从主目录页获取所有`selector`图片页地址
// selector的值为`div.cartoon_online_border`或`div.cartoon_online_border_other`，
// 返回2个切片，元素均为map[int]string
// imageOtherPageInfoList key为图片页序号，value为图片页地址
// indexToNameMap key为图片页序号，value为图片页名字
func getAllImagePageInfoBySelector(selector string, doc *goquery.Document) (imageOtherPageInfoList []map[int]string, indexToNameMap []map[int]string) {
	imageInfoStack := stack.Stack{}
	// 找到<div class="cartoon_online_border">
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				imageName := strings.TrimSpace(a.Text())
				imageInfo := map[string]string{
					imageName: "https://manhua.dmzj.com" + href,
				}
				imageInfoStack.Push(imageInfo)
			}
		})
	})

	index := 1
	//直接处理得到的是逆序序列，通过栈转换为正序
	for !imageInfoStack.IsEmpty() {
		item := imageInfoStack.Pop()
		imageInfo := item.(map[string]string)
		for imageName, imageUrl := range imageInfo {
			imageOtherPageInfo := map[int]string{
				index: imageUrl,
			}
			imageOtherPageInfoList = append(imageOtherPageInfoList, imageOtherPageInfo)
			indexToNameMap = append(indexToNameMap, map[int]string{index: imageName})
			index++
		}
	}
	return imageOtherPageInfoList, indexToNameMap
}

// getImageUrlFromPage 从单个图片页获取图片地址
func getImageUrlFromPage(doc *goquery.Document) []string {
	var imageUrlList []string
	//找到<div class="scrollbar-demo-item"
	doc.Find("div.scrollbar-demo-item").Each(func(i int, s *goquery.Selection) {
		s.Find("img").Each(func(j int, img *goquery.Selection) {
			src, exists := img.Attr("src")
			if exists {
				imageUrlList = append(imageUrlList, src)
			}
		})
	})
	return imageUrlList
}

func buildJpegRequestHeaders() http.Header {
	headers := http.Header{
		"authority": []string{"images.idmzj.com"},
		"method":    []string{"GET"},
		//"path": []string{
		//	"/c%2F%E6%88%90%E4%B8%BA%E5%A4%BA%E5%BF%83%E9%AD%94%E7%9A%84%E5%BF%85%E8%A6%81%2F%E7%AC%AC02%E8%AF%9D_1597930984%2F41.jpg",
		//},
		"scheme": []string{"https"},
		"Accept": []string{
			"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		},
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
		"User-Agent": []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		},
	}

	return headers
}

// batchDownloadImage 按batchSize分组渲染页面，获取图片url并保存图片
func batchDownloadImage(cookiesParam []*network.CookieParam, imagePageInfoList []map[int]string, saveDir string) {
	for batchIndex := 0; batchIndex < len(imagePageInfoList); batchIndex += batchSize {
		//每次循环都重新初始化channel和切片
		subImagePageInfoList := imagePageInfoList[batchIndex:utils.MinInt(batchIndex+batchSize, len(imagePageInfoList))]
		imagePageInfoListChannel := make(chan map[int]string, len(subImagePageInfoList))
		//每个图片页中最多有maxImageInOnePage张图片
		imageInfoChannelSize := maxImageInOnePage * len(subImagePageInfoList)
		imageInfoChannel := make(chan map[string]string, imageInfoChannelSize)
		var imageInfoList []map[string]string

		for _, info := range subImagePageInfoList {
			imagePageInfoListChannel <- info
		}
		close(imagePageInfoListChannel)

		sumImage := utils.SyncParsePage(getImageUrlFromPage, imagePageInfoListChannel, imageInfoChannel, cookiesParam, numWorkers)
		close(imageInfoChannel)
		//FIXME:当sumImage>imageInfoChannelSize时，会导致程序死锁
		//在这个channel里只有sumImage个元素，所以只需要循环sumImage次
		for i := 0; i < sumImage; i++ {
			imageInfo := <-imageInfoChannel
			imageInfoList = append(imageInfoList, imageInfo)
		}
		// 进行本次处理目录中所有图片的批量保存
		baseCollector := client.InitCollector(buildJpegRequestHeaders())
		err := utils.SaveImages(baseCollector, imageInfoList, saveDir)
		utils.ErrorCheck(err)

		//防止被ban，每保存一组图片就sleep 5-15 seconds
		sleepTime := utils.TrueRandFloat(5, 15)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	mainBeginIndex := 0
	otherBeginIndex := 0
	needUpdate := false

	cookies, err := client.ReadCookiesFromFile(cookiesPath)
	utils.ErrorCheck(err)
	cookiesParam := client.ConvertCookies(cookies)
	menuDoc := client.GetHtmlDoc(cookiesParam, galleryUrl)

	//获取画廊信息
	galleryInfo := getGalleryInfo(menuDoc, galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		//读取缓存文件
		var lastGalleryInfo GalleryInfo
		err := utils.LoadCache(filepath.Join(safeTitle, infoJsonPath), &lastGalleryInfo)
		utils.ErrorCheck(err)

		needUpdate = checkUpdate(lastGalleryInfo.LastUpdateTime, galleryInfo.LastUpdateTime)
		if needUpdate {
			fmt.Println("发现新章节，更新下载记录")
			err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
			utils.ErrorCheck(err)
		} else {
			fmt.Println("无需更新下载记录")
		}
		mainImagePath, err := filepath.Abs(safeTitle)
		utils.ErrorCheck(err)
		mainBeginIndex = utils.GetBeginIndex(mainImagePath, []string{".jpg", ".png"})

		otherImagePath, err := filepath.Abs(filepath.Join(safeTitle, otherDir))
		utils.ErrorCheck(err)
		otherBeginIndex = utils.GetBeginIndex(otherImagePath, []string{".jpg", ".png"})
	} else {
		//生成缓存文件
		err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
		utils.ErrorCheck(err)
		if onlyInfo {
			fmt.Println("画廊信息获取完毕，程序自动退出。")
			return
		}
	}
	fmt.Println("mainBeginIndex=", mainBeginIndex)
	fmt.Println("otherBeginIndex=", otherBeginIndex)

	imagePageInfoList, indexToNameMap := getAllImagePageInfoBySelector("div.cartoon_online_border", menuDoc)
	imagePageInfoList = imagePageInfoList[mainBeginIndex:]
	otherImagePageInfoList, otherIndexToNameMap := getAllImagePageInfoBySelector("div.cartoon_online_border_other", menuDoc)
	otherImagePageInfoList = otherImagePageInfoList[otherBeginIndex:]

	err = utils.BuildCache(safeTitle, "menu.json", indexToNameMap)
	utils.ErrorCheck(err)
	otherPath := filepath.Join(safeTitle, otherDir)
	if otherImagePageInfoList != nil {
		err = utils.BuildCache(otherPath, "menu.json", otherIndexToNameMap)
		utils.ErrorCheck(err)
	}
	fmt.Println("正在下载主线剧情...")
	batchDownloadImage(cookiesParam, imagePageInfoList, safeTitle)
	fmt.Println("主线剧情下载完毕")
	fmt.Println("正在下载其他系列...")
	batchDownloadImage(cookiesParam, otherImagePageInfoList, otherPath)
	fmt.Println("其他系列下载完毕")
}
