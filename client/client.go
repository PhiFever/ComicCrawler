package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	DebugMode = true
	DelayMs   = 330
)

func TrueRandFloat(min, max float64) float64 {
	// 使用当前时间作为种子值
	seed := time.Now().Unix()
	source := rand.NewSource(seed)
	randomGenerator := rand.New(source)

	// 生成范围在 [min, max) 内的随机浮点数
	randomFloat := min + randomGenerator.Float64()*(max-min)
	return randomFloat
}

func TrueRandInt(min, max int) int {
	// 使用当前时间作为种子值
	seed := time.Now().Unix()
	source := rand.NewSource(seed)
	randomGenerator := rand.New(source)

	// 生成范围在 [min, max) 内的随机整数
	randomInt := min + randomGenerator.Intn(max-min)
	return randomInt
}

func InitJPEGCollector(headers http.Header) *colly.Collector {
	c := colly.NewCollector()
	////TODO: 限制规则似乎不起效果，需要进一步研究
	////限制采集规格
	//rule := &colly.LimitRule{
	//	//理论上来说每次请求前会有访问延迟，但是实际使用的时候感觉不出来，不知道为什么
	//	RandomDelay: 5 * time.Second,
	//	Parallelism: 5, //并发数
	//}
	//_ = c.Limit(rule)

	//设置超时时间
	c.SetRequestTimeout(30 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
		r.Headers = &headers
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnResponse(func(r *colly.Response) {
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
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
		}
	}
	return cookieParams
}

// InitChromedpContext 实际在每次调用时可以派生一个新的超时context，然后在这个新的context中执行任务，可以避免卡住
// //timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
// //defer cancel()
func InitChromedpContext(imageEnabled bool) (context.Context, context.CancelFunc) {
	log.Println("正在初始化 Chromedp 上下文")
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !DebugMode),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("–disable-plugins", true),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36`),
		chromedp.Flag("blink-settings", "imagesEnabled="+fmt.Sprintf("%t", imageEnabled)),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

	chromeCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	return chromeCtx, cancel
}

// GetRenderedPage 固定等待5秒，获取经过JavaScript渲染后的页面
func GetRenderedPage(ctx context.Context, cookieParams []*network.CookieParam, url string) []byte {
	log.Println("正在渲染页面:", url)

	var htmlContent string
	// 具体任务放在这里
	var tasks = chromedp.Tasks{
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		//chromedp.WaitVisible("???", chromedp.ByQuery),
		chromedp.Sleep(5 * time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	}

	//开始执行任务
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("渲染完毕", url)
	return []byte(htmlContent)
}

// GetWaitVisibleRenderedPage 等待指定元素可见，获取经过JavaScript渲染后的页面
func GetWaitVisibleRenderedPage(ctx context.Context, cookieParams []*network.CookieParam, url string, selector string) []byte {
	log.Println("正在渲染页面:", url)

	var htmlContent string
	// 具体任务放在这里
	var tasks = chromedp.Tasks{
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	}

	//开始执行任务
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("渲染完毕", url)
	return []byte(htmlContent)
}

// GetClickedRenderedPage 获取需要点击展开的经过JavaScript渲染后的页面
func GetClickedRenderedPage(ctx context.Context, cookieParams []*network.CookieParam, url string, clickSelector string) []byte {
	log.Println("正在渲染页面:", url)

	var htmlContent string
	// 具体任务放在这里
	var tasks = chromedp.Tasks{
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		chromedp.WaitVisible(clickSelector, chromedp.ByQuery),
		chromedp.Sleep(time.Millisecond * time.Duration(TrueRandInt(DelayMs, 2*DelayMs+100))),
		chromedp.Click(clickSelector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	}

	//开始执行任务
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("渲染完毕", url)
	return []byte(htmlContent)
}

// GetScrolledRenderedPage 获取需要整个页面滚动到底部后经过JavaScript渲染的页面
func GetScrolledRenderedPage(ctx context.Context, cookieParams []*network.CookieParam, url string) []byte {
	log.Println("正在渲染页面:", url)

	var height int
	// 具体任务放在这里
	var tasks = chromedp.Tasks{
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		//获取当前页面的高度
		chromedp.Evaluate(`document.body.scrollHeight`, &height),
	}
	//开始执行任务
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatal(err)
	}

	//每次滚动的距离（像素）
	scrollLength := 1500
	//增加滚轮滚动的任务
	var scrollTask = chromedp.Tasks{}
	for i := 0; i < height; i += scrollLength {
		//scrollTask = append(scrollTask, chromedp.Sleep(1*time.Second))
		scrollTask = append(scrollTask, chromedp.ActionFunc(func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 2 * time.Duration(DelayMs))
			// 在页面的（200，200）坐标的位置
			p := input.DispatchMouseEvent(input.MouseWheel, 200, 200)
			p = p.WithDeltaX(0)
			// 滚轮向下滚动1000单位
			p = p.WithDeltaY(float64(scrollLength))
			err := p.Do(ctx)
			return err
		}))
	}

	var htmlContent string
	scrollTask = append(scrollTask, chromedp.OuterHTML("html", &htmlContent))

	//fmt.Println(height)
	//开始执行任务
	err = chromedp.Run(ctx, scrollTask)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("渲染完毕", url)
	return []byte(htmlContent)
}

// GetHtmlDoc 从[]byte中读取html内容，返回goquery.Document
func GetHtmlDoc(htmlContent []byte) *goquery.Document {
	// 将 []byte 转换为 io.Reader
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

// ReadHtmlDoc 从文件中读取html内容，返回goquery.Document
func ReadHtmlDoc(filePath string) *goquery.Document {
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

	doc, err := goquery.NewDocumentFromReader(file)
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
