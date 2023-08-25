package utils

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestToSafeFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "case1",
			in:   `[sfs]\24r/f4?*<q>|:`,
			want: `[sfs]_24r_f4___q___`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSafeFilename(tt.in); got != tt.want {
				t.Errorf("ToSafeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

var imageDataList = []map[string]string{
	{
		"imageName": "1.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP.SQmqQt18WUcWYyuX8fGGGAHaE8?pid=ImgDet&rs=1`,
	},
	{
		"imageName": "2.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP.6L7shpwxVAIr279rA0B1JQHaE7?pid=ImgDet&rs=1`,
	},
	{
		"imageName": "3.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP.i242SBVfAPAhfxY5omlfgQHaLP?pid=ImgDet&rs=1`,
	},
	{
		"imageName": "4.jpg",
		"imageUrl":  `https://th.bing.com/th/id/OIP._0UYsgLTgJ8WAUYXFXKHRQHaEK?pid=ImgDet&rs=1`,
	},
}

var saveDir = "../test"
var cacheFile = "cache.json"

func TestBuildCache(t *testing.T) {
	err := BuildCache(saveDir, cacheFile, imageDataList)
	if err != nil {
		t.Errorf("buildCache() = %s; want nil", err)
	}
}

func TestLoadCache(t *testing.T) {
	var cacheImageDataList []map[string]string
	err := LoadCache(filepath.Join(saveDir, cacheFile), &cacheImageDataList)
	if err != nil {
		t.Errorf("loadCache() = %s; want nil", err)
	}
	for index, data := range cacheImageDataList {
		if data["imageName"] != imageDataList[index]["imageName"] || data["imageUrl"] != imageDataList[index]["imageUrl"] {
			t.Errorf("loadCache() = %s; want %s", data, imageDataList[index])
		}
	}

}

func TestSaveFile(t *testing.T) {
	data := []byte(cast.ToString(time.Now()))
	filePath, _ := filepath.Abs(filepath.Join(saveDir, "testSaveFile.txt"))

	err := SaveFile(filePath, data)
	if err != nil {
		t.Errorf("SaveFile() = %s; want nil", err)
	}
}

func TestRandFloat(t *testing.T) {
	type args struct {
		min float64
		max float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				min: 5,
				max: 15,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num := TrueRandFloat(tt.args.min, tt.args.max)
			fmt.Printf("TrueRandFloat() = %v; want %v\n", num, tt.args)
			if num < tt.args.min || num > tt.args.max {
				t.Errorf("TrueRandFloat() = %v; want %v", TrueRandFloat(tt.args.min, tt.args.max), tt.args)
			}
		})
	}
}

func TestGetFileTotal(t *testing.T) {
	type args struct {
		dirPath    string
		fileSuffix []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "../test",
			args: args{
				dirPath:    "../test",
				fileSuffix: []string{".jpg"},
			},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if total := GetFileTotal(tt.args.dirPath, tt.args.fileSuffix); total != tt.want {
				t.Errorf("GetFileTotal() = %v, want %v", total, tt.want)
			}
		})
	}
}

func TestReadListFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "case1",
			args: args{
				filePath: "../test/list.txt",
			},
			want: []string{
				"https://e-hentai.org/g/1111111/1111111111/",
				"https://e-hentai.org/g/2222222/2222222222/",
				"https://e-hentai.org/g/3333333/3333333333/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadListFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadListFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadListFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveImages(t *testing.T) {
	c := colly.NewCollector(
		colly.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	)
	absPath, err := filepath.Abs(saveDir)
	fmt.Println(absPath)
	if err != nil {
		t.Errorf("filepath.Abs() = %s; want nil", err)
	}
	err = SaveImages(c, imageDataList, absPath)
	if err != nil {
		t.Errorf("saveImages() = %s; want nil", err)
	}
	for _, data := range imageDataList {
		imagePath := filepath.Join(saveDir, data["imageName"])
		if !FileExists(imagePath) {
			t.Errorf("image not exists: %s", imagePath)
		}
	}
}

func TestExtractNumberFromText(t *testing.T) {
	type args struct {
		pattern string
		text    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "第08话",
			args: args{
				pattern: `第(\d+)话`,
				text:    "第08话",
			},
			want:    "08",
			wantErr: false,
		},
		{
			name: "第37话叉尾猫",
			args: args{
				pattern: `第(\d+)话`,
				text:    "第37话叉尾猫",
			},
			want:    "37",
			wantErr: false,
		},
		{
			name: "错误测试",
			args: args{
				pattern: `第(\d+)话`,
				text:    "sdasf",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSubstringFromText(tt.args.pattern, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSubstringFromText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractSubstringFromText() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortMapsByIntKey(t *testing.T) {
	type args struct {
		ascending bool
	}
	tests := []struct {
		name string
		args args
		want []map[int]string
	}{
		{
			name: "ascending",
			args: args{
				ascending: true,
			},
			want: []map[int]string{
				{0: "orange"}, {1: "apple"}, {3: "banana"}, {5: "pear"},
			},
		},
		{
			name: "descending",
			args: args{
				ascending: false,
			},
			want: []map[int]string{
				{5: "pear"}, {3: "banana"}, {1: "apple"}, {0: "orange"},
			},
		},
	}
	inputMap := []map[int]string{
		{3: "banana"},
		{1: "apple"},
		{5: "pear"},
		{0: "orange"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortMapsByIntKey(inputMap, tt.args.ascending); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortMapsByIntKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
