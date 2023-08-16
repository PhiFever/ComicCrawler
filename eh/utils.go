package eh

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ErrorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ToSafeFilename(in string) string {
	//https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
	//全部替换为_
	rp := strings.NewReplacer(
		"/", "_",
		`\`, "_",
		"<", "_",
		">", "_",
		":", "_",
		`"`, "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	rt := rp.Replace(in)
	return rt
}

// SaveFile 用于保存文件
func SaveFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	//fmt.Println(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// BuildCache 用于生成utf-8格式的图片列表缓存文件
func BuildCache(saveDir string, cacheFile string, imageDataList []map[string]string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	ErrorCheck(err)

	// 打开文件用于写入数据
	file, err := os.Create(filepath.Join(dir, cacheFile))
	if err != nil {
		fmt.Println("File creation error:", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	// 创建 JSON 编码器，并指定输出流为文件
	encoder := json.NewEncoder(file)
	// 设置编码器的输出流为 UTF-8
	encoder.SetIndent("", "    ") // 设置缩进，可选
	encoder.SetEscapeHTML(false)  // 禁用转义 HTML
	err = encoder.Encode(imageDataList)
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		return err
	}

	return nil
}

func LoadCache(filePath string) ([]map[string]string, error) {
	var imageDataList []map[string]string
	// 打开utf-8文件用于读取数据
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File open error:", err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)
	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)
	// 设置解码器的输入流为 UTF-8
	err = decoder.Decode(&imageDataList)
	if err != nil {
		return nil, err
	}
	return imageDataList, nil
}

func CacheFileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}
