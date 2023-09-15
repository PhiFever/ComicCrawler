package happymh

import (
	"ComicCrawler/client"
	"github.com/PuerkitoBio/goquery"
	"reflect"
	"testing"
)

// 主函数和测试函数调用路径的区别
const localCookiesPath = "../../happymh_cookies.json"

var (
	cookies      = client.ReadCookiesFromFile(localCookiesPath)
	cookiesParam = client.ConvertCookies(cookies)
)

func Test_getImagePageInfoList(t *testing.T) {
	// 初始化 Chromedp 上下文
	ctx, cancel := client.InitChromedpContext(false)
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
				//doc: client.ReadHtmlDoc("../../static/SWEETHOME/menu.html"),
				doc: client.GetHtmlDoc(client.GetClickedRenderedPage(ctx, cookiesParam, "https://m.happymh.com/manga/SWEETHOME", "#expandButton")),
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
				t.Errorf("getImagePageInfoList() gotImagePageInfoList = %v, wantUrlList %v", gotImagePageInfoList, tt.wantImagePageInfoList)
			}
			if !reflect.DeepEqual(gotIndexToTitleMapList, tt.wantIndexToTitleMapList) {
				t.Errorf("getImagePageInfoList() gotIndexToTitleMapList = %v, wantUrlList %v", gotIndexToTitleMapList, tt.wantIndexToTitleMapList)
			}
		})
	}
}

func Test_getImageUrlListFromPage(t *testing.T) {
	// 初始化 Chromedp 上下文
	ctx, cancel := client.InitChromedpContext(true)
	defer cancel()
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name              string
		args              args
		wantUrlList       []string
		wantUrlListLength int
	}{
		{
			name: "SWEET HOME 序幕",
			args: args{
				doc: client.GetHtmlDoc(client.GetScrolledRenderedPage(ctx, cookiesParam, "https://m.happymh.com/reads/SWEETHOME/1946867")),
				//doc: client.ReadHtmlDoc("../../static/SWEETHOME/page.html"),
			},
			wantUrlList: []string{
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/0d1ecb53f86000d7d0f95d23cfd2015e.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/e908dced3fa3e39406e08a0d20b31dcb.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/9023acfb3f394c36cd608474d775aa22.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/ea1931741cef35ca30d701f7b568c23d.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/7a1788ab5dfe057f311e4642ce655244.jpg",
				"https://ruicdn.happymh.com/1f290a226753ed7e0c3d3689e1c84102/bfa183442d678a0921a943c6edb323cd.jpg",
			},
			wantUrlListLength: 36,
		},
		{
			name: "SWEET HOME 第1话",
			args: args{
				doc: client.GetHtmlDoc(client.GetScrolledRenderedPage(ctx, cookiesParam, "https://m.happymh.com/reads/SWEETHOME/1946868")),
			},
			wantUrlList:       []string{},
			wantUrlListLength: 50,
		},
		{
			name: "SWEET HOME 第127话",
			args: args{
				doc: client.GetHtmlDoc(client.GetScrolledRenderedPage(ctx, cookiesParam, "https://m.happymh.com/reads/SWEETHOME/1947010")),
			},
			wantUrlList: []string{
				`https://ruicdn.happymh.com/4c3d3575013ec8d5faafae58c4722a98/02dd2c351655516a5522637c0b3be705.jpg`,
				`https://ruicdn.happymh.com/4c3d3575013ec8d5faafae58c4722a98/c12d4f90046fe380f79ececde19470d8.jpg`,
				`https://ruicdn.happymh.com/4c3d3575013ec8d5faafae58c4722a98/25658aadea21ae23f31409c4c5f64a6c.jpg`,
				`https://ruicdn.happymh.com/4c3d3575013ec8d5faafae58c4722a98/b1fa0264793fb0aaf116b171347617db.jpg`,
				`https://ruicdn.happymh.com/4c3d3575013ec8d5faafae58c4722a98/ed3e39fd6590c0ae49ce33472eb8d4b8.jpg`,
				`https://ruicdn.happymh.com/4c3d3575013ec8d5faafae58c4722a98/57f0b34e86a15200dfb9fe5a0d475374.jpg`,
			},
			wantUrlListLength: 75,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getImageUrlListFromPage(tt.args.doc)
			if len(got) != tt.wantUrlListLength {
				t.Errorf("getImageUrlListFromPage() got = %v, wantUrlListLength %v", len(got), tt.wantUrlListLength)
			}
			got = got[:6]
			for i, gotUrl := range got {
				if gotUrl != tt.wantUrlList[i] {
					t.Errorf("getImageUrlListFromPage() got = %v, wantUrlList %v", got, tt.wantUrlList)
				}
			}
		})
	}
}
