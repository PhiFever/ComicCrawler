package client

import (
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
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
		RandomDelay: 2 * time.Second,
		Parallelism: 5, //并发数
	}
	_ = c.Limit(rule)

	//设置超时时间
	c.SetRequestTimeout(60 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
		r.Headers = &headers
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

func BuildJpegHeaders() http.Header {
	headers := http.Header{
		"Accept":             {"image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8"},
		"Accept-Encoding":    {"gzip, deflate, br"},
		"Accept-Language":    {"zh-CN,zh;q=0.9"},
		"Connection":         {"keep-alive"},
		"Dnt":                {"1"},
		"Host":               {"dqoaprm.qgankvrkxxiw.hath.network"},
		"Referer":            {"https://e-hentai.org/"},
		"Sec-Ch-Ua":          {"\"Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"115\", \"Chromium\";v=\"115\""},
		"Sec-Ch-Ua-Mobile":   {"?0"},
		"Sec-Ch-Ua-Platform": {"\"Windows\""},
		"Sec-Fetch-Dest":     {"image"},
		"Sec-Fetch-Mode":     {"no-cors"},
		"Sec-Fetch-Site":     {"cross-site"},
		"Sec-Gpc":            {"1"},
		"User-Agent":         {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
	}

	//for key, values := range headers {
	//	fmt.Printf("%s: %s\n", key, values)
	//}
	return headers
}
