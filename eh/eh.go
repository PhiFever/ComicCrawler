package eh

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cast"
	"log"
	"net/url"
	"regexp"
	"strings"
)

type GalleryInfo struct {
	URL        string              `json:"gallery_url"`
	Title      string              `json:"gallery_title"`
	TotalImage int                 `json:"total_image"`
	TagList    map[string][]string `json:"tag_list"`
}

func GetGalleryInfo(c *colly.Collector, galleryUrl string) GalleryInfo {
	var galleryInfo GalleryInfo
	galleryInfo.TagList = make(map[string][]string)
	galleryInfo.URL = galleryUrl

	//找到<h1 id="gn">标签,即为文章标题
	c.OnHTML("h1#gn", func(e *colly.HTMLElement) {
		galleryInfo.Title = e.Text
	})

	//找到<td class="gdt2">标签
	reMaxPage := regexp.MustCompile(`(\d+) pages`)
	c.OnHTML("td.gdt2", func(e *colly.HTMLElement) {
		if reMaxPage.MatchString(e.Text) {
			//转换为int
			galleryInfo.TotalImage, _ = cast.ToIntE(reMaxPage.FindStringSubmatch(e.Text)[1])
		}
	})

	// 找到<div id="taglist">标签
	c.OnHTML("div#taglist", func(e *colly.HTMLElement) {
		// 查找<div id="taglist">标签下的<table>元素
		e.ForEach("table", func(_ int, el *colly.HTMLElement) {
			// 在每个<table>元素中查找<tr>元素
			el.ForEach("tr", func(_ int, el *colly.HTMLElement) {
				//获取<tr>元素的<td class="tc">标签
				key := el.ChildText("td.tc")
				rp := strings.NewReplacer(":", "")
				localKey := rp.Replace(key) // 创建局部变量来保存循环迭代中的key值
				//fmt.Printf("key=%s: \n", localKey)
				el.ForEach("td", func(_ int, el *colly.HTMLElement) {
					el.ForEach("div", func(_ int, el *colly.HTMLElement) {
						//fmt.Println(el.Text)
						if _, ok := galleryInfo.TagList[localKey]; ok {
							galleryInfo.TagList[localKey] = append(galleryInfo.TagList[localKey], el.Text)
						} else {
							galleryInfo.TagList[localKey] = []string{el.Text}
						}
					})
				})
				//fmt.Println()
			})
		})
	})

	err := c.Visit(galleryUrl)
	if err != nil {
		log.Fatal(err)
		return galleryInfo
	}
	//fmt.Println(galleryInfo.TagList)
	return galleryInfo
}

func GenerateIndexURL(urlStr string, page int) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	if page == 0 {
		return u.String()
	}

	q := u.Query()
	q.Set("p", cast.ToString(page))
	u.RawQuery = q.Encode()

	return u.String()
}

// GetAllImagePageUrl 获取图片页面的url
func GetAllImagePageUrl(c *colly.Collector, indexUrl string) []string {
	var imagePageUrls []string
	c.OnHTML("div[id='gdt']", func(e *colly.HTMLElement) {
		//找到其下所有<div class="gdtm">标签
		e.ForEach("div.gdtm", func(_ int, el *colly.HTMLElement) {
			//找到<a href="xxx">标签，只有一个
			imgUrl := el.ChildAttr("a", "href")
			imagePageUrls = append(imagePageUrls, imgUrl)
		})
	})
	err := c.Visit(indexUrl)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return imagePageUrls
}

func GetImageUrl(c *colly.Collector, imagePageUrl string) string {
	//id="img"的src属性
	var imageUrl string
	c.OnHTML("img[id='img']", func(e *colly.HTMLElement) {
		imageUrl = e.Attr("src")
	})
	err := c.Visit(imagePageUrl)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return imageUrl
}
