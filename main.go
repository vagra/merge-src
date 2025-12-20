package main

import (
	"fmt"
	"os"
	"time"

	"merge-src/internal" // 引用根目录下的 internal 包
)

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

	// 2. 创建输出
	timestamp := time.Now().Format("20060102-1504")
	outputName := fmt.Sprintf("%s-%s.txt", cfg.OutputPrefix, timestamp)
	outFile, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// 3. 运行引擎
	eng := internal.NewEngine(cfg)
	eng.Run(".", outFile)

	fmt.Printf("\n%s✔ 完成: %s%s\n", internal.ColorGreen, outputName, internal.ColorReset)
}
