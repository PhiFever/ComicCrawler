package main

import (
	"ComicCrawler/dmzj"
	"ComicCrawler/eh"
	"ComicCrawler/utils"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"regexp"
	"time"
)

const infoJsonPath = "galleryInfo.json"

var (
	galleryUrl string
	onlyInfo   bool
	listFile   string
	buildTime  string
	goVersion  string
	version    = "v1.0.0"
)

type GalleryDownloader struct{}

func (gd GalleryDownloader) Download(infoJson string, url string, onlyInfo bool) error {
	// 根据正则表达式判断是哪个软件包的gallery，并调用相应的下载函数
	if matched, _ := regexp.MatchString(`^https://e-hentai.org/g/[a-z0-9]{7}/[a-z0-9]{10}/$`, url); matched {
		//fmt.Println("调用eh包的DownloadGallery函数")
		eh.DownloadGallery(infoJson, url, onlyInfo)
	} else if matched, _ := regexp.MatchString(`^https://manhua.dmzj.com/[a-z0-9]*/$`, url); matched {
		//fmt.Println("调用dmzj包的DownloadGallery函数")
		dmzj.DownloadGallery(infoJson, url, onlyInfo)
	} else {
		return fmt.Errorf("未知的url格式：%s", url)
	}
	return nil
}

func initArgsParse() {
	flag.StringVar(&galleryUrl, "url", "", "待下载的画廊地址（必填）")
	flag.StringVar(&galleryUrl, "u", "", "待下载的画廊地址（必填）")
	flag.BoolVar(&onlyInfo, "info", false, "只获取画廊信息(true/false)，默认为false")
	flag.BoolVar(&onlyInfo, "i", false, "只获取画廊信息(true/false)，默认为false")
	flag.StringVar(&listFile, "list", "", "待下载的画廊地址列表文件，每行一个url。(不能与参数-url同时使用)")
	flag.StringVar(&listFile, "l", "", "待下载的画廊地址列表文件，每行一个url。(不能与参数-url同时使用)")
}

func main() {
	//版本信息
	args := os.Args
	if len(args) == 2 && (args[1] == "--version" || args[1] == "-v") {
		fmt.Printf("Version: %s \n", version)
		fmt.Printf("Build TimeStamp: %s \n", buildTime)
		fmt.Printf("GoLang Version: %s \n", goVersion)
		os.Exit(0)
	}

	initArgsParse()
	flag.Parse()

	var galleryUrlList []string

	switch {
	case galleryUrl == "" && listFile == "":
		fmt.Println("本程序为命令行程序，请在命令行中运行参数-h以查看帮助")
		os.Exit(-1)
	case galleryUrl != "" && listFile != "":
		fmt.Println("参数错误，请在命令行中运行参数-h以查看帮助")
		os.Exit(-1)
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

	//创建下载器
	downloader := GalleryDownloader{}
	//设置输出颜色
	success := color.New(color.Bold, color.FgGreen).FprintlnFunc()
	fail := color.New(color.Bold, color.FgRed).FprintlnFunc()

	for _, url := range galleryUrlList {
		success(os.Stdout, "开始下载gallery:", url)
		err := downloader.Download(infoJsonPath, url, onlyInfo)
		if err != nil {
			fail(os.Stderr, "下载失败:", err, "\n")
			continue
		}
		success(os.Stdout, "gallery下载完毕:", url, "\n")
	}

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	success(os.Stdout, "所有gallery下载完毕，共耗时:", endTime.Sub(startTime).Seconds(), "秒")
}
