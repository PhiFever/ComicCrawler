package main

import (
	"EH_downloader/client"
	"EH_downloader/utils"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func getGalleryInfo(c *colly.Collector, galleryUrl string) (string, int) {
	var title string
	var sumImage int
	//找到<h1 id="gn">标签,即为文章标题
	c.OnHTML("h1#gn", func(e *colly.HTMLElement) {
		title = e.Text
	})

	//找到<td class="gdt2">标签
	reMaxPage := regexp.MustCompile(`(\d+) pages`)
	c.OnHTML("td.gdt2", func(e *colly.HTMLElement) {
		if reMaxPage.MatchString(e.Text) {
			//转换为int
			sumImage, _ = cast.ToIntE(reMaxPage.FindStringSubmatch(e.Text)[1])
		}
	})
	err := c.Visit(galleryUrl)
	if err != nil {
		log.Fatal(err)
		return "", 0
	}
	return title, sumImage
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

// 获取图片页面的url
func getAllImagePageUrl(c *colly.Collector, indexUrl string) []string {
	var imagePageUrls []string
	c.OnHTML("div[id='gdt']", func(e *colly.HTMLElement) {
		//找到其下所有<div class="gdtm">标签
		e.ForEach("div.gdtm", func(_ int, el *colly.HTMLElement) {
			//找到<a href="xxx">标签，只有一个
			imgUrl := el.ChildAttr("a", "href")
			imagePageUrls = append(imagePageUrls, imgUrl)
		})
	})
	err := c.Visit(indexUrl)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return imagePageUrls
}

func getImageUrl(c *colly.Collector, imagePageUrl string) string {
	//id="img"的src属性
	var imageUrl string
	c.OnHTML("img[id='img']", func(e *colly.HTMLElement) {
		imageUrl = e.Attr("src")
	})
	err := c.Visit(imagePageUrl)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return imageUrl
}

// saveImages 保存imageDataList中的所有图片，imageDataList中的每个元素都是一个map，包含两个键值对，imageName和imageUrl
func saveImages(baseCollector *colly.Collector, imageDataList []map[string]string, saveDir string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	utils.ErrorCheck(err)

	i := 0
	baseCollector.OnResponse(func(r *colly.Response) {
		imageName := imageDataList[i]["imageName"]
		//imageUrl := imageDataList[i]["imageUrl"]
		filePath, err := filepath.Abs(filepath.Join(dir, imageName))
		utils.ErrorCheck(err)
		err = utils.SaveFile(filePath, r.Body)
		if err != nil {
			fmt.Println("Error saving image:", err)
		} else {
			fmt.Println("Image saved:", filePath)
		}
		//fmt.Println("Current value:", i)
		if i < len(imageDataList)-1 {
			i++
			err = baseCollector.Request("GET", imageDataList[i]["imageUrl"], nil, nil, nil)
			utils.ErrorCheck(err)
		}
	})

	err = baseCollector.Request("GET", imageDataList[i]["imageUrl"], nil, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func buildImageInfo(c *colly.Collector, imagePageUrl string) (string, string) {
	imageIndex := imagePageUrl[strings.LastIndex(imagePageUrl, "-")+1:]
	imageUrl := getImageUrl(c, imagePageUrl)
	imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
	imageName := fmt.Sprintf("%s%s", imageIndex, imageSuffix)
	return imageName, imageUrl
}

func main() {
	galleryUrl := "https://e-hentai.org/g/1838806/61460acecb/"
	imageInOnepage := 40
	beginIndex := 0
	cacheFile := `cache.json`

	//记录开始时间
	startTime := time.Now()
	collector := colly.NewCollector(
		//模拟浏览器
		colly.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 Edg/115.0.1901.203`),
	)

	title, sumImage := getGalleryInfo(collector, galleryUrl)
	if title == "" || sumImage == 0 {
		log.Fatal("Invalid gallery url")
	}

	fmt.Println("Article Title:", title)
	fmt.Println("Sum Image:", sumImage)

	safeTitle := utils.ToSafeFilename(title)
	fmt.Println(safeTitle)

	//创建map{'imageName':imageName,'imageUrl':imageUrl}
	var imageDataList []map[string]string
	//cachePath := filepath.Join(safeTitle, cacheFile)

	//测试用
	cachePath := filepath.Join("test", cacheFile)

	//重新初始化Collector
	headers := make(http.Header)
	headers.Set(`User-Agent`, `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.82`)
	headers.Set("Accept", "image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	headers.Set("Upgrade-Insecure-Requests", "1")
	collector = client.InitCollector(headers)

	if utils.CacheFileExists(cachePath) {
		fmt.Println("Cache file exists")
		imageDataList, _ = utils.LoadCache(cachePath)
	} else {
		sumPage := int(math.Ceil(float64(sumImage) / float64(imageInOnepage)))
		for i := beginIndex; i < sumPage; i++ {
			fmt.Println("Current value:", i)
			indexUrl := generateIndexURL(galleryUrl, i)
			fmt.Println(indexUrl)
			imagePageUrls := getAllImagePageUrl(collector, indexUrl)
			for _, imagePageUrl := range imagePageUrls {
				imageName, imageUrl := buildImageInfo(collector, imagePageUrl)
				imageDataList = append(imageDataList, map[string]string{
					"imageName": imageName,
					"imageUrl":  imageUrl,
				})
			}
		}

		err := utils.BuildCache(safeTitle, cacheFile, imageDataList)
		utils.ErrorCheck(err)
	}

	////测试用
	//err := buildCache(`./`, cacheFile, imageDataList)
	//if err != nil {
	//	return
	//}

	// 进行图片批量保存
	err := saveImages(collector, imageDataList, safeTitle)
	utils.ErrorCheck(err)

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	fmt.Println("执行时间：", endTime.Sub(startTime).Seconds(), "秒")
}
