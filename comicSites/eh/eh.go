package eh

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/carlmjohnson/requests"
	"github.com/spf13/cast"
	"github.com/ybbus/httpretry"
	"log"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	imageInOnepage = 40
)

type GalleryInfo struct {
	URL        string              `json:"gallery_url"`
	Title      string              `json:"gallery_title"`
	TotalImage int                 `json:"total_image"`
	TagList    map[string][]string `json:"tag_list"`
}

func generateIndexURL(urlStr string, page int) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	if page == 0 {
		return u.String()
	}

	q := u.Query()
	q.Set("p", cast.ToString(page))
	u.RawQuery = q.Encode()

	return u.String()
}

func getGalleryInfo(c *http.Client, galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.TagList = make(map[string][]string)
	galleryInfo.URL = galleryUrl

	var buffer bytes.Buffer
	err := requests.URL(galleryUrl).
		Client(c).
		UserAgent(client.ChromeUserAgent).
		ToBytesBuffer(&buffer).
		Fetch(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(&buffer)
	if err != nil {
		log.Fatal(err)
	}
	galleryInfo.Title = doc.Find("h1#gn").Text()
	pageText := doc.Find("#gdd > table > tbody > tr:nth-child(6) > td.gdt2").Text()
	reMaxPage := regexp.MustCompile(`(\d+) pages`)
	if reMaxPage.MatchString(pageText) {
		//转换为int
		galleryInfo.TotalImage = cast.ToInt(reMaxPage.FindStringSubmatch(pageText)[1])
	}

	doc.Find("div#taglist table").Each(func(_ int, s *goquery.Selection) {
		s.Find("tr").Each(func(_ int, s *goquery.Selection) {
			key := strings.TrimSpace(s.Find("td.tc").Text())
			localKey := strings.ReplaceAll(key, ":", "")
			s.Find("td div").Each(func(_ int, s *goquery.Selection) {
				value := strings.TrimSpace(s.Text())
				galleryInfo.TagList[localKey] = append(galleryInfo.TagList[localKey], value)
			})
		})
	})

	return galleryInfo
}

func getImagePageUrlList(c *http.Client, indexUrl string) []string {
	var imagePageUrls []string
	var buffer bytes.Buffer
	err := requests.
		URL(indexUrl).
		Client(c).
		UserAgent(client.ChromeUserAgent).
		ToBytesBuffer(&buffer).
		Fetch(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(&buffer)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div#gdt div.gdtm a").Each(func(_ int, s *goquery.Selection) {
		imgUrl, _ := s.Attr("href")
		imagePageUrls = append(imagePageUrls, imgUrl)
	})

	return imagePageUrls
}

func getImageUrl(c *http.Client, imagePageUrl string) string {
	var imageUrl string
	var buffer bytes.Buffer
	err := requests.
		URL(imagePageUrl).
		Client(c).
		UserAgent(client.ChromeUserAgent).
		ToBytesBuffer(&buffer).
		Fetch(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(&buffer)
	if err != nil {
		log.Fatal(err)
	}
	imageUrl, _ = doc.Find("img#img").Attr("src")
	return imageUrl
}

func buildJPEGRequestHeaders() http.Header {
	return http.Header{
		"Accept":             {"image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8"},
		"Accept-Encoding":    {"gzip, deflate, br"},
		"Accept-Language":    {"zh-CN,zh;q=0.9"},
		"Connection":         {"keep-alive"},
		"Dnt":                {"1"},
		"Host":               {"dqoaprm.qgankvrkxxiw.hath.network"},
		"Referer":            {"https://e-hentai.org/"},
		"Sec-Ch-Ua":          {"\"Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"115\", \"Chromium\";v=\"115\""},
		"Sec-Ch-Ua-Mobile":   {"?0"},
		"Sec-Ch-Ua-Platform": {"\"Windows\""},
		"Sec-Fetch-Dest":     {"image"},
		"Sec-Fetch-Mode":     {"no-cors"},
		"Sec-Fetch-Site":     {"cross-site"},
		"Sec-Gpc":            {"1"},
		"User-Agent":         {client.ChromeUserAgent},
	}
}

func getImageInfoFromPage(c *http.Client, imagePageUrl string) (string, string) {
	imageIndex := imagePageUrl[strings.LastIndex(imagePageUrl, "-")+1:]
	//imageUrl := getImageUrl(c, imagePageUrl)
	imageUrl := getImageUrl(c, imagePageUrl)
	imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
	imageTitle := fmt.Sprintf("%s%s", imageIndex, imageSuffix)
	return imageTitle, imageUrl
}

func DownloadGallery(infoJsonPath string, galleryUrl string, onlyInfo bool) error {
	//目录号
	beginIndex := 0
	//余数
	remainder := 0

	// create a new http client with retry
	c := httpretry.NewDefaultClient(
		// retry up to 5 times
		httpretry.WithMaxRetryCount(5),
		// retry on status >= 500, if err != nil, or if response was nil (status == 0)
		httpretry.WithRetryPolicy(func(statusCode int, err error) bool {
			return err != nil || statusCode >= 500 || statusCode == 0
		}),
		// every retry should wait one more second
		httpretry.WithBackoffPolicy(func(attemptNum int) time.Duration {
			return time.Duration(attemptNum+1) * 1 * time.Second
		}),
	)

	//获取画廊信息
	galleryInfo := getGalleryInfo(c, galleryUrl)
	fmt.Println("Total Image:", galleryInfo.TotalImage)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.FileExists(filepath.Join(safeTitle, infoJsonPath)) {
		fmt.Println("发现下载记录")
		//获取已经下载的图片数量
		downloadedImageCount := utils.GetFileTotal(safeTitle, []string{".jpg", ".png"})
		fmt.Println("Downloaded image count:", downloadedImageCount)
		//计算剩余图片数量
		remainImageCount := galleryInfo.TotalImage - downloadedImageCount
		if remainImageCount == 0 {
			fmt.Println("本gallery已经下载完毕")
			return nil
		} else if remainImageCount < 0 {
			return fmt.Errorf("下载记录有误！")
		} else {
			fmt.Println("剩余图片数量:", remainImageCount)
			beginIndex = int(math.Floor(float64(downloadedImageCount) / float64(imageInOnepage)))
			remainder = downloadedImageCount - imageInOnepage*beginIndex
		}
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
	//重新初始化Collector

	sumPage := int(math.Ceil(float64(galleryInfo.TotalImage) / float64(imageInOnepage)))
	for i := beginIndex; i < sumPage; i++ {
		fmt.Println("\nCurrent index:", i)
		indexUrl := generateIndexURL(galleryUrl, i)
		fmt.Println(indexUrl)
		var imagePageUrlList []string
		//imagePageUrlList = getImagePageUrlList(collector, indexUrl)
		imagePageUrlList = getImagePageUrlList(c, indexUrl)
		if i == beginIndex {
			//如果是第一次处理目录，需要去掉前面的余数
			imagePageUrlList = imagePageUrlList[remainder:]
		}

		var imageInfoList []utils.ImageInfo
		//根据imagePageUrls获取imageDataList
		for _, imagePageUrl := range imagePageUrlList {
			imageTitle, imageUrl := getImageInfoFromPage(c, imagePageUrl)
			imageInfoList = append(imageInfoList, utils.ImageInfo{
				Title: imageTitle,
				Url:   imageUrl,
			})
		}
		//防止被ban，每处理一篇目录就sleep 5-10 seconds
		sleepTime := client.TrueRandFloat(5, 10)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)

		// 进行本次处理目录中所有图片的批量保存
		//utils.SaveImagesNew(collector, imageInfoList, safeTitle)
		utils.SaveImagesWithRequest(c, buildJPEGRequestHeaders(), imageInfoList, safeTitle)

		//防止被ban，每保存一篇目录中的所有图片就sleep 5-15 seconds
		sleepTime = client.TrueRandFloat(5, 15)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return nil
}
