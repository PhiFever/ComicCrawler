package client

import (
	"fmt"
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
		// 重置重试计数为初始值
		retryCount = 0
		//log.Println("Visited", r.Request.URL)
		//fmt.Println(string(r.Body))
	})
	return c
}
