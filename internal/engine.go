package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Engine 合并处理引擎
type Engine struct {
	Cfg         *Config
	IgnorePaths map[string]bool // 内置跳过路径（如自身输出文件、.mergerule）
	MergedCount int             // 合并文件总数统计
}

// NewEngine 创建并初始化引擎
func NewEngine(cfg *Config, ignoreFiles ...string) *Engine {
	ignoreMap := make(map[string]bool)
	for _, f := range ignoreFiles {
		ignoreMap[filepath.ToSlash(f)] = true
	}
	return &Engine{
		Cfg:         cfg,
		IgnorePaths: ignoreMap,
	}
}

// Run 启动引擎并遍历
func (e *Engine) Run(startDir string, outFile *os.File) {
	fmt.Println("--- 智能合并目录 (Gitignore Engine) ---")
	e.processDir(startDir, "", outFile)
}

func (e *Engine) processDir(dirPath string, relDir string, outFile *os.File) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	// 1. 处理当前目录下的所有文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		fullPath := filepath.Join(dirPath, fileName)
		slashPath := filepath.ToSlash(fullPath)

		// 检查内置保护文件（如 .mergerule、.gitignore、输出文件）
		if fileName == ".mergerule" || fileName == ".gitignore" || e.IgnorePaths[slashPath] || e.IgnorePaths[fileName] {
			continue
		}

		// 构建 Git 相对路径切片
		parts := makePathParts(relDir, fileName)

		// Gitignore 语义：Match 返回 true 代表被排除；返回 false 代表包含
		if e.Cfg.Matcher != nil && e.Cfg.Matcher.Match(parts, false) {
			continue
		}

		// 命中包含：获取大小并实时汇报
		info, _ := entry.Info()
		sizeStr := "???"
		if info != nil {
			sizeStr = FormatByteSize(info.Size())
		}
		fmt.Printf("%s合并: [%s] %s%s\n", ColorGreen, sizeStr, slashPath, ColorReset)

		// 自动探测后缀注释符并写入文件头
		sepLine, cStart, cEnd := GetCommentedSeparator(fileName)
		outFile.WriteString(fmt.Sprintf("\n%s\n", sepLine))
		outFile.WriteString(fmt.Sprintf("%s FILE: %s %s\n", cStart, slashPath, cEnd))
		outFile.WriteString(fmt.Sprintf("%s\n", sepLine))

		// 写入文件正文并换行防粘连
		content, err := os.ReadFile(fullPath)
		if err == nil {
			outFile.Write(content)
			outFile.WriteString("\n")
		} else {
			outFile.WriteString(fmt.Sprintf("// 读取文件失败: %v\n", err))
		}
		e.MergedCount++
	}

	// 2. 递归遍历子目录
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		// 内置忽略版本控制目录
		if dirName == ".git" {
			continue
		}

		subRelDir := dirName
		if relDir != "" {
			subRelDir = relDir + "/" + dirName
		}
		dirParts := strings.Split(subRelDir, "/")

		// 目录级别剪枝判定：如果目录本身被排除，直接截断不再深入扫描
		if e.Cfg.Matcher != nil && e.Cfg.Matcher.Match(dirParts, true) {
			continue
		}

		subDir := filepath.Join(dirPath, dirName)
		e.processDir(subDir, subRelDir, outFile)
	}
}

// makePathParts 将目录与文件名组合为 gitignore 所需的路径段切片
func makePathParts(relDir string, fileName string) []string {
	if relDir == "" {
		return []string{fileName}
	}
	parts := strings.Split(relDir, "/")
	return append(parts, fileName)
}
