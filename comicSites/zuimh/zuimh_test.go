package zuimh

import (
	"ComicCrawler/client"
	"reflect"
	"testing"
)

const (
	localCookiesPath = `../../cookies.json`
)

var baseCollector = client.InitCookiesCollector(client.ReadCookiesFromFile(localCookiesPath), "https://www.zuimh.com/")

func Test_getGalleryInfo(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want GalleryInfo
	}{
		{
			name: "SWEET HOME",
			args: args{
				"https://www.zuimh.com/manhua/SWEETHOME/#chapters",
			},
			want: GalleryInfo{
				URL:         "https://www.zuimh.com/manhua/SWEETHOME/#chapters",
				Title:       "SWEET HOME",
				LastChapter: "阿萨布鲁柑仔店土丸仔",
				TagList:     map[string][]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getGalleryInfo(baseCollector, tt.args.url)
			if !reflect.DeepEqual(got, tt.want) {
				if got.URL != tt.want.URL {
					t.Errorf("getGalleryInfo() URL got = %v, want %v", got.URL, tt.want.URL)
				}
				if got.Title != tt.want.Title {
					t.Errorf("getGalleryInfo() Title got = %v, want %v", got.Title, tt.want.Title)
				}
				if got.LastChapter != tt.want.LastChapter {
					t.Errorf("getGalleryInfo() LastChapter got = %v, want %v", got.LastChapter, tt.want.LastChapter)
				}
				if !reflect.DeepEqual(got.TagList, tt.want.TagList) {
					t.Errorf("getGalleryInfo() TagList got = %v, want %v", got.TagList, tt.want.TagList)
				}
			}
		})
	}
}

func Test_getAllImagePageInfo(t *testing.T) {
	type args struct {
		indexUrl string
	}
	tests := []struct {
		name                  string
		args                  args
		wantImagePageInfoList []map[int]string
		wantIndexToNameMap    []map[int]string
	}{
		{
			name: "SWEET HOME",
			args: args{
				"https://www.zuimh.com/manhua/SWEETHOME/#chapters",
			},
			wantImagePageInfoList: []map[int]string{
				{0: "https://www.zuimh.com/manhua/SWEETHOME/885365.html"},
				{1: "https://www.zuimh.com/manhua/SWEETHOME/885366.html"},
				{2: "https://www.zuimh.com/manhua/SWEETHOME/885367.html"},
			},
			wantIndexToNameMap: []map[int]string{
				{0: "序幕"},
				{1: "1话"},
				{2: "2话"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImagePageInfoList, gotIndexToNameMap := getAllImagePageInfo(baseCollector, tt.args.indexUrl)
			gotImagePageInfoList = gotImagePageInfoList[:3]
			gotIndexToNameMap = gotIndexToNameMap[:3]
			if !reflect.DeepEqual(gotImagePageInfoList, tt.wantImagePageInfoList) {
				for i, gotImagePageInfo := range gotImagePageInfoList {
					for k, v := range gotImagePageInfo {
						if v != tt.wantImagePageInfoList[i][k] {
							t.Errorf("getAllImagePageInfo() gotImagePageInfoList[%d][%d] got = %v, want %v", i, k, v, tt.wantImagePageInfoList[i][k])
						}
					}
				}
			}
			if !reflect.DeepEqual(gotIndexToNameMap, tt.wantIndexToNameMap) {
				t.Errorf("getAllImagePageInfo() gotIndexToNameMap = %v, want %v", gotIndexToNameMap, tt.wantIndexToNameMap)
			}
		})
	}
}
