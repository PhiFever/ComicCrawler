package mmmlf

import (
	"ComicCrawler/client"
	"github.com/gocolly/colly/v2"
	"reflect"
	"testing"
)

var baseCollector = colly.NewCollector(colly.UserAgent(client.ChromeUserAgent))

func Test_getImagePageInfoList(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name                    string
		args                    args
		wantImagePageInfoList   []map[int]string
		wantIndexToTitleMapList []map[int]string
	}{
		{
			name: "甜蜜家园",
			args: args{
				url: "https://mmmlf.com/book/52020",
			},
			wantImagePageInfoList: []map[int]string{
				{1: "https://mmmlf.com/chapter/1069556"},
				{2: "https://mmmlf.com/chapter/1069560"},
				{3: "https://mmmlf.com/chapter/1069564"},
			},
			wantIndexToTitleMapList: []map[int]string{
				{1: "第1话"},
				{2: "第2话"},
				{3: "第3话"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImagePageInfoList, gotIndexToTitleMapList := getImagePageInfoList(baseCollector, tt.args.url)
			gotImagePageInfoList = gotImagePageInfoList[:3]
			gotIndexToTitleMapList = gotIndexToTitleMapList[:3]
			if !reflect.DeepEqual(gotImagePageInfoList, tt.wantImagePageInfoList) {
				for key, value := range gotImagePageInfoList {
					if !reflect.DeepEqual(value, tt.wantImagePageInfoList[key]) {
						t.Errorf("getImagePageInfoList() gotImagePageInfoList[%v] = %v, want %v", key, value, tt.wantImagePageInfoList[key])
					}
				}
			}
			if !reflect.DeepEqual(gotIndexToTitleMapList, tt.wantIndexToTitleMapList) {
				for key, value := range gotIndexToTitleMapList {
					if !reflect.DeepEqual(value, tt.wantIndexToTitleMapList[key]) {
						t.Errorf("getImagePageInfoList() gotIndexToTitleMapList[%v] = %v, want %v", key, value, tt.wantIndexToTitleMapList[key])
					}
				}
			}
		})
	}
}

func Test_getImageUrlListFromPage(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "甜蜜家园",
			args: args{
				url: "https://mmmlf.com/chapter/1069556",
			},
			want: []string{
				"https://16t.765567.xyz/cdn/copymgpic/tianmijiayuan/1/e238ac42-b0d5-11eb-af55-024352452ce0.jpg",
				"https://16t.765567.xyz/cdn/copymgpic/tianmijiayuan/1/e2bf437e-b0d5-11eb-af55-024352452ce0.jpg",
				"https://16t.765567.xyz/cdn/copymgpic/tianmijiayuan/1/e350e342-b0d5-11eb-af55-024352452ce0.jpg",
				"https://16t.765567.xyz/cdn/copymgpic/tianmijiayuan/1/e3dd421a-b0d5-11eb-af55-024352452ce0.jpg",
				"https://16t.765567.xyz/cdn/copymgpic/tianmijiayuan/1/e467b8dc-b0d5-11eb-af55-024352452ce0.jpg",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getImageUrlListFromPage(baseCollector, tt.args.url)[:5]
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImageUrlListFromPage() = %v, want %v", got, tt.want)
			}
		})
	}
}
