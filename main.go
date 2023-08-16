package main

import (
	"EH_downloader/eh"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
	eh.ErrorCheck(err)

	// 定义回调函数
	onResponse := func(imagePath, imageName string) func(r *colly.Response) {
		return func(r *colly.Response) {
			filePath, err := filepath.Abs(filepath.Join(imagePath, imageName))
			eh.ErrorCheck(err)
			err = eh.SaveFile(filePath, r.Body)
			if err != nil {
				fmt.Println("Error saving image:", err)
			} else {
				fmt.Println("Image saved:", filePath)
			}
		}
	}

	// 创建一个 WaitGroup，以便等待所有 goroutines 完成
	var wg sync.WaitGroup
	for _, imageData := range imageDataList {
		wg.Add(1)

		go func(data map[string]string) {
			defer wg.Done()
			// 为每张图片创建一个 Collector
			c := baseCollector.Clone()
			// 设置回调函数
			c.OnResponse(onResponse(dir, data["imageName"]))
			// 使用 Colly 发起请求并保存图片
			err := c.Visit(data["imageUrl"])
			if err != nil {
				fmt.Println("Error visiting URL:", err)
			}
		}(imageData)
	}

	// 等待所有 goroutines 完成
	wg.Wait()

	return nil
}

func buildImageInfo(imagePageUrl string) (string, string) {
	imageIndex := imagePageUrl[strings.LastIndex(imagePageUrl, "-")+1:]
	imageUrl := getImageUrl(eh.InitCollector(), imagePageUrl)
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
	c := colly.NewCollector(
		//模拟浏览器
		colly.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 Edg/115.0.1901.203`),
	)

	title, sumImage := getGalleryInfo(c, galleryUrl)
	if title == "" || sumImage == 0 {
		log.Fatal("Invalid gallery url")
	}

	fmt.Println("Article Title:", title)
	fmt.Println("Sum Image:", sumImage)

	safeTitle := eh.ToSafeFilename(title)
	fmt.Println(safeTitle)

	//创建map{'imageName':imageName,'imageUrl':imageUrl}
	var imageDataList []map[string]string
	cachePath := filepath.Join(safeTitle, cacheFile)
	if eh.CacheFileExists(cachePath) {
		imageDataList, _ = eh.LoadCache(cachePath)
	} else {
		//重新初始化Collector
		c = eh.InitCollector()
		sumPage := int(math.Ceil(float64(sumImage) / float64(imageInOnepage)))
		for i := beginIndex; i < sumPage; i++ {
			fmt.Println("Current value:", i)
			indexUrl := generateIndexURL(galleryUrl, i)
			fmt.Println(indexUrl)
			imagePageUrls := getAllImagePageUrl(c, indexUrl)
			for _, imagePageUrl := range imagePageUrls {
				imageName, imageUrl := buildImageInfo(imagePageUrl)
				imageDataList = append(imageDataList, map[string]string{
					"imageName": imageName,
					"imageUrl":  imageUrl,
				})
			}
		}

		err := eh.BuildCache(safeTitle, cacheFile, imageDataList)
		eh.ErrorCheck(err)
	}

	////测试用
	//err := buildCache(`./`, cacheFile, imageDataList)
	//if err != nil {
	//	return
	//}

	// 进行图片批量保存
	err := saveImages(c, imageDataList, safeTitle)
	eh.ErrorCheck(err)

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	fmt.Println("执行时间：", endTime.Sub(startTime).Seconds(), "秒")
}
