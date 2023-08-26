package dmzj

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/spf13/cast"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cookiesPath       = `dmzj_cookies.json`
	numWorkers        = 5  //页面处理的并发量
	batchSize         = 10 //每次下载的图片页面数量，建议为numWorkers的整数倍
	maxImageInOnePage = 40 //单个图片页中最大图片数量
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

	//找到<div class="anim-main_list">
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

	rp := strings.NewReplacer("第", "", "话", "")
	lastChapter, _ := regexp.MatchString(`第(\d+)话`, galleryInfo.TagList["最新收录"][0])
	if lastChapter {
		galleryInfo.LastChapter = rp.Replace(galleryInfo.TagList["最新收录"][0])
	} else {
		galleryInfo.LastChapter = "未知"
	}
	return galleryInfo
}

// getAllImagePageInfo 从主目录页获取所有图片页地址，返回一个切片，元素为map[int]string，key为图片页序号，value为图片页地址
func getAllImagePageInfo(doc *goquery.Document) []map[int]string {
	var imagePageInfoList []map[int]string
	// 找到<div class="cartoon_online_border">
	doc.Find("div.cartoon_online_border").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				imageName := strings.TrimSpace(a.Text())
				//从图片页名字中提取图片页序号
				indexStr, err := utils.ExtractSubstringFromText(`(\d+)`, imageName)
				utils.ErrorCheck(err)
				//cast库在转换时字符串若是以 "0" 开头，"07" 转换后得到整型 7，而 "08" 转换后得到整型 0
				//https://iokde.com/post/golang-cast64-snare.html
				imageIndex, _ := strconv.Atoi(indexStr)

				imageInfo := map[int]string{
					imageIndex: "https://manhua.dmzj.com" + href,
				}
				imagePageInfoList = append(imagePageInfoList, imageInfo)
			}
		})
	})

	return imagePageInfoList
}

// getAllOtherImagePageInfo 从主目录页获取所有`其他系列`图片页地址，返回一个切片，元素为map[string]string，key为图片页名字，value为图片页地址
func getAllOtherImagePageInfo(doc *goquery.Document) []map[string]string {
	var imageOtherPageInfoList []map[string]string

	// 找到<div class="cartoon_online_border">
	doc.Find("div.cartoon_online_border_other").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				imageName := strings.TrimSpace(a.Text())
				imageInfo := map[string]string{
					imageName: "https://manhua.dmzj.com" + href,
				}
				imageOtherPageInfoList = append(imageOtherPageInfoList, imageInfo)
			}
		})
	})
	return imageOtherPageInfoList
}

// getImageUrlFromPage 从单个图片页获取图片地址
func getImageUrlFromPage(doc *goquery.Document) []string {
	var imageUrl []string
	//找到<div class="scrollbar-demo-item"
	doc.Find("div.scrollbar-demo-item").Each(func(i int, s *goquery.Selection) {
		s.Find("img").Each(func(j int, img *goquery.Selection) {
			src, exists := img.Attr("src")
			if exists {
				imageUrl = append(imageUrl, src)
			}
		})
	})
	return imageUrl
}

// syncParsePage 并发sync.WaitGroup解析页面，获取图片地址，并发量为numWorkers，返回实际获取的图片地址数量(int)
func syncParsePage(ImageInfoMapChannel <-chan map[int]string, imageInfoChannel chan<- map[string]string,
	cookiesParam []*network.CookieParam, numWorkers int) int {
	sumImage := 0
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for info := range ImageInfoMapChannel {
				for index, url := range info {
					//fmt.Println(index, url)
					pageDoc := client.GetHtmlDoc(cookiesParam, url)
					//获取图片地址
					imageUrlLists := getImageUrlFromPage(pageDoc)
					for i, imageUrl := range imageUrlLists {
						imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
						imageInfo := map[string]string{
							"imageName": cast.ToString(index) + "_" + cast.ToString(i) + imageSuffix,
							"imageUrl":  imageUrl,
						}
						imageInfoChannel <- imageInfo
						sumImage++
					}

				}
			}
		}()
	}

	wg.Wait() // 等待所有任务完成
	return sumImage
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

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) {
	beginIndex := 0
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

		needUpdate = lastGalleryInfo.LastChapter < galleryInfo.LastChapter
		if needUpdate {
			fmt.Println("发现新章节，更新下载记录")
			err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
			utils.ErrorCheck(err)
		} else {
			fmt.Println("无需更新下载记录")
		}
		imagePath, err := filepath.Abs(safeTitle)
		beginIndex = utils.GetBeginIndex(imagePath, []string{".jpg", ".png"})
	} else {
		//生成缓存文件
		err := utils.BuildCache(safeTitle, infoJsonPath, galleryInfo)
		utils.ErrorCheck(err)
		if onlyInfo {
			fmt.Println("画廊信息获取完毕，程序自动退出。")
			return
		}
	}
	fmt.Println("beginIndex=", beginIndex)
	imagePageInfoList := getAllImagePageInfo(menuDoc)
	//OtherImagePageInfoList := getAllOtherImagePageInfo(menuDoc)
	sortedImagePageInfoList := utils.SortListByMapsIntKey(imagePageInfoList, true)[beginIndex:]
	//sortedOtherImagePageInfoList := utils.SortListByMapsIntKey(OtherImagePageInfoList, true)[beginIndex:]

	//TODO:提个函数出来
	//按batchSize分组获取url并保存图片
	for batchIndex := 0; batchIndex < len(sortedImagePageInfoList); batchIndex += batchSize {
		//每次循环都重新初始化channel和切片
		subImagePageInfoList := sortedImagePageInfoList[batchIndex:utils.MinInt(batchIndex+batchSize, len(sortedImagePageInfoList))]
		imagePageInfoListChannel := make(chan map[int]string, len(subImagePageInfoList))
		//每个图片页中最多有maxImageInOnePage张图片
		imageInfoChannelSize := maxImageInOnePage * len(subImagePageInfoList)
		imageInfoChannel := make(chan map[string]string, imageInfoChannelSize)
		var imageInfoList []map[string]string

		for _, info := range subImagePageInfoList {
			imagePageInfoListChannel <- info
		}
		close(imagePageInfoListChannel)

		sumImage := syncParsePage(imagePageInfoListChannel, imageInfoChannel, cookiesParam, numWorkers)
		close(imageInfoChannel)
		//在这个channel里只有sumImage个元素，所以只需要循环sumImage次
		for i := 0; i < sumImage; i++ {
			imageInfo := <-imageInfoChannel
			imageInfoList = append(imageInfoList, imageInfo)
		}
		// 进行本次处理目录中所有图片的批量保存
		baseCollector := client.InitCollector(buildJpegRequestHeaders())
		err = utils.SaveImages(baseCollector, imageInfoList, safeTitle)
		utils.ErrorCheck(err)

		//防止被ban，每保存一组图片就sleep 5-15 seconds
		sleepTime := utils.TrueRandFloat(5, 15)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

}
