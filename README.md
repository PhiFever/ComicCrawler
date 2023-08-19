## Todo

1. [ ] Add a `--version` option
2. [fixed ] 增加缓存上次下载的进度，使下次下载同一gallery时从上次失败的进度开始下载
3. [ ] 增加批量下载功能，从文件中读取gallery_url，下载所有的gallery
4. [fixed ]实现每处理一个主页就下载一次图片，而不是等到所有主页处理完毕后再下载图片
5. [fixed ]main.saveImages的实现不够优雅，需要重构
6. [ ]重构eh的相关函数，使主函数能根据传入参数的不同而调用不同网站的支持接口

```
缓存文件格式
process.json
{
    "gallery_url": "https://e-hentai.org/g/xxxxxx/xxxxxxxxxx/",
    "gallery_title": "xxxxxx",
    "total_image": 100,
    "tag_list": {
        "artist": "xyz",
        "male": [
            "aaa",
            "bbb",
        ],
        "female":[
            "xxx",
            "yyy",
        ],
        ........
    },
}
```

## 编译release版本命令

```powershell
go build -ldflags="-s -w" -o eh_downloader.exe main.go
```