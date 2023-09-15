package main

import (
	"ComicCrawler/client"
	"ComicCrawler/comicSites/dmzj"
	"ComicCrawler/comicSites/eh"
	"ComicCrawler/comicSites/happymh"
	"ComicCrawler/comicSites/mmmlf"
	"ComicCrawler/utils"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const infoJsonPath = "galleryInfo.json"

var (
	galleryUrl string
	onlyInfo   bool
	listFile   string
	update     bool
	buildTime  string
	goVersion  string
	version    = "v1.0.0"
)

type GalleryInfo struct {
	URL   string `json:"gallery_url"`
	Title string `json:"gallery_title"`
}

type GalleryDownloader struct{}

func (gd GalleryDownloader) Download(infoJson string, url string, onlyInfo bool) error {
	// 根据正则表达式判断是哪个软件包的gallery，并调用相应的下载函数
	if matched, _ := regexp.MatchString(`^https://e-hentai.org/g/[a-z0-9]*/[a-z0-9]{10}/$`, url); matched {
		err := eh.DownloadGallery(infoJson, url, onlyInfo)
		if err != nil {
			return err
		}
	} else if matched, _ := regexp.MatchString(`^https://manhua.dmzj.com/[a-z0-9]*/$`, url); matched {
		err := dmzj.DownloadGallery(infoJson, url, onlyInfo)
		if err != nil {
			return err
		}
	} else if matched, _ := regexp.MatchString(`^https://mmmlf.com/book/[0-9]*$`, url); matched {
		mmmlf.DownloadGallery(infoJson, url, onlyInfo)
	} else if matched, _ := regexp.MatchString(`^https://m.happymh.com/manga/[a-zA-z0-9]*$`, url); matched {
		//因为cloudflare的反爬机制比较严格，所以这里需要设置DebugMode为1
		client.DebugMode = "1"
		err := happymh.DownloadGallery(infoJson, url, onlyInfo)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("未知的url格式：%s", url)
	}
	return nil
}

func getDownLoadedGalleryUrl() []string {
	galleryInfo := GalleryInfo{}
	// 将下载过的画廊地址添加到列表中
	var downloadedGalleryUrlList []string

	currentDir, err := os.Getwd()
	if err != nil {
		log.Println("获取当前目录时出错：", err)
		return downloadedGalleryUrlList
	}

	// 递归遍历目录
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("获取当前目录时出错：", err)
			return err
		}

		// 检查是否是文件夹并且文件名是 galleryInfo.json
		if !info.IsDir() && info.Name() == "galleryInfo.json" {
			// 解析 JSON 数据
			err = utils.LoadCache(path, &galleryInfo)
			if err != nil {
				return err
			}

			downloadedGalleryUrlList = append(downloadedGalleryUrlList, galleryInfo.URL)
			//log.Println(galleryInfo)
		}

		return nil
	})

	if err != nil {
		fmt.Println("遍历目录时出错：", err)
	}
	return downloadedGalleryUrlList
}

func initArgsParse() {
	flag.StringVar(&galleryUrl, "url", "", "待下载的画廊地址（必填）")
	flag.StringVar(&galleryUrl, "u", "", "待下载的画廊地址（必填）")
	flag.BoolVar(&onlyInfo, "info", false, "只获取画廊信息(true/false)，默认为false")
	flag.BoolVar(&onlyInfo, "i", false, "只获取画廊信息(true/false)，默认为false")
	flag.StringVar(&listFile, "list", "", "待下载的画廊地址列表文件，每行一个url。(不能与参数-url同时使用)")
	flag.StringVar(&listFile, "l", "", "待下载的画廊地址列表文件，每行一个url。(不能与参数-url同时使用)")
	flag.BoolVar(&update, "update", false, "更新全部已下载的漫画，不能与其他任何参数一起使用")
}

func getExecutionTime(startTime time.Time, endTime time.Time) string {
	//按时:分:秒格式输出
	duration := endTime.Sub(startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d时%d分%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}

func main() {
	//版本信息
	args := os.Args
	//--version 或 -v
	if len(args) == 2 && (args[1] == "--version" || args[1] == "-v") {
		fmt.Printf("Version: %s \n", version)
		fmt.Printf("Build TimeStamp: %s \n", buildTime)
		fmt.Printf("GoLang Version: %s \n", goVersion)
		os.Exit(0)
	}

	var galleryUrlList []string
	//解析flag参数
	initArgsParse()
	flag.Parse()

	switch {
	case update:
		galleryUrlList = getDownLoadedGalleryUrl()
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
	errCount := 0
	for _, url := range galleryUrlList {
		success(os.Stdout, "开始下载gallery:", url)
		err := downloader.Download(infoJsonPath, url, onlyInfo)
		if err != nil {
			fail(os.Stderr, "下载失败:", err, "\n")
			errCount++
			continue
		}
		success(os.Stdout, "gallery下载完毕:", url, "\n")
	}

	//记录结束时间
	endTime := time.Now()
	//计算执行时间，单位为秒
	success(os.Stdout, "所有gallery下载完毕，共耗时:", getExecutionTime(startTime, endTime))
	if errCount > 0 {
		fail(os.Stderr, "其中有", errCount, "个下载失败")
	}
}
