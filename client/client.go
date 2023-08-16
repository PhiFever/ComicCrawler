package client

import (
	"github.com/gocolly/colly/v2"
	"log"
	"time"
)

func InitCollector() *colly.Collector {
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
		RandomDelay: 2 * time.Second,
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
