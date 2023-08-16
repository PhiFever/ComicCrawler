package main

import (
	"encoding/json"
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

func ErrorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ToSafeFilename(in string) string {
	//https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
	//全部替换为_
	rp := strings.NewReplacer(
		"/", "_",
		`\`, "_",
		"<", "_",
		">", "_",
		":", "_",
		`"`, "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	rt := rp.Replace(in)
	return rt
}

func initCollector() *colly.Collector {
	c := colly.NewCollector(
		//这次在colly.NewCollector里面加了一项colly.Async(true)，表示抓取时异步的
		//colly.Async(true),
		//允许重复访问
		//colly.AllowURLRevisit(),
		//模拟浏览器
		colly.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.82`),
	)
	//限制采集规格
	rule := &colly.LimitRule{
		RandomDelay: 5 * time.Second,
		Parallelism: 5, //并发数
	}
	_ = c.Limit(rule)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Fatal("Something went wrong:", err)
	})

	//c.OnResponse(func(r *colly.Response) {
	//	log.Println("Visited", r.Request.URL)
	//	//fmt.Println(string(r.Body))
	//})
	return c
}

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

// SaveImages 保存imageDataList中的所有图片，imageDataList中的每个元素都是一个map，包含两个键值对，imageName和imageUrl
func SaveImages(baseCollector *colly.Collector, imageDataList []map[string]string, saveDir string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	ErrorCheck(err)

	// 定义回调函数
	onResponse := func(imagePath, imageName string) func(r *colly.Response) {
		return func(r *colly.Response) {
			filePath, err := filepath.Abs(filepath.Join(imagePath, imageName))
			ErrorCheck(err)
			err = SaveFile(filePath, r.Body)
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

// SaveFile 用于保存文件
func SaveFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	//fmt.Println(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// buildCache 用于生成utf-8格式的图片列表缓存文件
func buildCache(saveDir string, cacheFile string, imageDataList []map[string]string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	ErrorCheck(err)

	// 打开文件用于写入数据
	file, err := os.Create(filepath.Join(dir, cacheFile))
	if err != nil {
		fmt.Println("File creation error:", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	// 创建 JSON 编码器，并指定输出流为文件
	encoder := json.NewEncoder(file)
	// 设置编码器的输出流为 UTF-8
	encoder.SetIndent("", "    ") // 设置缩进，可选
	encoder.SetEscapeHTML(false)  // 禁用转义 HTML
	err = encoder.Encode(imageDataList)
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		return err
	}

	return nil
}

func loadCache(filePath string) ([]map[string]string, error) {
	var imageDataList []map[string]string
	// 打开utf-8文件用于读取数据
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File open error:", err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)
	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)
	// 设置解码器的输入流为 UTF-8
	err = decoder.Decode(&imageDataList)
	if err != nil {
		return nil, err
	}
	return imageDataList, nil
}

func buildImageInfo(imagePageUrl string) (string, string) {
	imageIndex := imagePageUrl[strings.LastIndex(imagePageUrl, "-")+1:]
	imageUrl := getImageUrl(initCollector(), imagePageUrl)
	imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
	imageName := fmt.Sprintf("%s%s", imageIndex, imageSuffix)
	return imageName, imageUrl
}

func cacheFileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
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

	safeTitle := ToSafeFilename(title)
	fmt.Println(safeTitle)

	//创建map{'imageName':imageName,'imageUrl':imageUrl}
	var imageDataList []map[string]string
	cachePath := filepath.Join(safeTitle, cacheFile)
	if cacheFileExists(cachePath) {
		imageDataList, _ = loadCache(cachePath)
	} else {
		//重新初始化Collector
		c = initCollector()
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

		err := buildCache(safeTitle, cacheFile, imageDataList)
		ErrorCheck(err)
	}

	////测试用
	//err := buildCache(`./`, cacheFile, imageDataList)
	//if err != nil {
	//	return
	//}

	// 进行图片批量保存
	err := SaveImages(c, imageDataList, safeTitle)
	ErrorCheck(err)

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	fmt.Println("执行时间：", endTime.Sub(startTime).Seconds(), "秒")
}
