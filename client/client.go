package client

import (
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

const DEBUG_MODE = false

func InitJPEGCollector(headers http.Header) *colly.Collector {
	c := colly.NewCollector()
	//TODO: 限制规则似乎不起效果，需要进一步研究
	//限制采集规格
	rule := &colly.LimitRule{
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

		//TODO:如果是“远程主机强迫关闭了一个现有的连接”才进行重试

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
func ReadCookiesFromFile(filePath string) []Cookie {
	var cookies []Cookie

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &cookies)
	if err != nil {
		log.Fatal(err)
	}

	return cookies
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

func InitializeChromedpContext() (context.Context, context.CancelFunc) {
	log.Println("正在初始化 Chromedp 上下文")
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !DEBUG_MODE),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("–disable-plugins", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36`),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

	chromeCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	return chromeCtx, cancel
}

// GetRenderedPage 获取经过JavaScript渲染后的页面
// TODO:在实际应用中连续爬取多个网页不需要每次都重新创建chromeCtx，这个函数需要进一步优化
func GetRenderedPage(ctx context.Context, url string, cookieParams []*network.CookieParam) ([]byte, error) {
	log.Println("正在渲染页面:", url)

	//这个超时时间似乎是整个chromedp实例的超时时间，而不是单个url请求的超时时间，所以暂时不使用
	//// 超时时间为30秒
	//timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	//defer cancel()

	var htmlContent string
	//err := chromedp.Run(timeoutCtx,
	err := chromedp.Run(ctx,
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	log.Println("渲染完毕", url)
	if err != nil {
		log.Fatal(err)
	}
	return []byte(htmlContent), nil
}

// GetHtmlDoc 读取cookies文件，获取经过JavaScript渲染后的页面
func GetHtmlDoc(ctx context.Context, cookiesParam []*network.CookieParam, galleryUrl string) *goquery.Document {
	//实际使用时的代码
	htmlContent, err := GetRenderedPage(ctx, galleryUrl, cookiesParam)
	// 将 []byte 转换为 io.Reader
	reader := bytes.NewReader(htmlContent)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func InitCookiesCollector(cookies []Cookie, baseUrl string) *colly.Collector {
	//初始化Collector
	baseCollector := colly.NewCollector(
		colly.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 Edg/115.0.1901.203`),
	)
	// 将Cookies添加到Collector
	for _, cookie := range cookies {
		err := baseCollector.SetCookies(baseUrl, []*http.Cookie{
			{
				Name:  cookie.Name,
				Value: cookie.Value,
			},
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
	return baseCollector
}
