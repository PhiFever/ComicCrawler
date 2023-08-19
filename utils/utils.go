package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// BuildCache 用于生成utf-8格式的缓存文件
// data为待写入数据结构
func BuildCache(saveDir string, cacheFile string, data interface{}) error {
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
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		return err
	}

	return nil
}

// LoadCache 用于加载utf-8格式的缓存文件
// result是一个指向目标数据结构的指针
func LoadCache(filePath string, result interface{}) error {
	// 打开utf-8格式的文件用于读取数据
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File open error:", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)
	// 设置解码器的输入流为 UTF-8
	err = decoder.Decode(result)
	if err != nil {
		return err
	}
	return nil
}

func CacheFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

func TrueRandFloat(min, max float64) float64 {
	// 使用当前时间的纳秒部分作为种子值
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	randomGenerator := rand.New(source)

	// 生成范围在 [min, max) 内的随机浮点数
	randomFloat := min + randomGenerator.Float64()*(max-min)
	return randomFloat
}

func GetFileTotal(dirPath string, fileSuffix string) int {
	var total int
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, fileSuffix) {
			total++
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return total
}
