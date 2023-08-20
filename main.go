package main

import (
	"EH_downloader/eh"
	"EH_downloader/utils"
	"flag"
	"fmt"
	"log"
	"time"
)

var (
	galleryUrl string
	onlyInfo   bool
	listFile   string
)

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

	var galleryUrlList []string

	switch {
	case galleryUrl == "" && listFile == "":
		fmt.Println("本程序为命令行程序，请在命令行中运行参数-h以查看帮助")
		return
	case galleryUrl != "" && listFile != "":
		fmt.Println("参数错误，请在命令行中运行参数-h以查看帮助")
		return
	case listFile != "":
		UrlList, err := utils.ReadListFile(listFile)
		utils.ErrorCheck(err)
		//UrlList... 使用了展开操作符（...），将 UrlList 切片中的所有元素一个一个地展开，作为参数传递给 append 函数
		galleryUrlList = append(galleryUrlList, UrlList...)
	case galleryUrl != "":
		galleryUrlList = append(galleryUrlList, galleryUrl)
	default:
		log.Fatal("未知错误")
	}

	//记录开始时间
	startTime := time.Now()

	for _, url := range galleryUrlList {
		fmt.Println("Current gallery:", url)
		eh.DownloadGallery(cacheFile, imageInOnepage, url, onlyInfo)
	}

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	fmt.Println("程序结束，执行时间：", endTime.Sub(startTime).Seconds(), "秒")
}
