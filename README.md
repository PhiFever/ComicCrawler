## Todo
1. [ ] Add a `--version` option
2. [ ] 增加上次下载的进度，使下次下载同一gallery时从上次失败的进度开始下载
3. [fixed ]实现每处理一个主页就下载一次图片，而不是等到所有主页处理完毕后再下载图片
4. [fixed ]main.saveImages的实现不够优雅，需要重构
```
process.json
{
    "gallery_url": 123456,
    "total_page": 10,
    "total_image": 100,
    "downloaded_image": 50
}
```
## 编译release版本命令
```powershell
go build -ldflags="-s -w" -o eh_downloader.exe main.go
```