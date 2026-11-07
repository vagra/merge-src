package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"merge-src/internal" // 引用根目录下的 internal 包
)

// getProjectDirName 安全提取当前项目目录名称作为输出前缀
func getProjectDirName() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "merged_src"
	}
	name := filepath.Base(pwd)
	if name == "/" || name == "." || name == "\\" || name == "" {
		return "merged_src"
	}
	return name
}

func main() {
	// 1. 解析配置
	configPath := ".mergerule"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("%s错误: 未找到 .mergerule%s\n", internal.ColorRed, internal.ColorReset)
		os.Exit(1)
	}

	cfg, err := internal.ParseConfig(configPath)
	if err != nil {
		fmt.Printf("配置解析错误: %v\n", err)
		os.Exit(1)
	}

	// 2. 创建输出 (采用当前项目目录名作为前缀)
	timestamp := time.Now().Format("20060102-1504")
	outputPrefix := getProjectDirName()
	outputName := fmt.Sprintf("%s-%s.txt", outputPrefix, timestamp)
	outFile, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// 3. 运行引擎 (注入 outputName 以免引擎自身合并正在生成的输出文件)
	eng := internal.NewEngine(cfg, outputName)
	eng.Run(".", outFile)

	// 4. 获取结果文件大小
	// 确保内容已写入磁盘 (虽然 Write 主要是内存操作，但为了 Stat 准确性)
	outFile.Sync()
	stat, _ := outFile.Stat()
	sizeStr := "???"
	if stat != nil {
		sizeStr = internal.FormatByteSize(stat.Size())
	}

	// 输出格式： ✔ 完成: 共合并 X 个文件 [  XX KB] filename.txt
	fmt.Printf("\n%s✔ 完成: 共合并 %d 个文件 [%s] %s%s\n", internal.ColorGreen, eng.MergedCount, sizeStr, outputName, internal.ColorReset)
}
