package utils

import (
	"fmt"
	"github.com/spf13/cast"
	"path/filepath"
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
	cacheImageDataList, err := LoadCache(filepath.Join(saveDir, cacheFile))
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
