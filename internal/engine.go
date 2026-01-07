package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Engine 结构体
type Engine struct {
	Cfg               *Config
	ExplicitFiles     map[string]bool
	ExpectationsByDir map[string]map[string]bool
	MergedFiles       map[string]bool
	VisitedDirs       map[string]bool
	TraversePaths     map[string]bool
}

func NewEngine(cfg *Config) *Engine {
	e := &Engine{
		Cfg:               cfg,
		ExplicitFiles:     make(map[string]bool),
		ExpectationsByDir: make(map[string]map[string]bool),
		MergedFiles:       make(map[string]bool),
		VisitedDirs:       make(map[string]bool),
		TraversePaths:     make(map[string]bool),
	}
	e.initRules()
	return e
}

func (e *Engine) initRules() {
	for _, r := range e.Cfg.Rules {
		if r.IsInclude {
			// 1. 计算 Vital Path (BaseDir 的父级必须遍历)
			if r.BaseDir != "" {
				parts := strings.Split(r.BaseDir, "/")
				// 注意：如果 BaseDir 是 a/b，我们需要保证 a 被遍历
				for i := 1; i <= len(parts); i++ {
					e.TraversePaths[strings.Join(parts[:i], "/")] = true
				}
			} else {
				e.TraversePaths[""] = true // 根目录
			}

			// 2. 生成期待列表 (仅针对精确文件名)
			// 如果 Pattern 不含通配符 (*, ?, [)，则视为显式文件期待
			if !r.CheckExts && !hasMeta(r.Pattern) {
				// 这是一个精确文件规则，如 +path/file.txt
				full := filepath.Join(r.BaseDir, r.Pattern)
				full = filepath.ToSlash(full)
				
				e.ExplicitFiles[full] = true
				
				dir := r.BaseDir
				if e.ExpectationsByDir[dir] == nil {
					e.ExpectationsByDir[dir] = make(map[string]bool)
				}
				e.ExpectationsByDir[dir][r.Pattern] = true
			}
		}
	}
}

// 简单的检查是否包含 Glob 通配符
func hasMeta(path string) bool {
	return strings.ContainsAny(path, "*?[]")
}

func (e *Engine) Run(startDir string, outFile *os.File) {
	fmt.Println("--- 智能遍历目录 (Glob Engine) ---")
	e.processDir(startDir, outFile)
	e.reportInvalidDirs()
}

func (e *Engine) processDir(dirPath string, outFile *os.File) {
	entries, err := os.ReadDir(dirPath)
	if err != nil { return }

	relDir := filepath.ToSlash(dirPath)
	if relDir == "." { relDir = "" }
	e.VisitedDirs[relDir] = true

	// [删除] 原来的这一行，因为现在不能在循环外只生成一次了
	// sepLine, cStart, cEnd := GetCommentedSeparator(e.Cfg.CommentStyle)

	// 文件处理
	for _, entry := range entries {
		if entry.IsDir() { continue }

		fileName := entry.Name()
		fullPath := filepath.Join(dirPath, fileName)
		relPath := filepath.ToSlash(fullPath)

		// 核心判定
		if e.shouldMerge(relPath, relDir, fileName) {
			// [修改] 强制转为正斜杠输出
			displayPath := filepath.ToSlash(fullPath)
			fmt.Printf("%s合并: %s%s\n", ColorGreen, displayPath, ColorReset)

			// [新增] 针对当前文件生成特定的分隔线
			// 传入 fileName 进行探测，传入 e.Cfg.CommentStyle 作为保底
			sepLine, cStart, cEnd := GetCommentedSeparator(fileName, e.Cfg.CommentStyle)

			outFile.WriteString(fmt.Sprintf("\n%s\n", sepLine))
			// [修改] 文件内部的标记也建议用正斜杠，保持一致
			outFile.WriteString(fmt.Sprintf("%s FILE: %s %s\n", cStart, displayPath, cEnd))
			outFile.WriteString(fmt.Sprintf("%s\n", sepLine))

			content, err := os.ReadFile(fullPath)
			if err == nil {
				outFile.Write(content)
				// 确保文件末尾换行，防止下一个 header 粘连
				outFile.WriteString("\n")
			} else {
				outFile.WriteString(fmt.Sprintf("// Error reading file: %v\n", err))
			}
			e.MergedFiles[relPath] = true
		}
	}

	// 缺失检查
	if expectedFiles, ok := e.ExpectationsByDir[relDir]; ok {
		for fname := range expectedFiles {
			var expectedRel string
			if relDir == "" { expectedRel = fname } else { expectedRel = relDir + "/" + fname }
			if !e.MergedFiles[expectedRel] {
				// [修改] 强制转为正斜杠输出
				fmt.Printf("%s缺失: %s%s\n", ColorRed, filepath.ToSlash(expectedRel), ColorReset)
			}
		}
	}

	// 递归
	for _, entry := range entries {
		if !entry.IsDir() { continue }
		subDir := filepath.Join(dirPath, entry.Name())
		relSub := filepath.ToSlash(subDir)
		if e.shouldTraverse(relSub) {
			e.processDir(subDir, outFile)
		}
	}
}

func (e *Engine) shouldMerge(relPath string, relDir string, fileName string) bool {
	bestLen := -1
	bestAction := 0 

	for _, r := range e.Cfg.Rules {
		matched := false
		
		// 1. 检查目录是否匹配
		dirMatch := false
		if r.Recursive {
			// 递归模式: 必须是 BaseDir 或者是 BaseDir 的子目录
			if r.BaseDir == "" || relDir == r.BaseDir || strings.HasPrefix(relDir, r.BaseDir+"/") {
				dirMatch = true
			}
		} else {
			// 平铺模式: 必须在 BaseDir 当前层级
			if relDir == r.BaseDir {
				dirMatch = true
			}
		}

		// 2. 检查文件/后缀是否匹配
		if dirMatch {
			if r.CheckExts {
				// 目录规则 (+path/): 检查全局 extensions
				if e.checkGlobalExt(fileName) {
					matched = true
				}
			} else {
				// 模式规则 (+path/*.c): 使用 filepath.Match 进行标准 Glob 匹配
				// r.Pattern 可能是 *, *.c, name.*, test_?.c 等
				if ok, _ := filepath.Match(r.Pattern, fileName); ok {
					matched = true
				}
			}
		}

		if matched {
			if len(r.Raw) > bestLen {
				bestLen = len(r.Raw)
				if r.IsInclude { bestAction = 1 } else { bestAction = 2 }
			}
		}
	}
	return bestAction == 1
}

func (e *Engine) checkGlobalExt(fileName string) bool {
	ext := filepath.Ext(fileName)
	for _, valid := range e.Cfg.Extensions {
		if strings.EqualFold(ext, valid) { return true }
	}
	return false
}

func (e *Engine) shouldTraverse(relDir string) bool {
	if e.TraversePaths[relDir] { return true }
	
	// 如果有递归规则覆盖了此目录，也必须遍历
	for _, r := range e.Cfg.Rules {
		if r.IsInclude && r.Recursive {
			// 检查 relDir 是否在 r.BaseDir 之下 (或者就是 r.BaseDir)
			// 注意：这里逻辑要反过来，如果我们还没走到 BaseDir，当然要往深了走
			// 如果我们已经在 BaseDir 下面了，并且是 Recursive，那更要走
			
			// 情况 A: relDir 是 BaseDir 的上级 (正在前往 BaseDir 的路上) -> TraversePaths 已经处理了
			// 情况 B: relDir 是 BaseDir 的下级 -> 允许进入
			if r.BaseDir == "" || strings.HasPrefix(relDir, r.BaseDir+"/") {
				return true
			}
		}
	}
	return false
}

func (e *Engine) reportInvalidDirs() {
	// 保持不变
	hasInvalid := false
	for dir := range e.ExpectationsByDir {
		if dir == "" { continue }
		if !e.VisitedDirs[dir] {
			if !hasInvalid {
				fmt.Println("\n--- 检查无效路径 ---")
				hasInvalid = true
			}
			// [修改] 强制转为正斜杠输出
			fmt.Printf("%s目录未找到: %s (其下指定文件均缺失)%s\n", ColorRed, filepath.ToSlash(dir), ColorReset)
		}
	}
}
