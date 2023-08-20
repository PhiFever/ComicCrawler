package main

import (
	"EH_downloader/client"
	"EH_downloader/eh"
	"EH_downloader/utils"
	"flag"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	galleryUrl string
	onlyInfo   bool
	listFile   string
)

// saveImages 保存imageDataList中的所有图片，imageDataList中的每个元素都是一个map，包含两个键值对，imageName和imageUrl
func saveImages(baseCollector *colly.Collector, imageDataList []map[string]string, saveDir string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	utils.ErrorCheck(err)

	var imageContent []byte

	baseCollector.OnResponse(func(r *colly.Response) {
		imageContent = r.Body
	})

	for _, data := range imageDataList {
		imageName := data["imageName"]
		imageUrl := data["imageUrl"]
		filePath, err := filepath.Abs(filepath.Join(dir, imageName))
		utils.ErrorCheck(err)
		err = baseCollector.Request("GET", imageUrl, nil, nil, nil)
		utils.ErrorCheck(err)
		err = utils.SaveFile(filePath, imageContent)
		if err != nil {
			fmt.Println("Error saving image:", err)
		} else {
			fmt.Println("Image saved:", filePath)
		}
	}

	return nil
}

func buildImageInfo(c *colly.Collector, imagePageUrl string) (string, string) {
	imageIndex := imagePageUrl[strings.LastIndex(imagePageUrl, "-")+1:]
	imageUrl := eh.GetImageUrl(c, imagePageUrl)
	imageSuffix := imageUrl[strings.LastIndex(imageUrl, "."):]
	imageName := fmt.Sprintf("%s%s", imageIndex, imageSuffix)
	return imageName, imageUrl
}

func initArgsParse() {
	flag.StringVar(&galleryUrl, "url", "", "待下载的画廊地址（必填）")
	flag.StringVar(&galleryUrl, "u", "", "待下载的画廊地址（必填）")
	flag.BoolVar(&onlyInfo, "info", false, "只获取画廊信息(true/false)，默认为false")
	flag.BoolVar(&onlyInfo, "i", false, "只获取画廊信息(true/false)，默认为false")
	flag.StringVar(&listFile, "list", "", "待下载的画廊地址列表文件")
	flag.StringVar(&listFile, "l", "", "待下载的画廊地址列表文件")
}

func main() {
	//待配置的参数
	const imageInOnepage = 40
	const cacheFile = "galleryInfo.json"

	initArgsParse()
	flag.Parse()
	if galleryUrl == "" && listFile == "" {
		fmt.Println("本程序为命令行程序，请在命令行中运行参数-h以查看帮助")
		return
	}

	beginIndex := 0

	//记录开始时间
	startTime := time.Now()

	collector := colly.NewCollector(
		//模拟浏览器
		colly.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 Edg/115.0.1901.203`),
	)

	//获取画廊信息
	galleryInfo := eh.GetGalleryInfo(collector, galleryUrl)
	fmt.Println("Total Image:", galleryInfo.TotalImage)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)

	if utils.CacheFileExists(filepath.Join(safeTitle, cacheFile)) {
		fmt.Println("发现下载记录")
		//获取已经下载的图片数量
		downloadedImageCount := utils.GetFileTotal(safeTitle, "jpg")
		fmt.Println("Downloaded image count:", downloadedImageCount)
		//计算剩余图片数量
		remainImageCount := galleryInfo.TotalImage - downloadedImageCount
		if remainImageCount == 0 {
			fmt.Println("所有图片已下载完成！程序自动退出。")
			return
		} else if remainImageCount < 0 {
			fmt.Println("下载记录有误！")
			return
		} else {
			fmt.Println("剩余图片数量:", remainImageCount)
			beginIndex = int(math.Floor(float64(downloadedImageCount) / float64(imageInOnepage)))
		}
	} else {
		//生成缓存文件
		err := utils.BuildCache(safeTitle, cacheFile, galleryInfo)
		utils.ErrorCheck(err)
		if onlyInfo {
			fmt.Println("画廊信息获取完毕，程序自动退出。")
			return
		}
	}

	//创建map{'imageName':imageName,'imageUrl':imageUrl}
	var imageDataList []map[string]string

	//重新初始化Collector
	collector = client.InitCollector(eh.BuildJpegRequestHeaders())

	sumPage := int(math.Ceil(float64(galleryInfo.TotalImage) / float64(imageInOnepage)))
	for i := beginIndex; i < sumPage; i++ {
		fmt.Println("\nCurrent index:", i)
		indexUrl := eh.GenerateIndexURL(galleryUrl, i)
		fmt.Println(indexUrl)
		imagePageUrls := eh.GetAllImagePageUrl(collector, indexUrl)

		//清空imageDataList中的数据
		imageDataList = []map[string]string{}

		//根据imagePageUrls获取imageDataList
		for _, imagePageUrl := range imagePageUrls {
			imageName, imageUrl := buildImageInfo(collector, imagePageUrl)
			imageDataList = append(imageDataList, map[string]string{
				"imageName": imageName,
				"imageUrl":  imageUrl,
			})
		}
		//防止被ban，每处理一篇目录就sleep 5-10 seconds
		sleepTime := utils.TrueRandFloat(5, 10)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)

		// 进行本次处理目录中所有图片的批量保存
		err := saveImages(collector, imageDataList, safeTitle)
		utils.ErrorCheck(err)

		//防止被ban，每保存一篇目录就sleep 5-10 seconds
		sleepTime = utils.TrueRandFloat(5, 10)
		log.Println("Sleep ", cast.ToString(sleepTime), " seconds...")
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	fmt.Println("执行时间：", endTime.Sub(startTime).Seconds(), "秒")
}
