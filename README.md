## Todo
1.[ ] Add a `--version` option
## 编译release版本命令
```powershell
go build -ldflags="-s -w" -o eh_downloader.exe main.go
```