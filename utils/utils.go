package utils

import (
	"ComicCrawler/client"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/gocolly/colly/v2"
	"github.com/smallnest/chanx"
	"github.com/spf13/cast"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	NumWorkers        = 5  //页面处理的并发量
	BatchSize         = 10 //每次下载的图片页面数量，建议为numWorkers的整数倍
	MaxImageInOnePage = 30 //单个图片页中最大图片数量，用于初始化imageInfoChannelSize，设置一个合适的数量可以减少无限有缓冲channel的扩容消耗
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

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
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

// BuildCache 用于生成utf-8格式的缓存文件 data为待写入数据结构
func BuildCache(saveDir string, cacheFile string, data interface{}) error {
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
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		return err
	}

	return nil
}

// LoadCache 用于加载utf-8格式的缓存文件 result是一个指向目标数据结构的指针
func LoadCache(filePath string, result interface{}) error {
	// 打开utf-8格式的文件用于读取数据
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File open error:", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)
	// 设置解码器的输入流为 UTF-8
	err = decoder.Decode(result)
	if err != nil {
		return err
	}
	return nil
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

// GetFileTotal 用于获取指定目录下指定后缀的文件数量
func GetFileTotal(dirPath string, fileSuffixes []string) int {
	var count int // 用于存储文件数量的变量

	// 使用Walk函数遍历指定目录及其子目录中的所有文件和文件夹
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查是否为文件
		if !info.IsDir() {
			// 获取文件的扩展名
			ext := filepath.Ext(path)
			// 将扩展名转换为小写，以便比较
			ext = strings.ToLower(ext)
			// 检查文件扩展名是否在指定的后缀列表中
			for _, suffix := range fileSuffixes {
				if ext == suffix {
					count++
					break // 找到匹配的后缀，停止循环
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("遍历目录出错:", err)
	}

	return count
}

// GetBeginIndex 用于获取指定目录下指定格式和后缀的文件中最大的序号，用于计算剩余图片数（目前只支持`数字_数字.后缀`的格式）
func GetBeginIndex(dirPath string, fileSuffixes []string) int {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return 0
	}

	maxIndex := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		for _, suffix := range fileSuffixes {
			if strings.HasSuffix(file.Name(), suffix) {
				name := strings.TrimSuffix(file.Name(), suffix)
				parts := strings.Split(name, "_")
				if len(parts) != 2 {
					continue
				}

				index, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}

				if index > maxIndex {
					maxIndex = index
				}
			}
		}
	}

	return maxIndex
}

// ReadListFile 用于按行读取列表文件，返回一个字符串切片
func ReadListFile(filePath string) ([]string, error) {
	var list []string
	file, err := os.Open(filePath)
	if err != nil {
		return list, err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	var line string
	for {
		_, err := fmt.Fscanln(file, &line)
		if err != nil {
			break
		}
		list = append(list, line)
	}
	return list, nil
}

// SaveImages 保存imageInfoList中的所有图片，imageInfoMap中的每个元素都是一个map，包含两个键值对，imageTitle:title和imageUrl:url
func SaveImages(JPEGCollector *colly.Collector, imageInfoList []map[string]string, saveDir string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	ErrorCheck(err)

	var imageContent []byte

	JPEGCollector.OnResponse(func(r *colly.Response) {
		imageContent = r.Body
	})

	for _, data := range imageInfoList {
		imageTitle := data["imageTitle"]
		imageUrl := data["imageUrl"]
		filePath, err := filepath.Abs(filepath.Join(dir, imageTitle))
		ErrorCheck(err)
		err = JPEGCollector.Request("GET", imageUrl, nil, nil, nil)
		ErrorCheck(err)
		//增加延时，防止被ban
		time.Sleep(time.Millisecond * time.Duration(client.DelayMs))
		err = SaveFile(filePath, imageContent)
		if err != nil {
			fmt.Println("Error saving image:", err)
		} else {
			fmt.Println("Image saved:", filePath)
		}
	}

	return nil
}

// ExtractSubstringFromText 按照Pattern在text里匹配，找到了就返回匹配到的部分
func ExtractSubstringFromText(pattern string, text string) (string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	match := regex.FindStringSubmatch(text)
	if match != nil {
		number := match[1]
		return number, nil
	} else {
		return "", fmt.Errorf("在pattern中未找到匹配的数字")
	}
}

func CheckUpdate(lastUpdateTime string, newTime string) bool {
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

// ElementInSlice 判断slice中是否存在某个item
func ElementInSlice(value interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(value, s.Index(i).Interface()) {
				return true
			}
		}
	}
	return false
}

// SyncParsePage 并发sync.WaitGroup，通过chromedp解析页面，获取图片地址，并发量为numWorkers，返回实际获取的图片地址数量(int)
// localGetImageUrlFromPage为不同软件包的内部函数，用于从页面中获取图片地址
// localGetPage为client的内部函数，用于在不同情况下获取页面内容
// chromeCtx为chromedp的上下文
func SyncParsePage(
	localGetImageUrlListFromPage func(*goquery.Document) []string,
	localGetPage func(context.Context, []*network.CookieParam, string) []byte,
	chromeCtxChannel chan context.Context, //FIXME:这个地方应该设置成一个固定大小的chromeCtx池，而不是slice
	imagePageInfoListChannel <-chan map[int]string, imageInfoListChannel *chanx.UnboundedChan[map[string]string],
	cookiesParam []*network.CookieParam) {

	var wg sync.WaitGroup
	//WaitGroup 使用计数器来工作。当创建 WaitGroup 时，其计数器初始值为 0
	//当调用 Add 方法时，计数器增加 1，当调用 Done 方法时，计数器减少 1。当调用 Wait 方法时，goroutine 将会阻塞，直至计数器数值为 0。
	for i := 0; i < NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//每个goroutine从channel中取出一个map，map中包含了图片页的序号和url
			//若channel非空
			if info, ok := <-imagePageInfoListChannel; ok {
				for index, url := range info {
					//fmt.Println(index, url)
					pageDoc := func() *goquery.Document {
						chromeCtx := <-chromeCtxChannel
						timeoutCtx, cancel := context.WithTimeout(chromeCtx, 120*time.Second)
						defer cancel()
						pageDoc := client.GetHtmlDoc(localGetPage(timeoutCtx, cookiesParam, url))
						//pageDoc := client.GetHtmlDoc(localGetPage(chromeCtx, cookiesParam, url))
						chromeCtxChannel <- chromeCtx
						return pageDoc
					}()

					//获取图片地址
					imageUrlList := localGetImageUrlListFromPage(pageDoc)
					for k, imageUrl := range imageUrlList {
						imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
						imageInfo := map[string]string{
							"imageTitle": cast.ToString(index) + "_" + cast.ToString(k) + imageSuffix,
							"imageUrl":   imageUrl,
						}
						imageInfoListChannel.In <- imageInfo
					}
				}
			}
		}()

	}

	wg.Wait() // 等待所有任务完成
}

// BatchDownloadImage 按batchSize分组渲染页面，获取图片url并保存图片
func BatchDownloadImage(
	localGetImageUrlListFromPage func(*goquery.Document) []string,
	localBuildJPEGRequestHeaders func() http.Header,
	localGetPage func(context.Context, []*network.CookieParam, string) []byte,
	cookiesParam []*network.CookieParam, imagePageInfoList []map[int]string, saveDir string) {

	//syncParsePage使用的channel
	chromeCtxChannel := make(chan context.Context, NumWorkers)
	//使用结束后统一手动关闭ctx
	var cancelList []context.CancelFunc
	for i := 0; i < NumWorkers; i++ {
		chromeCtx, cancel := client.InitChromedpContext(true)
		chromeCtxChannel <- chromeCtx
		cancelList = append(cancelList, cancel)
	}

	for batchIndex := 0; batchIndex < len(imagePageInfoList); batchIndex += BatchSize {
		//每次循环都重新初始化channel和切片
		subImagePageInfoList := imagePageInfoList[batchIndex:MinInt(batchIndex+BatchSize, len(imagePageInfoList))]
		imagePageInfoListChannel := make(chan map[int]string, len(subImagePageInfoList))
		//每个图片页中最多有maxImageInOnePage张图片
		imageInfoChannelSize := MaxImageInOnePage * len(subImagePageInfoList)
		//创建一个无限大小有缓冲channel，用于存储所有图片的信息
		imageInfoListChannel := chanx.NewUnboundedChan[map[string]string](imageInfoChannelSize)
		var imageInfoList []map[string]string

		for _, info := range subImagePageInfoList {
			imagePageInfoListChannel <- info
		}
		close(imagePageInfoListChannel)

		SyncParsePage(localGetImageUrlListFromPage, localGetPage,
			chromeCtxChannel, imagePageInfoListChannel, imageInfoListChannel, cookiesParam)
		close(imageInfoListChannel.In)

		for imageInfo := range imageInfoListChannel.Out {
			imageInfoList = append(imageInfoList, imageInfo)
		}
		// 进行本次处理目录中所有图片的批量保存
		baseCollector := client.InitJPEGCollector(localBuildJPEGRequestHeaders())
		err := SaveImages(baseCollector, imageInfoList, saveDir)
		ErrorCheck(err)

		//防止被ban，每保存一组图片就sleep 5-15 seconds
		sleepTime := client.TrueRandFloat(5, 15)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	for _, cancel := range cancelList {
		cancel()
	}
}
