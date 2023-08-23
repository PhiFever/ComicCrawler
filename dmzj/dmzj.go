package dmzj

import (
	"ComicDownloader/utils"
	"fmt"
)

type GalleryInfo struct {
	URL     string              `json:"gallery_url"`
	Title   string              `json:"gallery_title"`
	TagList map[string][]string `json:"tag_list"`
}

func GetGalleryInfo(galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	return galleryInfo
}

func DownloadGallery(infoJson string, galleryUrl string, onlyInfo bool) {
	//获取画廊信息
	galleryInfo := GetGalleryInfo(galleryUrl)
	safeTitle := utils.ToSafeFilename(galleryInfo.Title)
	fmt.Println(safeTitle)
}
