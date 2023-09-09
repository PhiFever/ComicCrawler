// Package happymh m.happymh.com
package happymh

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

// GetRenderedPage 获取经过JavaScript渲染后的页面
func GetRenderedPage(ctx context.Context, url string, cookieParams []*network.CookieParam, clickSelector string) ([]byte, error) {
	log.Println("正在渲染页面:", url)

	var htmlContent string
	err := chromedp.Run(ctx,
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.Click(clickSelector),
		chromedp.OuterHTML("html", &htmlContent),
	)
	log.Println("渲染完毕", url)
	if err != nil {
		log.Fatal(err)
	}
	return []byte(htmlContent), nil
}
