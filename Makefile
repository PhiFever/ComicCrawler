# 判断操作系统
ifeq ($(OS),Windows_NT)
    EXE_SUFFIX := .exe
else
    EXE_SUFFIX :=
endif

build:
	go build -ldflags "-X 'ComicCrawler/client.DebugMode=0'" -o ComicCrawler$(EXE_SUFFIX) main.go