## 支持站点
目前支持`e-hentai.org`,`mmmlf.com`,`manhua.dmzj.com`(需要将cookies放在同级目录下的dmzj_cookies.json，可以通过EditThisCookie导出，注意调整翻页模式，使其中的fanyemodeval=2)
## TODO
1. dmzj目录带分页的情况无法处理（或许应该使用chromedp模拟点击处理，不过这种情况比较少，暂时懒得改了:)）
## 缓存文件格式
`galleryInfo.json`

1. eh
```json
{
    "gallery_url": "https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/",
    "gallery_title": "xxxxxx",
    "total_image": 100,
    "tag_list": {
        "artist": "xyz",
        "male": [
            "aaa",
            "bbb"
        ],
        "female":[
            "xxx",
            "yyy"
        ]
    }
}
```
2. dmzj
```json
{
    "gallery_url": "https://manhua.dmzj.com/safdgfbxbxvxc/",
    "gallery_title": "xxx",
    "last_chapter": "149",
    "last_update_time": "2023-08-25",
    "tag_list": {
        "作者": [
            "aaa"
        ],
        "分类": [
            "bbb"
        ],
        "地域": [
            "ccc"
        ],
        "最新收录": [
            "第149话"
        ],
        "状态": [
            "连载中"
        ],
        "题材": [
            "fsv",
            "cvfhr",
            "cnczef"
        ]
    }
}
```
##  使用说明
1. 获取详细说明（对应的短参数名）
```powershell
./ComicCrawler.exe -h
```
2. 获取gallery信息，并下载图片(url为gallery目录页的url，尾部注意带上`/`)
```powershell
./ComicCrawler.exe -url https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/
```
3. 只获取gallery信息，不下载图片
```powershell
./ComicCrawler.exe -url https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/ -info true
```
4. 下载gallery列表中的所有gallery（不能与-url一起使用）
```powershell
./ComicCrawler.exe -list gallery_list.txt
```
## 编译release版本命令
```powershell
go build -ldflags "-X 'main.buildTime=$(git show -s --format=%cd)' -X 'main.goVersion=$(go version)' -X 'ComicCrawler/client.DebugMode=0'" -o ComicCrawler.exe main.go
```
## 注意事项
- 请确保您的网络连接正常，并且能够访问支持的站点。
- 请遵守站点的相关规定和版权要求。
- 请使用合法、合规的方式进行爬取，遵守网站的爬虫规范和使用协议。
- 请尊重网站的服务器负载和带宽限制，避免对其造成过大的压力。
- 请避免频繁的请求和大量的并发连接，以免对网站的正常运行造成干扰。