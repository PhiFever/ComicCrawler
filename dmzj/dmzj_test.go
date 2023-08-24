package dmzj

import (
	"ComicCrawler/client"
	"github.com/PuerkitoBio/goquery"
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
				LastChapter:    "148",
				LastUpdateTime: "2023-08-07",
				TagList: map[string][]string{
					"作者":   {"赖惟智"},
					"地域":   {"港台"},
					"状态":   {"连载中"},
					"人气":   {"30490729"},
					"分类":   {"青年漫画"},
					"题材":   {"欢乐向", "治愈", "西方魔幻"},
					"最新收录": {"第148话"},
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
			got := GetAllImagePageInfo(tt.args.doc)[0:3]
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllImagePageInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
