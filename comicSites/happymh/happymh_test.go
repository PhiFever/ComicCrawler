package happymh

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"os"
	"testing"
)

// 主函数和测试函数调用路径的区别
const localCookiesPath = "../../cookies.json"

var (
	cookies      = client.ReadCookiesFromFile(localCookiesPath)
	cookiesParam = client.ConvertCookies(cookies)
)

func Test_getRenderedPage(t *testing.T) {
	// 初始化 Chromedp 上下文
	ctx, cancel := client.InitializeChromedpContext()
	defer cancel()
	html, err := GetRenderedPage(ctx, "https://m.happymh.com/manga/lieqiangshaonian", cookiesParam, "#expandbutton")
	utils.ErrorCheck(err)
	//把html内容写入文件
	err = os.WriteFile("../../static/SWEETHOME/menu.html", html, 0666)
	utils.ErrorCheck(err)

	//html, err := client.GetRenderedPage(ctx, "https://m.happymh.com/reads/SWEETHOME/1946867", cookiesParam)
	//utils.ErrorCheck(err)
	////把html内容写入文件
	//err = os.WriteFile("../../static/SWEETHOME/page.html", html, 0666)
	//utils.ErrorCheck(err)
}
