## 支持站点
目前支持`e-hentai.org`,`manhua.dmzj.com`(需要放在同级目录下的cookies.json，可以通过EditThisCookie导出，注意调整翻页模式，使其中的fanyemodeval=2)

## TODO
1. dmzj目录带分页的情况无法处理
2. dmzj目录的`其他系列`尚未处理
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
./comic_crawler.exe -h
```
2. 获取gallery信息，并下载图片
```powershell
./comic_crawler.exe -url https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/
```
3. 只获取gallery信息，不下载图片
```powershell
./comic_crawler.exe -url https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/ -info true
```
4. 下载gallery列表中的所有gallery（不能与-url一起使用）
```powershell
./comic_crawler.exe -list gallery_list.txt
```
## 编译release版本命令

```powershell
go build -ldflags="-s -w" -ldflags "-X 'main.buildTime=$(git show -s --format=%cd)' -X 'main.goVersion=$(go version)'" -o comic_crawler.exe main.go
```
```powershell
go build -ldflags "-X 'main.buildTime=$(git show -s --format=%cd)' -X 'main.goVersion=$(go version)'" -o comic_crawler.exe main.go
```