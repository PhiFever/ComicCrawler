## Todo

1. [fixed ] Add a `--version` option
2. [fixed ] 增加缓存上次下载的进度，使下次下载同一gallery时从上次失败的进度开始下载
3. [fixed ] 增加批量下载功能，从文件中按行读取gallery_url，下载所有的gallery
4. [fixed ]实现每处理一个主页就下载一次图片，而不是等到所有主页处理完毕后再下载图片
5. [fixed ]main.saveImages的实现不够优雅，需要重构
6. [  ]重构相关函数，使主函数能根据传入参数的不同而调用不同网站的支持接口
7. [  ]增加对exhentai的支持
8. [fixed ]增加控制台彩色输出

缓存文件格式
`galleryInfo.json`
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

##  使用说明
1. 获取详细说明（对应的短参数名）
```powershell
./comic_downloader.exe -h
```
2. 获取gallery信息，并下载图片
```powershell
./comic_downloader.exe -url https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/
```
3. 只获取gallery信息，不下载图片
```powershell
./comic_downloader.exe -url https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/ -info true
```
4. 下载gallery列表中的所有gallery（不能与-url一起使用）
```powershell
./comic_downloader.exe -list gallery_list.txt
```
## 编译release版本命令

```powershell
go build -ldflags="-s -w" -ldflags "-X 'main.buildTime=$(git show -s --format=%cd)' -X 'main.goVersion=$(go version)'" -o comic_downloader.exe main.go
```
```powershell
go build -ldflags "-X 'main.buildTime=$(git show -s --format=%cd)' -X 'main.goVersion=$(go version)'" -o comic_downloader.exe main.go
```