package dmzj

import (
	"ComicCrawler/client"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"reflect"
	"testing"
)

const localCookiesPath = "../cookies.json"

var (
	cookies, _   = client.ReadCookiesFromFile(localCookiesPath)
	cookiesParam = client.ConvertCookies(cookies)
)

func TestGetGalleryInfo(t *testing.T) {
	tests := []struct {
		name       string
		galleryUrl string
		want       GalleryInfo
	}{
		{
			name:       "成为夺心魔的必要",
			galleryUrl: "https://manhua.dmzj.com/chengweiduoxinmodebiyao/",
			want: GalleryInfo{
				URL:            "https://manhua.dmzj.com/chengweiduoxinmodebiyao/",
				Title:          "成为夺心魔的必要",
				LastChapter:    "149",
				LastUpdateTime: "2023-08-25",
				TagList: map[string][]string{
					"作者":   {"赖惟智"},
					"地域":   {"港台"},
					"状态":   {"连载中"},
					"人气":   {"30497493"},
					"分类":   {"青年漫画"},
					"题材":   {"欢乐向", "治愈", "西方魔幻"},
					"最新收录": {"第149话"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := client.GetHtmlDoc(cookiesParam, tt.galleryUrl)
			if got := GetGalleryInfo(doc, tt.galleryUrl); !reflect.DeepEqual(got, tt.want) {
				if got.Title != tt.want.Title {
					t.Errorf("title got: %v, want: %v", got.Title, tt.want.Title)
				}
				if got.LastChapter != tt.want.LastChapter {
					t.Errorf("lastChapter got: %v, want: %v", got.LastChapter, tt.want.LastChapter)
				}
				if !reflect.DeepEqual(got.TagList, tt.want.TagList) {
					for k, v := range got.TagList {
						if !reflect.DeepEqual(v, tt.want.TagList[k]) {
							t.Errorf("tagList got: %v, want: %v", v, tt.want.TagList[k])
							for i, j := range v {
								if j != tt.want.TagList[k][i] {
									t.Errorf("tag got: %v, want: %v", j, tt.want.TagList[k][i])
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestGetAllImagePageUrl(t *testing.T) {
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name string
		args args
		want []map[int]string
	}{
		{
			name: "成为夺心魔的必要",
			args: args{
				doc: client.GetHtmlDoc(cookiesParam, "https://manhua.dmzj.com/chengweiduoxinmodebiyao/"),
			},
			want: []map[int]string{
				{
					149: "https://manhua.dmzj.com/chengweiduoxinmodebiyao/139899.shtml#1",
				},
				{
					148: "https://manhua.dmzj.com/chengweiduoxinmodebiyao/139388.shtml#1",
				},
				{
					147: "https://manhua.dmzj.com/chengweiduoxinmodebiyao/138862.shtml#1",
				},
				{
					146: "https://manhua.dmzj.com/chengweiduoxinmodebiyao/138033.shtml#1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAllImagePageInfo(tt.args.doc)[0:4]
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllImagePageInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetImageUrlFromPage(t *testing.T) {
	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "02话",
			args: args{
				doc: client.GetHtmlDoc(cookiesParam, "https://manhua.dmzj.com/chengweiduoxinmodebiyao/102022.shtml#1"),
			},
			want: "https://images.idmzj.com/c%2F%E6%88%90%E4%B8%BA%E5%A4%BA%E5%BF%83%E9%AD%94%E7%9A%84%E5%BF%85%E8%A6%81%2F%E7%AC%AC02%E8%AF%9D_1597930984%2F41.jpg",
		},
		{
			name: "137话",
			args: args{
				doc: client.GetHtmlDoc(cookiesParam, "https://manhua.dmzj.com/chengweiduoxinmodebiyao/135075.shtml#1"),
			},
			want: "https://images.idmzj.com/c%2F%E6%88%90%E4%B8%BA%E5%A4%BA%E5%BF%83%E9%AD%94%E7%9A%84%E5%BF%85%E8%A6%81%2F%E7%AC%AC137%E8%AF%9D%2F137%E7%A0%94%E7%A9%B6%E6%9D%90%E6%96%99%20%E6%8B%B7%E8%B4%9D.jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetImageUrlFromPage(tt.args.doc); got != tt.want {
				t.Errorf("GetImageUrlFromPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncParsePage(t *testing.T) {
	type args struct {
		tasksData        map[int]string
		numWorkers       int
		tasks            chan map[int]string    //此处与原代码不同，原代码为<-chan map[int]string，但是这样会导致无法读取channel
		imageInfoChannel chan map[string]string //此处与原代码不同，原代码为chan<- map[string]string，但是这样会导致无法读取channel
		cookiesParam     []*network.CookieParam
	}
	tests := []struct {
		name string
		args args
		want []map[string]string
	}{
		{
			name: "成为夺心魔的必要",
			args: args{
				numWorkers:       5,
				tasks:            make(chan map[int]string, 5),
				imageInfoChannel: make(chan map[string]string, 5),
				cookiesParam:     cookiesParam,
				tasksData: map[int]string{
					137: "https://manhua.dmzj.com/chengweiduoxinmodebiyao/135075.shtml#1",
					2:   "https://manhua.dmzj.com/chengweiduoxinmodebiyao/102022.shtml#1",
					// ... 根据测试用例的实际数据添加更多条目
				},
			},
			want: []map[string]string{
				{
					"imageName": "137.jpg",
					"imageUrl":  `https://images.idmzj.com/c%2F%E6%88%90%E4%B8%BA%E5%A4%BA%E5%BF%83%E9%AD%94%E7%9A%84%E5%BF%85%E8%A6%81%2F%E7%AC%AC137%E8%AF%9D%2F137%E7%A0%94%E7%A9%B6%E6%9D%90%E6%96%99%20%E6%8B%B7%E8%B4%9D.jpg`,
				},
				{
					"imageName": "2.jpg",
					"imageUrl":  `https://images.idmzj.com/c%2F%E6%88%90%E4%B8%BA%E5%A4%BA%E5%BF%83%E9%AD%94%E7%9A%84%E5%BF%85%E8%A6%81%2F%E7%AC%AC02%E8%AF%9D_1597930984%2F41.jpg`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go syncParsePage(tt.args.tasks, tt.args.imageInfoChannel, tt.args.cookiesParam, tt.args.numWorkers) // 启动并发执行

			// 发送任务数据到tasks通道
			tt.args.tasks <- tt.args.tasksData

			// 接收所有发送到imageInfoChannel通道的数据
			var got []map[string]string
			for i := 0; i < len(tt.want); i++ {
				imageInfo := <-tt.args.imageInfoChannel
				got = append(got, imageInfo)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("syncParsePage() = %v, want %v", got, tt.want)
			}
		})
	}
}
