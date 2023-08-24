package client

import (
	"ComicCrawler/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func InitCollector(headers http.Header) *colly.Collector {
	c := colly.NewCollector(
	//表示异步抓取
	//colly.Async(true),
	//允许重复访问
	//colly.AllowURLRevisit(),
	)
	//限制采集规格
	rule := &colly.LimitRule{
		//TODO: 限制规则似乎不起效果，需要进一步研究
		//理论上来说每次请求前会有访问延迟，但是实际使用的时候感觉不出来，不知道为什么
		RandomDelay: 5 * time.Second,
		Parallelism: 5, //并发数
	}
	_ = c.Limit(rule)

	//以下的所有设置在重新设置后都会失效

	//设置超时时间
	c.SetRequestTimeout(60 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
		r.Headers = &headers
	})

	maxRetries := 3
	retryCount := 0
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)

		// 检查是否达到最大重试次数
		if retryCount < maxRetries {
			retryCount++
			fmt.Printf("Retry attempt %d out of %d...\n", retryCount, maxRetries)
			//TODO: 重试策略似乎不起效果，需要进一步研究
			//等待指数退避
			//waitSeconds := math.Pow(2, float64(2*(retryCount+1)))
			//等待retryCount分钟
			waitSeconds := retryCount * 60
			fmt.Printf("Waiting %d seconds...\n", waitSeconds)
			time.Sleep(time.Duration(waitSeconds) * time.Second)

			// 重新尝试连接
			err := r.Request.Retry()
			log.Println("Retry error:", err)
		} else {
			log.Fatal("Exceeded maximum retry attempts. Cannot establish connection.")
		}
	})

	c.OnResponse(func(r *colly.Response) {
		//TODO: 重试策略似乎不起效果，需要进一步研究
		// 重置重试计数为初始值
		retryCount = 0
		//log.Println("Visited", r.Request.URL)
		//fmt.Println(string(r.Body))
	})
	return c
}

// Cookie 以下是使用chromedp的相关代码
// Cookie 从 Chrome 中使用EditThisCookie导出的 Cookies
type Cookie struct {
	Domain     string  `json:"domain"`
	Expiration float64 `json:"expirationDate"`
	HostOnly   bool    `json:"hostOnly"`
	HTTPOnly   bool    `json:"httpOnly"`
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	SameSite   string  `json:"sameSite"`
	Secure     bool    `json:"secure"`
	Session    bool    `json:"session"`
	StoreID    string  `json:"storeId"`
	Value      string  `json:"value"`
	ID         int     `json:"id"`
}

// ReadCookiesFromFile 从文件中读取 Cookies
func ReadCookiesFromFile(filePath string) ([]Cookie, error) {
	var cookies []Cookie

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

// ConvertCookies 将从文件中读取的 Cookies 转换为 chromedp 需要的格式
func ConvertCookies(cookies []Cookie) []*network.CookieParam {
	cookieParams := make([]*network.CookieParam, len(cookies))
	for i, cookie := range cookies {
		cookieParams[i] = &network.CookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			HTTPOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
		}
	}
	return cookieParams
}

// GetRenderedPage 获取经过JavaScript渲染后的页面
func GetRenderedPage(url string, cookieParams []*network.CookieParam) ([]byte, error) {
	log.Println("正在渲染页面...", url)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),    // 是否以无头模式运行
		chromedp.Flag("disable-gpu", true), // 禁用GPU
		chromedp.Flag("no-sandbox", true),  // 禁用沙盒模式
		chromedp.Flag("–disable-plugins", true),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		network.SetCookies(cookieParams),
		chromedp.Navigate(url), // 替换为你想要访问的网址
		// 等待<div class="anim-main_list">节点加载完毕（实际上不好用）
		//chromedp.WaitVisible("div.anim-main_list", chromedp.ByQuery),
		// 等待5秒，保证页面加载完毕
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	log.Println("页面渲染完毕")
	if err != nil {
		log.Fatal(err)
	}
	return []byte(htmlContent), nil
}

// GetHtmlDoc 读取cookies文件，获取经过JavaScript渲染后的页面
func GetHtmlDoc(cookiesParam []*network.CookieParam, galleryUrl string) *goquery.Document {
	//实际使用时的代码
	htmlContent, err := GetRenderedPage(galleryUrl, cookiesParam)
	// 将 []byte 转换为 io.Reader
	reader := bytes.NewReader(htmlContent)
	doc, err := goquery.NewDocumentFromReader(reader)
	utils.ErrorCheck(err)

	////TODO: 以下代码是为了方便调试，实际使用时需要注释掉
	////获取当前路径
	//dir, err := os.Getwd()
	//utils.ErrorCheck(err)
	//var htmlCachePath string
	//if dir == `E:\Go_project\WorkSpace\ComicCrawler\dmzj` {
	//	htmlCachePath = `../static/dmzj_chromedp.html`
	//} else if dir == `E:\Go_project\WorkSpace\ComicCrawler` {
	//	htmlCachePath = `./static/dmzj_chromedp.html`
	//}
	//htmlContent, _ := os.Open(htmlCachePath)
	//doc, err := goquery.NewDocumentFromReader(htmlContent)
	//utils.ErrorCheck(err)

	return doc
}
