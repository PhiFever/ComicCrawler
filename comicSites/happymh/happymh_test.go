package happymh

import (
	"ComicCrawler/client"
	"ComicCrawler/utils"
	"github.com/PuerkitoBio/goquery"
	"os"
	"reflect"
	"testing"
)

// 主函数和测试函数调用路径的区别
const localCookiesPath = "../../cookies.json"

var (
	cookies      = client.ReadCookiesFromFile(localCookiesPath)
	cookiesParam = client.ConvertCookies(cookies)
)

func TestGetClickedRenderedPage(t *testing.T) {
	// 初始化 Chromedp 上下文
	ctx, cancel := client.InitializeChromedpContext(false)
	defer cancel()
	html := client.GetClickedRenderedPage(ctx, "https://m.happymh.com/manga/SWEETHOME", cookiesParam, "#expandButton")
	//把html内容写入文件
	err := os.WriteFile("../../static/SWEETHOME/menu.html", html, 0666)
	utils.ErrorCheck(err)
}

func TestGetScrolledPage(t *testing.T) {
	ctx, cancel := client.InitializeChromedpContext(true)
	defer cancel()
	htmlContent := client.GetScrolledPage(ctx, cookiesParam, "https://m.happymh.com/reads/SWEETHOME/1946867")
	//把html内容写入文件
	err := os.WriteFile("../../static/SWEETHOME/page.htmlContent", htmlContent, 0666)
	utils.ErrorCheck(err)
}

func Test_getImagePageInfoList(t *testing.T) {
	// 初始化 Chromedp 上下文
	ctx, cancel := client.InitializeChromedpContext(false)
	defer cancel()
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name                    string
		args                    args
		wantImagePageInfoList   []map[int]string
		wantIndexToTitleMapList []map[int]string
	}{
		{
			name: "SWEET HOME",
			args: args{
				doc: GetMenuHtmlDoc(ctx, cookiesParam, "https://m.happymh.com/manga/SWEETHOME"),
			},
			wantImagePageInfoList: []map[int]string{
				{0: "https://m.happymh.com/reads/SWEETHOME/1946867"},
				{1: "https://m.happymh.com/reads/SWEETHOME/1946868"},
				{2: "https://m.happymh.com/reads/SWEETHOME/1946869"},
				{3: "https://m.happymh.com/reads/SWEETHOME/1946870"},
			},
			wantIndexToTitleMapList: []map[int]string{
				{0: "序幕"},
				{1: "第1话"},
				{2: "第2话"},
				{3: "第3话"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImagePageInfoList, gotIndexToTitleMapList := getImagePageInfoList(tt.args.doc)
			gotImagePageInfoList = gotImagePageInfoList[:4]
			gotIndexToTitleMapList = gotIndexToTitleMapList[:4]
			if !reflect.DeepEqual(gotImagePageInfoList, tt.wantImagePageInfoList) {
				t.Errorf("getImagePageInfoList() gotImagePageInfoList = %v, want %v", gotImagePageInfoList, tt.wantImagePageInfoList)
			}
			if !reflect.DeepEqual(gotIndexToTitleMapList, tt.wantIndexToTitleMapList) {
				t.Errorf("getImagePageInfoList() gotIndexToTitleMapList = %v, want %v", gotIndexToTitleMapList, tt.wantIndexToTitleMapList)
			}
		})
	}
}

func Test_getImageUrlListFromPage(t *testing.T) {
	// 初始化 Chromedp 上下文
	//ctx, cancel := client.InitializeChromedpContext(true)
	//defer cancel()
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "SWEET HOME 序幕",
			args: args{
				//doc: client.GetScrolledPage(ctx, "https://m.happymh.com/reads/SWEETHOME/1946867",cookiesParam),
				doc: client.ReadHtmlDoc("../../static/SWEETHOME/page.html"),
			},
			want: []string{
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/0d1ecb53f86000d7d0f95d23cfd2015e.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/e908dced3fa3e39406e08a0d20b31dcb.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/9023acfb3f394c36cd608474d775aa22.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/ea1931741cef35ca30d701f7b568c23d.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/7a1788ab5dfe057f311e4642ce655244.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/bfa183442d678a0921a943c6edb323cd.jpg",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getImageUrlListFromPage(tt.args.doc)[:6]
			for i, url := range got {
				if url != tt.want[i] {
					t.Errorf("getImageUrlListFromPage() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
