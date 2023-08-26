package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	DelayMs = 330
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

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// SortMapsByIntKey 用于按照map中的int键值对进行排序
func SortMapsByIntKey(maps []map[int]string, ascending bool) []map[int]string {
	getKey := func(m map[int]string) int {
		for key := range m {
			return key
		}
		return 0
	}
	sort.Slice(maps, func(i, j int) bool {
		keyI := getKey(maps[i])
		keyJ := getKey(maps[j])
		if ascending {
			return keyI < keyJ
		}
		return keyI > keyJ
	})
	return maps
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

// BuildCache 用于生成utf-8格式的缓存文件 data为待写入数据结构
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

// LoadCache 用于加载utf-8格式的缓存文件 result是一个指向目标数据结构的指针
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

func FileExists(filePath string) bool {
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

// GetFileTotal 用于获取指定目录下指定后缀的文件数量
func GetFileTotal(dirPath string, fileSuffixes []string) int {
	var count int // 用于存储文件数量的变量

	// 使用Walk函数遍历指定目录及其子目录中的所有文件和文件夹
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查是否为文件
		if !info.IsDir() {
			// 获取文件的扩展名
			ext := filepath.Ext(path)
			// 将扩展名转换为小写，以便比较
			ext = strings.ToLower(ext)
			// 检查文件扩展名是否在指定的后缀列表中
			for _, suffix := range fileSuffixes {
				if ext == suffix {
					count++
					break // 找到匹配的后缀，停止循环
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("遍历目录出错:", err)
	}

	return count
}

// GetBeginIndex 用于获取指定目录下指定格式和后缀的文件中最大的序号，用于计算剩余图片数（目前只支持`序号1_序号2.后缀`的格式）
func GetBeginIndex(dirPath string, fileSuffixes []string) int {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return -1
	}

	maxIndex := -1

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		for _, suffix := range fileSuffixes {
			if strings.HasSuffix(file.Name(), suffix) {
				name := strings.TrimSuffix(file.Name(), suffix)
				parts := strings.Split(name, "_")
				if len(parts) != 2 {
					continue
				}

				index, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}

				if index > maxIndex {
					maxIndex = index
				}
			}
		}
	}

	return maxIndex
}

// ReadListFile 用于按行读取列表文件，返回一个字符串切片
func ReadListFile(filePath string) ([]string, error) {
	var list []string
	file, err := os.Open(filePath)
	if err != nil {
		return list, err
	}
	defer func(file *os.File) {
		err := file.Close()
		ErrorCheck(err)
	}(file)

	var line string
	for {
		_, err := fmt.Fscanln(file, &line)
		if err != nil {
			break
		}
		list = append(list, line)
	}
	return list, nil
}

// SaveImages 保存imageDataList中的所有图片，imageInfoMap中的每个元素都是一个map，包含两个键值对，imageName和imageUrl
func SaveImages(baseCollector *colly.Collector, imageInfoMap []map[string]string, saveDir string) error {
	dir, err := filepath.Abs(saveDir)
	err = os.MkdirAll(dir, os.ModePerm)
	ErrorCheck(err)

	var imageContent []byte

	baseCollector.OnResponse(func(r *colly.Response) {
		imageContent = r.Body
	})

	for _, data := range imageInfoMap {
		imageName := data["imageName"]
		imageUrl := data["imageUrl"]
		filePath, err := filepath.Abs(filepath.Join(dir, imageName))
		ErrorCheck(err)
		err = baseCollector.Request("GET", imageUrl, nil, nil, nil)
		ErrorCheck(err)
		//增加延时，防止被ban
		time.Sleep(time.Millisecond * time.Duration(DelayMs))
		err = SaveFile(filePath, imageContent)
		if err != nil {
			fmt.Println("Error saving image:", err)
		} else {
			fmt.Println("Image saved:", filePath)
		}
	}

	return nil
}

// ExtractSubstringFromText 按照Pattern在text里匹配，找到了就返回匹配到的部分
func ExtractSubstringFromText(pattern string, text string) (string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	match := regex.FindStringSubmatch(text)
	if match != nil {
		number := match[1]
		return number, nil
	} else {
		return "", fmt.Errorf("在pattern中未找到匹配的数字")
	}
}
